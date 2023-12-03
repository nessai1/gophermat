package order

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nessai1/gophermat/internal/user"
	"net/http"
	"strconv"
	"time"
)

var ErrEnrollmentNotFound = errors.New("enrollment not found")
var ErrEnrollmentAlreadyExists = errors.New("enrollment already exists")

const (
	EnrollmentStatusNew        = "NEW"        // Заказ загружен в систему, но не попал в обработку
	EnrollmentStatusProcessing = "PROCESSING" // Вознаграждение за заказ рассчитывается
	EnrollmentStatusInvalid    = "INVALID"    // Система расчёта вознаграждений отказала в расчёте
	EnrollmentStatusProcessed  = "PROCESSED"  // Данные по заказу проверены и информация о расчёте успешно получена
)

// Статусы внешнего сервиса
const (
	orderAccrualStatusRegistered = "REGISTERED"
	orderAccrualStatusInvalid    = "INVALID"
	orderAccrualStatusProcessing = "PROCESSING"
	orderAccrualStatusProcessed  = "PROCESSED"
)

type OrderAccrualInfo struct {
	Order   string      `json:"order"`
	Status  string      `json:"status"`
	Accrual json.Number `json:"accrual"`
}

type EnrollmentController struct {
	orderServiceAddr string
	repository       EnrollmentRepository
	userController   *user.Controller
	dataSource       DataSource
}

type Enrollment struct {
	UserID     int
	OrderID    string
	Status     string
	Accrual    int64
	UploadedAt time.Time
}

type EnrollmentRepository interface {
	GetByID(ctx context.Context, orderID string) (*Enrollment, error)
	CreateNewOrder(ctx context.Context, orderID string, ownerID int) (*Enrollment, error)
	ChangeStatus(ctx context.Context, orderID, status string) error
	UpdateOrderAccrual(ctx context.Context, orderID string, accrual int) error
	GetListByUserID(ctx context.Context, userID int) ([]*Enrollment, error)
}

func NewEnrollmentController(orderServiceAddr string, repository EnrollmentRepository, userController *user.Controller) *EnrollmentController {
	return &EnrollmentController{orderServiceAddr: orderServiceAddr, repository: repository, userController: userController}
}

func (controller *EnrollmentController) RequireOrder(ctx context.Context, orderNumber string, userID int) (*Enrollment, error) {
	if !IsOrderNumberCorrect(orderNumber) {
		return nil, ErrInvalidOrderNumber
	}

	enrollment, err := controller.repository.GetByID(ctx, orderNumber)
	if err != nil && errors.Is(err, ErrEnrollmentNotFound) {
		enrollment, err = controller.repository.CreateNewOrder(ctx, orderNumber, userID)
		if err != nil {
			return nil, fmt.Errorf("error while create new enrollment order: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error while require enrollment order: %w", err)
	}

	if enrollment.UserID == userID && enrollment.Status == EnrollmentStatusNew {
		err = controller.LoadOrder(ctx, userID, enrollment.OrderID)
		if err != nil {
			return nil, fmt.Errorf("error while start order loading operation: %w", err)
		}
	}

	return enrollment, nil
}

func (controller *EnrollmentController) LoadOrder(ctx context.Context, ownerID int, orderNumber string) error {
	err := controller.repository.ChangeStatus(ctx, orderNumber, EnrollmentStatusProcessing)
	if err != nil {
		return fmt.Errorf("error while update order status in require: %w", err)
	}

	go func(serviceAddr, orderNumber string, ownerID int, enrollmentRepository EnrollmentRepository, userController *user.Controller) {
		for {
			resp, err := http.Get(serviceAddr + "/api/orders/" + orderNumber)
			if err != nil {
				break
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				retryAfter := resp.Header.Get("Retry-After")
				retryAfterInt, _ := strconv.Atoi(retryAfter)
				time.Sleep(time.Second * time.Duration(retryAfterInt))
				continue
			}

			if resp.StatusCode != http.StatusOK {
				break
			}

			var buffer bytes.Buffer
			buffer.ReadFrom(resp.Body)
			var accrualInfo OrderAccrualInfo
			json.Unmarshal(buffer.Bytes(), &accrualInfo)
			if accrualInfo.Status == orderAccrualStatusInvalid {
				enrollmentRepository.ChangeStatus(context.TODO(), orderNumber, EnrollmentStatusInvalid)
				return
			}

			if accrualInfo.Status == orderAccrualStatusProcessing || accrualInfo.Status == orderAccrualStatusRegistered {
				time.Sleep(time.Second * 5)
				continue
			}

			// need transaction

			enrollmentRepository.ChangeStatus(context.TODO(), orderNumber, EnrollmentStatusProcessed)
			owner, err := userController.GetUserByID(context.TODO(), ownerID)
			if err != nil {
				return
			}

			df, err := user.ParseBalance(string(accrualInfo.Accrual))
			if err != nil {
				return
			}

			enrollmentRepository.UpdateOrderAccrual(context.TODO(), orderNumber, int(df))

			balance := owner.Balance + df
			userController.SetUserBalanceByID(context.TODO(), ownerID, balance)
			return
		}

	}(controller.orderServiceAddr, orderNumber, ownerID, controller.repository, controller.userController)

	return nil
}

func (controller *EnrollmentController) GetUserOrderListByID(ctx context.Context, userID int) ([]*Enrollment, error) {
	enrollmentList, err := controller.repository.GetListByUserID(ctx, userID)

	return enrollmentList, err
}
