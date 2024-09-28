package rest

import (
	"encoding/json"
	"errors"
	"gophermart/internal/adapters/api/validation"
	"gophermart/internal/core/domain"
	"gophermart/internal/errs"
	"gophermart/internal/logger"
	"net/http"

	"go.uber.org/zap"
)

func (h *Handler) GetAllWithdrawals(w http.ResponseWriter, req *http.Request) {
	userID, err := getUserID(req)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	withdrawals, err := h.service.GetAllWithdrawals(req.Context(), userID)
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	} else if err != nil {
		logger.Log.Error("error occurred during getting all withdrawals", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set(contentType, applicationJSON)
	if err = json.NewEncoder(w).Encode(withdrawals); err != nil {
		logger.Log.Error("error encoding withdrawals", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *Handler) GetBalance(w http.ResponseWriter, req *http.Request) {
	userID, err := getUserID(req)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	balance, err := h.service.GetBalance(req.Context(), userID)
	if err != nil {
		logger.Log.Error("error occurred during getting balance", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set(contentType, applicationJSON)
	if err = json.NewEncoder(w).Encode(balance); err != nil {
		logger.Log.Error("error encoding balance", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *Handler) WithdrawBonuses(w http.ResponseWriter, req *http.Request) {
	var withdraw domain.WithdrawalIn
	userID, err := getUserID(req)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	if err = json.NewDecoder(req.Body).Decode(&withdraw); err != nil {
		logger.Log.Info("cannot decode withdraw JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err = validation.ValidateWithdrawIn(&withdraw); err != nil {
		logger.Log.Info("withdraw validation error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.service.WithdrawBonuses(req.Context(), userID, &withdraw)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	statusCode := http.StatusInternalServerError
	switch {
	case errors.Is(err, errs.ErrNotEnoughFunds):
		statusCode = http.StatusPaymentRequired
	case errors.Is(err, errs.ErrInvalidOrderNumber):
		statusCode = http.StatusUnprocessableEntity
	}
	logger.Log.Error("error occurred", zap.Error(err))
	http.Error(w, http.StatusText(statusCode), statusCode)
}
