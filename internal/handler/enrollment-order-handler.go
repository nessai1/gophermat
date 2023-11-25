package handler

import (
	"github.com/nessai1/gophermat/internal/order"
	"github.com/nessai1/gophermat/internal/user"
	"go.uber.org/zap"
	"net/http"
)

type EnrollmentOrderHandler struct {
	Logger               *zap.Logger
	EnrollmentController *order.EnrollmentController
}

func (handler *EnrollmentOrderHandler) HandleLoadOrders(writer http.ResponseWriter, request *http.Request) {
	ctxUserVal := request.Context().Value(AuthorizeUserContext)
	if ctxUserVal == nil {
		handler.Logger.Error("load orders handler must have user in context, but not found")
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctxUser, ok := ctxUserVal.(*user.User)
	if !ok {
		handler.Logger.Error("cannot cast user in request context while load order")
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Write([]byte(ctxUser.Login))
}

func (handler *EnrollmentOrderHandler) HandGetOrders(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Getting orders...."))
}
