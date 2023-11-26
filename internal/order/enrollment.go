package order

import (
	"context"
	"errors"
	"fmt"
)

var ErrEnrollmentNotFound = errors.New("enrollment not found")

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
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

type EnrollmentController struct {
	completeFetchOrderCh chan OrderAccrualInfo
	orderServiceAddr     string
	repository           EnrollmentRepository
}

type Enrollment struct {
	UserID        int
	OrderID       string
	Status        string
	Accrual       float32
	IsTransferred bool
}

type EnrollmentRepository interface {
	GetByID(ctx context.Context, orderID string) (*Enrollment, error)
	CreateNewOrder(ctx context.Context, orderID string) (*Enrollment, error)
	ChangeStatus(ctx context.Context, orderID, status string) error
}

func NewEnrollmentController(orderServiceAddr string, repository EnrollmentRepository) *EnrollmentController {
	return &EnrollmentController{orderServiceAddr: orderServiceAddr, repository: repository} // TODO: Create channel handler
}

func (controller *EnrollmentController) RequireOrder(ctx context.Context, orderNumber string, userID int) (*Enrollment, error) {
	if !IsOrderNumberCorrect(orderNumber) {
		return nil, ErrInvalidOrderNumber
	}

	enrollment, err := controller.repository.GetByID(ctx, orderNumber)
	if err != nil && errors.Is(err, ErrEnrollmentNotFound) {
		enrollment, err = controller.repository.CreateNewOrder(ctx, orderNumber)
		if err != nil {
			return nil, fmt.Errorf("error while create new enrollment order: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error while require enrollment order: %w", err)
	}

	if enrollment.UserID == userID && enrollment.Status == EnrollmentStatusNew {
		err = controller.LoadOrder(ctx, enrollment.OrderID)
		if err != nil {
			return nil, fmt.Errorf("error while start order loading operation: %w", err)
		}
	}

	return enrollment, nil
}

func (controller *EnrollmentController) LoadOrder(ctx context.Context, orderNumber string) error {
	err := controller.repository.ChangeStatus(ctx, orderNumber, EnrollmentStatusProcessing)
	if err != nil {
		return fmt.Errorf("error while update order status in require: %w", err)
	}

	go func(orderNumber string, completeFetchOrderCh chan<- OrderAccrualInfo) {

		// Тут создается клиент и делается запрос на внешний сервис - GET /api/orders/{number}
		// Если код ответа 429 - делается повторный запрос после таймера на Retry-After секунд. И так повторяется пока не будет другого ответа

		testAccrualInfo := OrderAccrualInfo{
			Order:   orderNumber,
			Status:  orderAccrualStatusRegistered,
			Accrual: 0,
		}

		completeFetchOrderCh <- testAccrualInfo
	}(orderNumber, controller.completeFetchOrderCh)

	return nil
}
