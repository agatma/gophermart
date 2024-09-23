package rest

import (
	"context"
	"encoding/json"
	"errors"
	"gophermart/cmd/pkg/errs"
	"gophermart/internal/core/domain"
	"gophermart/internal/logger"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID int, order *domain.OrderIn) error
	GetAllOrders(ctx context.Context, userID int) (domain.OrderOutList, error)
}

func (h *Handler) CreateOrder(w http.ResponseWriter, req *http.Request) {
	orderNumber, err := io.ReadAll(req.Body)
	if err != nil || len(orderNumber) == 0 {
		logger.Log.Info("cannot read body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if err = closeBody(req); err != nil {
		logger.Log.Info("cannot close body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID, err := getUserID(req)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if err = h.service.CreateOrder(req.Context(), userID, &domain.OrderIn{Number: string(orderNumber)}); err != nil {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, errs.ErrOrderAlreadyAdded):
			statusCode = http.StatusOK
		case errors.Is(err, errs.ErrUnreachableOrder):
			statusCode = http.StatusConflict
		case errors.Is(err, errs.ErrInvalidOrderNumber):
			statusCode = http.StatusUnprocessableEntity
		case errors.Is(err, errs.ErrOrderAlreadyExist):
			statusCode = http.StatusUnprocessableEntity
		}
		logger.Log.Error("error occurred", zap.Error(err))
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) GetAllOrders(w http.ResponseWriter, req *http.Request) {
	userID, err := getUserID(req)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	orders, err := h.service.GetAllOrders(req.Context(), userID)
	if err != nil {
		logger.Log.Error("error occurred", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set(contentType, applicationJSON)
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		if err = json.NewEncoder(w).Encode(orders); err != nil {
			logger.Log.Error("error encoding response", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}

}
