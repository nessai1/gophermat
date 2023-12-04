package order

import (
	"context"
	"errors"
	"fmt"
	"github.com/nessai1/gophermat/internal/intransaction"
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
	transaction    intransaction.Transaction
}

func NewWithdrawController(repository WithdrawRepository, userController *user.Controller, transaction intransaction.Transaction) *WithdrawController {
	return &WithdrawController{
		repository:     repository,
		userController: userController,
		transaction:    transaction,
	}
}

func (controller *WithdrawController) CreateWithdrawByUser(ctx context.Context, innerUser *user.User, orderID string, sum int64) (*Withdraw, error) {

	if !IsOrderNumberCorrect(orderID) {
		return nil, ErrInvalidOrderNumber
	}

	if sum == 0 {
		return nil, ErrEmptyBalance
	}

	if sum > innerUser.Balance {
		return nil, ErrNoMoney
	}

	var withdraw *Withdraw
	err := controller.transaction.InTransaction(ctx, func(innerCtx context.Context) error {
		balance := innerUser.Balance - sum
		txErr := controller.userController.SetUserBalanceByID(ctx, innerUser.ID, balance)
		if txErr != nil {
			return fmt.Errorf("error while set user balance: %w", txErr)
		}

		withdraw, txErr = controller.repository.AddWithdraw(ctx, innerUser.ID, orderID, sum)

		if txErr != nil {
			return fmt.Errorf("error while add withdraw: %w", txErr)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error while create withdraw by user: %w", err)
	}

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
