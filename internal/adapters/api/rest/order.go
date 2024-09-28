package rest

import (
	"encoding/json"
	"errors"
	"gophermart/internal/core/domain"
	"gophermart/internal/errs"
	"gophermart/internal/logger"
	"io"
	"net/http"

	"go.uber.org/zap"
)

func (h *Handler) CreateOrder(w http.ResponseWriter, req *http.Request) {
	orderNumber, err := io.ReadAll(req.Body)
	if err != nil || len(orderNumber) == 0 {
		logger.Log.Info("cannot read order number", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	userID, err := getUserID(req)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	err = h.service.CreateOrder(req.Context(), userID, &domain.OrderIn{Number: string(orderNumber)})
	if err == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	statusCode := http.StatusInternalServerError
	switch {
	case errors.Is(err, errs.ErrOrderAlreadyAdded):
		statusCode = http.StatusOK
	case errors.Is(err, errs.ErrUnreachableOrder):
		statusCode = http.StatusConflict
	case errors.Is(err, errs.ErrInvalidOrderNumber):
		statusCode = http.StatusUnprocessableEntity
	}
	logger.Log.Error("error occurred during creating order", zap.Error(err))
	http.Error(w, http.StatusText(statusCode), statusCode)
}

func (h *Handler) GetAllOrders(w http.ResponseWriter, req *http.Request) {
	userID, err := getUserID(req)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	orders, err := h.service.GetAllOrders(req.Context(), userID)
	if err != nil {
		logger.Log.Error("error occurred during getting all orders", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set(contentType, applicationJSON)
	if err = json.NewEncoder(w).Encode(orders); err != nil {
		logger.Log.Error("error encoding orders", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
