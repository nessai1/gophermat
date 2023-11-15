package user

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestRepository struct {
	data map[string]User
}

func (repository *TestRepository) GetUserByLogin(login string) (*User, error) {
	user, isFind := repository.data[login]

	if !isFind {
		return nil, ErrUserNotFound
	}

	return &user, nil
}

func TestController_GetUserByCredentials(t *testing.T) {
	userOnePassword := "superSecret"
	userOneLogin := "userOne"
	userOneBalance := float32(300)
	users := map[string]User{
		userOneLogin: {
			Login:    userOneLogin,
			Balance:  userOneBalance,
			password: buildPasswordHash(userOnePassword),
		},
	}

	controller := NewController(&TestRepository{data: users})

	user, err := controller.GetUserByCredentials(userOneLogin, userOnePassword)
	if assert.NoErrorf(t, err, "user with login %s must be found", userOneLogin) {
		assert.Equalf(t, userOneLogin, user.Login, "user login not equeal (%s != %s)", userOneLogin, user.Login)
		assert.Equalf(t, userOneBalance, user.Balance, "user balance not equeal (%f != %f)", userOneBalance, user.Balance)
	}

	user, err = controller.GetUserByCredentials(userOneLogin, "superSecrets")
	assert.ErrorIs(t, err, ErrIncorrectUserPassword, "method must be returned incorrect password error")
	assert.Nil(t, user, "user pointer must be nil on incorrect password find")

	user, err = controller.GetUserByCredentials("userTwo", userOnePassword)
	assert.ErrorIs(t, err, ErrUserNotFound, "method must be returned user not found error")
	assert.Nil(t, user, "user pointer must be nil on incorrect login find")
}

func Test_buildPasswordHash(t *testing.T) {
	superSecret := "superSecret"
	superSecretHash := "056b7fe47141b6e48e87caf8f8e5bb92120ac12c6e6944cf7dbcda2db23581cc"

	superSecrets := "superSecrets"

	h := buildPasswordHash(superSecret)
	assert.Truef(t, h == superSecretHash, "hash must be equal (%s == %s)", h, superSecretHash)

	h = buildPasswordHash(superSecrets)
	assert.Falsef(t, h == superSecretHash, "hash must be not equal (%s != %s)", h, superSecretHash)

	h = buildPasswordHash(superSecret) // second check for invariance
	assert.Truef(t, h == superSecretHash, "hash must be equal (%s == %s)", h, superSecretHash)
}
