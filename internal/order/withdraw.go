package order

import (
	"context"
	"errors"
	"github.com/nessai1/gophermat/internal/user"
	"time"
)

var ErrNoMoney = errors.New("user has no money to transit")

type Withdraw struct {
	Sum         int64
	OrderID     string
	ProcessedAt time.Time
}

type WithdrawRepository interface {
	AddWithdraw(ctx context.Context, userID int, orderID string, sum int64) (*Withdraw, error)
	GetWithdrawListByUserID(ctx context.Context, userID int) ([]*Withdraw, error)
	GetWithdrawSumByUserID(ctx context.Context, userID int) (int64, error)
}

type WithdrawController struct {
	repository     WithdrawRepository
	userController *user.Controller
}

func NewWithdrawController(repository WithdrawRepository, userController *user.Controller) *WithdrawController {
	return &WithdrawController{
		repository:     repository,
		userController: userController,
	}
}

func (controller *WithdrawController) CreateWithdrawByUser(ctx context.Context, innerUser *user.User, orderID string, sum int64) (*Withdraw, error) {
	// need transaction

	if !IsOrderNumberCorrect(orderID) {
		return nil, ErrInvalidOrderNumber
	}

	if sum > innerUser.Balance {
		return nil, ErrNoMoney
	}

	balance := innerUser.Balance - sum
	err := controller.userController.SetUserBalanceByID(ctx, innerUser.ID, balance)
	if err != nil {
		return nil, err
	}

	withdraw, err := controller.repository.AddWithdraw(ctx, innerUser.ID, orderID, sum)

	if err != nil {
		return nil, err
	}

	// stop transaction

	return withdraw, nil
}

func (controller *WithdrawController) GetWithdrawByUser(ctx context.Context, innerUser *user.User) ([]*Withdraw, error) {
	list, err := controller.repository.GetWithdrawListByUserID(ctx, innerUser.ID)
	return list, err
}

func (controller *WithdrawController) GetWithdrawSumByUser(ctx context.Context, innerUser *user.User) (int64, error) {
	sum, err := controller.repository.GetWithdrawSumByUserID(ctx, innerUser.ID)
	return sum, err
}
