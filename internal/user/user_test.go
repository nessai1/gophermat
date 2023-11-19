package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	controller := NewController(&MapRepository{data: users})
	ctx := context.TODO()

	user, err := controller.GetUserByCredentials(ctx, userOneLogin, userOnePassword)
	if assert.NoErrorf(t, err, "user with login %s must be found", userOneLogin) {
		assert.Equalf(t, userOneLogin, user.Login, "user login not equeal (%s != %s)", userOneLogin, user.Login)
		assert.Equalf(t, userOneBalance, user.Balance, "user balance not equeal (%f != %f)", userOneBalance, user.Balance)
	}

	user, err = controller.GetUserByCredentials(ctx, userOneLogin, "superSecrets")
	assert.ErrorIs(t, err, ErrIncorrectUserPassword, "method must be returned incorrect password error")
	assert.Nil(t, user, "user pointer must be nil on incorrect password find")

	user, err = controller.GetUserByCredentials(ctx, "userTwo", userOnePassword)
	assert.ErrorIs(t, err, ErrUserNotFound, "method must be returned user not found error")
	assert.Nil(t, user, "user pointer must be nil on incorrect login find")
}

func TestController_AddUser(t *testing.T) {
	repository := MapRepository{data: map[string]User{}}
	controller := NewController(&repository)

	userLogin := "userOne"
	userPassword := "passwordOne"

	ctx := context.TODO()

	_, err := controller.GetUserByCredentials(ctx, userLogin, userPassword)
	require.ErrorIs(t, err, ErrUserNotFound)

	user, err := controller.AddUser(ctx, userLogin, userPassword)
	require.NoError(t, err)
	require.Equalf(t, len(repository.data), 1, "repository len must be 1, got %d", len(repository.data))

	expectedUser := User{
		Login:    userLogin,
		Balance:  0,
		password: buildPasswordHash(userPassword),
	}

	assert.Equal(t, expectedUser, *user, "created and expected users are not equal")

	secondUser, err := controller.GetUserByCredentials(ctx, userLogin, userPassword)
	require.NoError(t, err)
	assert.Equal(t, *secondUser, *user)

	_, err = controller.AddUser(ctx, userLogin, userPassword)
	require.ErrorIs(t, err, ErrLoginAlreadyExists)
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
