package order

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/nessai1/gophermat/internal/intransaction"
	"github.com/nessai1/gophermat/internal/user"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type EnrollmentWorker struct {
	ch                   chan Enrollment
	client               http.Client
	enrollmentController *EnrollmentController
	userController       *user.Controller
	logger               *zap.Logger
	serviceAddr          string
	transaction          intransaction.Transaction
}

func (worker *EnrollmentWorker) Listen() {
	for enrollment := range worker.ch {
		worker.requireOrder(&enrollment)
	}
}

func (worker *EnrollmentWorker) requireOrder(enrollment *Enrollment) {
	worker.logger.Debug("start required order", zap.String("order id", enrollment.OrderID), zap.String("order init status", enrollment.Status))
	err := worker.enrollmentController.ChangeStatusByOrderID(context.TODO(), enrollment.OrderID, EnrollmentStatusProcessing)
	if err != nil {
		worker.logger.Error("error while try to change order status on processing in worker", zap.Error(err), zap.String("order id", enrollment.OrderID))
		go func(e Enrollment) {
			time.Sleep(time.Second * 5) // задержка что-бы не заспамить в случае какой-то поломки СУБД
			worker.ch <- e              // кладем в конец очереди на следующую попытку
		}(*enrollment)
	}

	for {
		ctx := context.TODO()
		resp, err := worker.client.Get(worker.serviceAddr + "/api/orders/" + enrollment.OrderID)
		if err != nil {
			break
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := resp.Header.Get("Retry-After")
			retryAfterInt, _ := strconv.Atoi(retryAfter)
			worker.logger.Info("got many request to accrual service", zap.Int("retry after", retryAfterInt))
			resp.Body.Close()
			time.Sleep(time.Second * time.Duration(retryAfterInt))
			continue
		}

		if resp.StatusCode != http.StatusOK {
			worker.logger.Error("got unsuccessful status from accrual service", zap.String("status", resp.Status))
			resp.Body.Close()
			time.Sleep(time.Second * 5)
			continue
		}

		var buffer bytes.Buffer
		_, err = buffer.ReadFrom(resp.Body)
		resp.Body.Close()
		if err != nil {
			worker.logger.Error("error while read response from accrual service", zap.String("order id", enrollment.OrderID))
			time.Sleep(time.Second * 5)
			continue
		}

		var accrualInfo OrderAccrualInfo
		err = json.Unmarshal(buffer.Bytes(), &accrualInfo)
		if err != nil {
			worker.logger.Error("error while unmarshal response from accrual service", zap.Error(err))
		}

		if accrualInfo.Status == orderAccrualStatusInvalid {
			err = worker.enrollmentController.ChangeStatusByOrderID(ctx, enrollment.OrderID, EnrollmentStatusInvalid)
			if err != nil {
				worker.logger.Error("error while change status in accrual worker", zap.Error(err), zap.String("status", EnrollmentStatusInvalid), zap.String("order id", enrollment.OrderID))
				time.Sleep(time.Second * 5)
				continue
			}

			worker.logger.Info("accrual complete", zap.String("status", EnrollmentStatusInvalid), zap.String("order id", enrollment.OrderID))
			return
		}

		if accrualInfo.Status == orderAccrualStatusProcessing || accrualInfo.Status == orderAccrualStatusRegistered {
			worker.logger.Info("order is still awaiting accrual", zap.String("current status", accrualInfo.Status), zap.String("order id", enrollment.OrderID))
			time.Sleep(time.Second * 5)
			continue
		}

		err = worker.transaction.InTransaction(ctx, func(innerCtx context.Context) error {

			txErr := worker.enrollmentController.ChangeStatusByOrderID(innerCtx, accrualInfo.Order, EnrollmentStatusProcessed)
			if txErr != nil {
				return fmt.Errorf("error while change order status: %w", txErr)
			}

			df, txErr := user.ParseBalance(string(accrualInfo.Accrual))
			if txErr != nil {
				return fmt.Errorf("error while parse balance: %w", txErr)
			}

			txErr = worker.enrollmentController.UpdateOrderAccrualByID(innerCtx, enrollment.OrderID, int(df))
			if txErr != nil {
				return fmt.Errorf("error while get user from controller: %w", txErr)
			}

			owner, txErr := worker.userController.GetUserByID(innerCtx, enrollment.UserID)
			if err != nil {
				return fmt.Errorf("error while get user: %w", txErr)
			}

			balance := owner.Balance + df
			txErr = worker.userController.SetUserBalanceByID(context.TODO(), owner.ID, balance)
			if txErr != nil {
				return fmt.Errorf("error while update user balance: %w", txErr)
			}

			worker.logger.Error("successful accruled order", zap.String("order id", enrollment.OrderID), zap.Int("sum", int(df)))

			return nil
		})

		if err != nil {
			worker.logger.Error("error while make user-accrual update transaction", zap.Error(err))
			time.Sleep(time.Second * 5)
			continue
		}

		return
	}
}

func StartEnrollmentWorker(userController *user.Controller, enrollmentController *EnrollmentController, logger *zap.Logger, serviceAddr string, transaction intransaction.Transaction) (chan<- Enrollment, error) {
	ch := make(chan Enrollment, 10)

	enrollmentList, err := enrollmentController.GetProcessedEnrollments(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("error while get new enrollments: %w", err)
	}

	if len(enrollmentList) > 0 {
		logger.Info("service has unresolved orders, start preload goroutine", zap.Int("enrollments len", len(enrollmentList)))
		go preloadOrders(enrollmentList, ch)
	}

	go func(ch chan Enrollment) {
		client := http.Client{
			Timeout: time.Second * 10,
		}

		worker := EnrollmentWorker{

			client:               client,
			ch:                   ch,
			enrollmentController: enrollmentController,
			userController:       userController,
			serviceAddr:          serviceAddr,
			logger:               logger,
			transaction:          transaction,
		}

		worker.Listen()
	}(ch)

	return ch, nil
}

func preloadOrders(enrollments []*Enrollment, ch chan<- Enrollment) {
	for _, enrollment := range enrollments {
		ch <- *enrollment
	}
}
