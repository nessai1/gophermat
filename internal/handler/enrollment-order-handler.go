package handler

import (
	"github.com/nessai1/gophermat/internal/order"
	"go.uber.org/zap"
	"net/http"
)

type EnrollmentOrderHandler struct {
	Logger               *zap.Logger
	EnrollmentController *order.EnrollmentController
}

func (handler *EnrollmentOrderHandler) HandleLoadOrders(writer http.ResponseWriter, request *http.Request) {

}

func (handler *EnrollmentOrderHandler) HandGetOrders(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Getting orders...."))
}
