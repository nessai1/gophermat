package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/nessai1/gophermat/internal/order"
	"github.com/nessai1/gophermat/internal/user"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type EnrollmentOrderHandler struct {
	Logger               *zap.Logger
	EnrollmentController *order.EnrollmentController
}

type EnrollmentItem struct {
	OrderID    string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float32   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
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

	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(request.Body)
	if err != nil {
		handler.Logger.Debug("user sends invalid request")
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	orderNumber := buffer.String()
	enrollmentOrder, err := handler.EnrollmentController.RequireOrder(request.Context(), orderNumber, ctxUser.ID)
	if err != nil && errors.Is(err, order.ErrInvalidOrderNumber) {
		handler.Logger.Debug("user register order number with invalid format", zap.String("order number", orderNumber))
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if enrollmentOrder.UserID != ctxUser.ID {
		handler.Logger.Debug(
			"user register someone else's order",
			zap.String("order number", orderNumber),
			zap.Int("request user id", ctxUser.ID),
			zap.Int("order owner user id", ctxUser.ID),
		)

		writer.WriteHeader(http.StatusConflict)
		return
	}

	if enrollmentOrder.Status == order.EnrollmentStatusNew {
		handler.Logger.Debug(
			"user successful load new order",
			zap.String("order number", orderNumber),
			zap.Int("user id", ctxUser.ID),
		)
		writer.WriteHeader(http.StatusAccepted)
		return
	}

	handler.Logger.Debug(
		"user try to load already exists own order",
		zap.String("order number", orderNumber),
		zap.Int("user id", ctxUser.ID),
	)

	writer.WriteHeader(http.StatusOK)
}

func (handler *EnrollmentOrderHandler) HandleGetOrders(writer http.ResponseWriter, request *http.Request) {
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

	enrollmentList, err := handler.EnrollmentController.GetUserOrderListByID(request.Context(), ctxUser.ID)
	if err != nil {
		handler.Logger.Error("error while get enrollment list for user", zap.Int("user id", ctxUser.ID), zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
	}

	if len(enrollmentList) == 0 {
		writer.WriteHeader(http.StatusNoContent)
	}

	resultEnrollmentList := make([]EnrollmentItem, len(enrollmentList))
	for i := 0; i < len(enrollmentList); i++ {
		item := EnrollmentItem{
			OrderID:    enrollmentList[i].OrderID,
			Status:     enrollmentList[i].Status,
			Accrual:    float32(enrollmentList[i].Accrual) / 100,
			UploadedAt: enrollmentList[i].UploadedAt,
		}

		resultEnrollmentList[i] = item
	}

	rs, err := json.Marshal(resultEnrollmentList)
	if err != nil {
		handler.Logger.Error("error while marshal list of user enrollments", zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(rs)
	if err != nil {
		handler.Logger.Error("error while write enrollment list to result body", zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
	}
}
