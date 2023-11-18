package handler

import (
	"github.com/nessai1/gophermat/internal/user"
	"net/http"

	"go.uber.org/zap"
)

type AuthHandler struct {
	Logger         *zap.Logger
	SecretKey      string
	UserController *user.Controller
}

func (handler *AuthHandler) HandleAuthUser(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("auth"))
}

func (handler *AuthHandler) HandleRegisterUser(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("register"))
}

func (handler *AuthHandler) MiddlewareAuthorizeRequest() func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			handler.Logger.Info("Got middleware")
			next.ServeHTTP(writer, request)
		})
	}
}
