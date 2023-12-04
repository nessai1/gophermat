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

func (repository *MapRepository) SetUserBalanceByID(_ context.Context, userID int, balance int64) error {
	for i, user := range repository.data {
		if user.ID == userID {
			user.Balance = balance
			repository.data[i] = user
		}
	}

	return ErrUserNotFound
}

func (repository *MapRepository) GetUserByID(ctx context.Context, userID int) (*User, error) {
	for _, user := range repository.data {
		if user.ID == userID {
			return &user, nil
		}
	}

	return nil, ErrUserNotFound
}
