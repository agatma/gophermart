package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"gophermart/cmd/pkg/errs"
	"gophermart/internal/adapters/api/validation"
	"gophermart/internal/core/domain"
	"gophermart/internal/logger"
	"net/http"

	"go.uber.org/zap"
)

func (h *Handler) SignUp(w http.ResponseWriter, req *http.Request) {
	var user domain.UserIn
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		logger.Log.Info("cannot decode userIn JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := validation.ValidateUserIn(&user); err != nil {
		logger.Log.Info("userIn validation error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := closeBody(req); err != nil {
		logger.Log.Info("cannot close body in signUp", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := h.service.CreateUser(req.Context(), &user); err != nil {
		handleAuthError(w, err)
		return
	}
	h.createToken(w, req, &user)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) SignIn(w http.ResponseWriter, req *http.Request) {
	var user domain.UserIn
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		logger.Log.Info("cannot decode userIn JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := validation.ValidateUserIn(&user); err != nil {
		logger.Log.Info("userIn validation error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := closeBody(req); err != nil {
		logger.Log.Info("cannot close body in signIn", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	h.createToken(w, req, &user)
}

func (h *Handler) createToken(w http.ResponseWriter, req *http.Request, user *domain.UserIn) {
	token, err := h.service.CreateToken(req.Context(), user)
	if err != nil {
		handleAuthError(w, err)
		return
	}
	w.Header().Set(authorization, fmt.Sprintf("Bearer %s", token))
	w.Header().Set(contentType, applicationJSON)
	if err = json.NewEncoder(w).Encode(domain.Token{Token: token}); err != nil {
		logger.Log.Error("error encoding token", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func handleAuthError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	if errors.Is(err, errs.ErrLoginAlreadyExist) {
		statusCode = http.StatusConflict
	}
	if errors.Is(err, errs.ErrInvalidLoginOrPassword) {
		statusCode = http.StatusUnauthorized
	}
	logger.Log.Error("error occurred", zap.Error(err))
	http.Error(w, http.StatusText(statusCode), statusCode)
}
