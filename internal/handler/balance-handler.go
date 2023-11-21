package handler

import (
	"go.uber.org/zap"
	"net/http"
)

type BalanceHandler struct {
	Logger *zap.Logger
}

func (handler *BalanceHandler) HandleGetBalance(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Getting balance...."))
}

func (handler *BalanceHandler) HandleWithdraw(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Withdraw...."))
}
