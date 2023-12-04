package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/nessai1/gophermat/internal/order"
	"github.com/nessai1/gophermat/internal/user"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type BalanceHandler struct {
	Logger             *zap.Logger
	WithdrawController *order.WithdrawController
}

type AddWithdrawRequest struct {
	Order string      `json:"order"`
	Sum   json.Number `json:"sum"`
}

type GetWithdrawResponse struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type GetWithdrawItemResponse struct {
	Order       string    `json:"order"`
	Sum         float32   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (handler *BalanceHandler) HandleGetBalance(writer http.ResponseWriter, request *http.Request) {
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

	withdrawSum, err := handler.WithdrawController.GetWithdrawSumByUser(request.Context(), ctxUser)
	if err != nil {
		handler.Logger.Error("error while get withdraw sum for user", zap.Int("user id", ctxUser.ID), zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	res := GetWithdrawResponse{
		Current:   float32(ctxUser.Balance) / 100,
		Withdrawn: float32(withdrawSum) / 100,
	}

	body, err := json.Marshal(res)
	if err != nil {
		handler.Logger.Error("cannot marshal info of user balance", zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(body)
	if err != nil {
		handler.Logger.Error("error while write ifo of user balance to body")
		writer.WriteHeader(http.StatusInternalServerError)
	}
}

func (handler *BalanceHandler) HandleAddWithdraw(writer http.ResponseWriter, request *http.Request) {
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

	bf := bytes.Buffer{}
	_, err := bf.ReadFrom(request.Body)
	if err != nil {
		handler.Logger.Debug("user sends invalid body to withdraw", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	var requestWithdraw AddWithdrawRequest
	err = json.Unmarshal(bf.Bytes(), &requestWithdraw)
	if err != nil || requestWithdraw.Sum == "" || requestWithdraw.Order == "" {
		handler.Logger.Debug("user sends invalid withdraw request object", zap.Error(err))
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	sum, err := user.ParseBalance(string(requestWithdraw.Sum))
	if err != nil {
		handler.Logger.Debug("user sends invalid sum of withdraw", zap.Error(err))
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	withdraw, err := handler.WithdrawController.CreateWithdrawByUser(request.Context(), ctxUser, requestWithdraw.Order, sum)
	if err != nil {
		if errors.Is(err, order.ErrInvalidOrderNumber) {
			writer.WriteHeader(http.StatusUnprocessableEntity)
			handler.Logger.Debug("user send invalid format of order number", zap.String("order number", requestWithdraw.Order))
		} else if errors.Is(err, order.ErrNoMoney) {
			writer.WriteHeader(http.StatusPaymentRequired)
			handler.Logger.Debug("user has no money to withdraw", zap.Int64("user balance", ctxUser.Balance), zap.Int64("required sum", sum))
		} else if errors.Is(err, order.ErrEmptyBalance) {
			writer.WriteHeader(http.StatusBadRequest)
			handler.Logger.Debug("user sends empty sum to withdraw")
		} else {
			writer.WriteHeader(http.StatusInternalServerError)
			handler.Logger.Error("server error while create withdraw", zap.Error(err))
		}

		return
	}

	handler.Logger.Debug("user successful register withdraw", zap.String("order ID", withdraw.OrderID), zap.Int64("sum", withdraw.Sum))
	writer.WriteHeader(http.StatusOK)
}

func (handler *BalanceHandler) HandleGetListWithdraw(writer http.ResponseWriter, request *http.Request) {
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

	withdrawList, err := handler.WithdrawController.GetWithdrawByUser(request.Context(), ctxUser)
	if err != nil {
		handler.Logger.Error("error while get list of withdraw", zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(withdrawList) == 0 {
		handler.Logger.Debug("user get list of withdraw but he have no one withdraw", zap.Int("user id", ctxUser.ID))
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	responseBody := make([]GetWithdrawItemResponse, len(withdrawList))
	for i := 0; i < len(withdrawList); i++ {
		responseBody[i] = GetWithdrawItemResponse{
			Order:       withdrawList[i].OrderID,
			Sum:         float32(withdrawList[i].Sum) / 100,
			ProcessedAt: withdrawList[i].ProcessedAt,
		}
	}

	rs, err := json.Marshal(responseBody)
	if err != nil {
		handler.Logger.Error("error while marshal list of withdraw", zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(rs)
	if err != nil {
		handler.Logger.Error("error while write list of withdraw", zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
	}
}
