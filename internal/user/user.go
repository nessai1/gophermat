package user

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrUserNotFound = errors.New("user not found")
var ErrIncorrectUserPassword = errors.New("user password is wrong")
var ErrLoginAlreadyExists = errors.New("input user login already exists")

type User struct {
	ID      int
	Login   string
	Balance int64

	password string
}

func ParseBalance(balance string) (int64, error) {
	parts := strings.Split(balance, ".")

	if len(parts) > 2 {
		return 0, fmt.Errorf("parts of balance must be less than 3, got %d", len(parts))
	}

	bigPart, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	val := int64(bigPart * 100)

	if len(parts) == 2 {
		smallPart, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, err
		}

		if smallPart >= 100 {
			return 0, fmt.Errorf("small part must be less than 100")
		}

		if smallPart < 10 {
			val += int64(smallPart * 10)
		} else {
			val += int64(smallPart)
		}
	}

	return val, nil
}

type Repository interface {
	GetUserByLogin(context.Context, string) (*User, error)
	CreateUser(context.Context, *User) error
	GetUserByID(context.Context, int) (*User, error)
	SetUserBalanceByID(context.Context, int, int64) error
}

type Controller struct {
	repository Repository
}

func NewController(repository Repository) *Controller {
	return &Controller{repository: repository}
}

func (controller *Controller) GetUserByCredentials(ctx context.Context, login, password string) (*User, error) {
	user, err := controller.repository.GetUserByLogin(ctx, login)

	if err != nil && errors.Is(err, ErrUserNotFound) {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("unhandled error while getting user by credetials: %w", err)
	}

	if user.password != buildPasswordHash(password) {
		return nil, ErrIncorrectUserPassword
	}

	return user, nil
}

func (controller *Controller) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	user, err := controller.repository.GetUserByLogin(ctx, login)
	return user, err
}

func (controller *Controller) GetUserByID(ctx context.Context, id int) (*User, error) {
	user, err := controller.repository.GetUserByID(ctx, id)
	return user, err
}

func (controller *Controller) AddUser(ctx context.Context, login, password string) (*User, error) {
	passwordHash := buildPasswordHash(password)

	user := User{
		Login:   login,
		Balance: 0,

		password: passwordHash,
	}

	err := controller.repository.CreateUser(ctx, &user)
	if err != nil && !errors.Is(err, ErrLoginAlreadyExists) {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("repository error while add user in controller: %w", err)
	}

	return &user, nil
}

func (controller *Controller) SetUserBalanceByID(ctx context.Context, userID int, balance int64) error {
	return controller.repository.SetUserBalanceByID(ctx, userID, balance)
}

//
//func (controller *Controller) AddBalanceByID(ctx context.Context, userID int, balance float32) error {
//	return controller.repository.AddBalanceByID(ctx, userID, balance)
//}

func buildPasswordHash(password string) string {
	shaSum := sha256.Sum256([]byte(password))

	return fmt.Sprintf("%x", shaSum)
}
