package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nessai1/gophermat/internal/user"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

const authCookieName = "GOPHERMAT_JWT"
const AuthorizeUserContext AuthorizeUserContextKey = "AuthorizeUserContext"
const TokenTTL = time.Hour * 24

type AuthorizeUserContextKey string

var ErrWrongSign = errors.New("got wrong sign")

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type userJWTClaims struct {
	jwt.RegisteredClaims
	Login string
}

type AuthHandler struct {
	Logger         *zap.Logger
	SecretKey      string
	UserController *user.Controller
}

func (handler *AuthHandler) HandleAuthUser(writer http.ResponseWriter, request *http.Request) {

	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(request.Body)
	if err != nil {
		handler.Logger.Debug("client sends invalid request", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	var credentials UserCredentials
	err = json.Unmarshal(buffer.Bytes(), &credentials)
	if err != nil {
		handler.Logger.Debug("cannot unmarshal client request", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if credentials.Password == "" || credentials.Login == "" {
		handler.Logger.Debug("user sends empty login/password")
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	fetchedUser, err := handler.UserController.GetUserByCredentials(request.Context(), credentials.Login, credentials.Password)
	if err != nil && (errors.Is(err, user.ErrUserNotFound) || errors.Is(err, user.ErrIncorrectUserPassword)) {
		handler.Logger.Debug("user send invalid credentials on login")
		writer.WriteHeader(http.StatusUnauthorized)
		return
	} else if err != nil {
		handler.Logger.Error("error while get user on login", zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	sign, err := handler.createSign(fetchedUser)
	if err != nil {
		handler.Logger.Error("error while sign created account", zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c := &http.Cookie{Name: authCookieName, Value: sign}
	http.SetCookie(writer, c)

	ctx := request.Context()
	request.WithContext(context.WithValue(ctx, AuthorizeUserContext, fetchedUser))

	writer.WriteHeader(http.StatusOK)
}

func (handler *AuthHandler) HandleRegisterUser(writer http.ResponseWriter, request *http.Request) {

	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(request.Body)
	if err != nil {
		handler.Logger.Debug("client sends invalid request", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	var credentials UserCredentials
	err = json.Unmarshal(buffer.Bytes(), &credentials)
	if err != nil {
		handler.Logger.Debug("cannot unmarshal client request", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if credentials.Password == "" || credentials.Login == "" {
		handler.Logger.Debug("user sends empty login/password")
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	createdUser, err := handler.UserController.AddUser(request.Context(), credentials.Login, credentials.Password)
	if err != nil && errors.Is(err, user.ErrLoginAlreadyExists) {
		handler.Logger.Debug("user try register existing account", zap.String("login", credentials.Login))
		writer.WriteHeader(http.StatusConflict)
		return
	} else if err != nil {
		handler.Logger.Error("error while create new user by register", zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	sign, err := handler.createSign(createdUser)
	if err != nil {
		handler.Logger.Error("error while sign created account", zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	c := &http.Cookie{Name: authCookieName, Value: sign}
	http.SetCookie(writer, c)

	ctx := request.Context()
	request.WithContext(context.WithValue(ctx, AuthorizeUserContext, createdUser))

	writer.WriteHeader(http.StatusOK)
}

func (handler *AuthHandler) MiddlewareAuthorizeRequest() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			cookie, err := request.Cookie(authCookieName)
			if err != nil && errors.Is(err, http.ErrNoCookie) {
				handler.Logger.Debug("user has no auth cookie", zap.String("client address", request.RemoteAddr))
				writer.WriteHeader(http.StatusUnauthorized)
				return
			} else if err != nil {
				handler.Logger.Error("undefined error occurred in auth middleware", zap.String("client address", request.RemoteAddr), zap.Error(err))
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			ctx := request.Context()
			authUser, err := handler.fetchUser(ctx, cookie.Value)
			if err != nil && errors.Is(err, ErrWrongSign) {
				handler.Logger.Debug("user sends invalid sign cookie", zap.Error(err))
				c := &http.Cookie{
					Value:  "",
					Name:   authCookieName,
					MaxAge: -1,
				}
				http.SetCookie(writer, c)
				writer.WriteHeader(http.StatusUnauthorized)
				return
			} else if err != nil {
				handler.Logger.Error("error while getting user for auth middleware", zap.Error(err))
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			request.WithContext(context.WithValue(ctx, AuthorizeUserContext, authUser))

			next.ServeHTTP(writer, request)
		})
	}
}

func (handler *AuthHandler) fetchUser(ctx context.Context, sign string) (*user.User, error) {
	claims := &userJWTClaims{}
	_, err := jwt.ParseWithClaims(sign, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(handler.SecretKey), nil
	})

	if err != nil {
		return nil, errors.Join(ErrWrongSign, err)
	}

	fetchedUser, err := handler.UserController.GetUserByLogin(ctx, claims.Login)
	if err != nil && errors.Is(err, user.ErrUserNotFound) {
		return nil, errors.Join(ErrWrongSign, err)
	} else if err != nil {
		return nil, fmt.Errorf("error while fetch user in jwt: %w", err)
	}

	return fetchedUser, nil
}

func (handler *AuthHandler) createSign(signedUser *user.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenTTL)),
		},
		Login: signedUser.Login,
	})

	tokenString, err := token.SignedString([]byte(handler.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
