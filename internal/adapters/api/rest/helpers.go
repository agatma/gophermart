package rest

import (
	"errors"
	"net/http"
	"strconv"
)

const (
	contentType     = "Content-Type"
	applicationJSON = "application/json"
	authorization   = "Authorization"
	userIDKey       = "userID"
)

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
