package user

import "context"

type MapRepository struct {
	data map[string]User
}

func (repository *MapRepository) GetUserByLogin(_ context.Context, login string) (*User, error) {
	user, isFind := repository.data[login]

	if !isFind {
		return nil, ErrUserNotFound
	}

	return &user, nil
}

func (repository *MapRepository) CreateUser(_ context.Context, user *User) error {
	_, isFound := repository.data[user.Login]
	if isFound {
		return ErrLoginAlreadyExists
	}

	repository.data[user.Login] = *user
	return nil
}
