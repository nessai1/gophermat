package user

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
)

var ErrUserNotFound = errors.New("user not found")
var ErrIncorrectUserPassword = errors.New("user password is wrong")
var ErrLoginAlreadyExists = errors.New("input user login already exists")

type User struct {
	Login   string
	Balance float32

	password string
}

type Repository interface {
	GetUserByLogin(context.Context, string) (*User, error)
	CreateUser(context.Context, *User) error
}

type Controller struct {
	repository Repository
}

func NewController(repository Repository) Controller {
	return Controller{repository: repository}
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

func buildPasswordHash(password string) string {
	shaSum := sha256.Sum256([]byte(password))

	return fmt.Sprintf("%x", shaSum)
}
