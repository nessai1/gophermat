package user

import (
	"crypto/sha256"
	"errors"
	"fmt"
)

var ErrUserNotFound = errors.New("user not found")
var ErrIncorrectUserPassword = errors.New("user password is wrong")

type User struct {
	Login   string
	Balance float32

	password string
}

type Repository interface {
	GetUserByLogin(string) (*User, error)
}

type Controller struct {
	repository Repository
}

func NewController(repository Repository) Controller {
	return Controller{repository: repository}
}

func (controller *Controller) GetUserByCredentials(login, password string) (*User, error) {
	user, err := controller.repository.GetUserByLogin(login)

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

func buildPasswordHash(password string) string {
	shaSum := sha256.Sum256([]byte(password))

	return fmt.Sprintf("%x", shaSum)
}
