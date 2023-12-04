package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nessai1/gophermat/internal/user"
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
	EnrollmentCh     chan<- Enrollment
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
	GetProcessedEnrollments(ctx context.Context) ([]*Enrollment, error)
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
		go controller.loadOrder(enrollment)
		if err != nil {
			return nil, fmt.Errorf("error while start order loading operation: %w", err)
		}
	}

	return enrollment, nil
}

func (controller *EnrollmentController) ChangeStatusByOrderID(ctx context.Context, orderID, status string) error {
	return controller.repository.ChangeStatus(ctx, orderID, status)
}

func (controller *EnrollmentController) GetProcessedEnrollments(ctx context.Context) ([]*Enrollment, error) {
	enrollmentList, err := controller.repository.GetProcessedEnrollments(ctx)

	return enrollmentList, err
}

func (controller *EnrollmentController) loadOrder(enrollment *Enrollment) {
	controller.EnrollmentCh <- *enrollment
}

func (controller *EnrollmentController) GetUserOrderListByID(ctx context.Context, userID int) ([]*Enrollment, error) {
	enrollmentList, err := controller.repository.GetListByUserID(ctx, userID)

	return enrollmentList, err
}

func (controller *EnrollmentController) UpdateOrderAccrualByID(ctx context.Context, orderID string, accrual int) error {
	return controller.repository.UpdateOrderAccrual(ctx, orderID, accrual)
}
