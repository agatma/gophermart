package rest

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	contentType     = "Content-Type"
	applicationJSON = "application/json"
	authorization   = "Authorization"
	userIDKey       = "userID"
)

func closeBody(req *http.Request) error {
	if _, err := io.Copy(io.Discard, req.Body); err != nil {
		return fmt.Errorf("discard body error: %w", err)
	}
	if err := req.Body.Close(); err != nil {
		return fmt.Errorf("close body error: %w", err)
	}
	return nil
}

func getUserID(req *http.Request) (int, error) {
	user := req.Header.Get(userIDKey)
	if user == "" {
		return 0, errors.New("userID not found")
	}
	userID, err := strconv.Atoi(user)
	if err != nil {
		return 0, errors.New("userID has invalid type")
	}

	return userID, nil
}
