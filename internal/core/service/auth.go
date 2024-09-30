package service

import (
	"context"
	"errors"
	"fmt"
	"gophermart/internal/config"
	"gophermart/internal/core/domain"
	"gophermart/internal/shared-kernel/hash"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type AuthService struct {
	storage Storage
	config  *config.Config
}

func newAuthService(storage Storage, config *config.Config) *AuthService {
	return &AuthService{storage: storage, config: config}
}

func (auth *AuthService) CreateUser(ctx context.Context, user *domain.UserIn) error {
	user.PasswordHash = hash.Encode([]byte(user.Password), auth.config.HashKey)
	if err := auth.storage.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("could not create user: %w", err)
	}
	return nil
}

func (auth *AuthService) CreateToken(ctx context.Context, user *domain.UserIn) (string, error) {
	user.PasswordHash = hash.Encode([]byte(user.Password), auth.config.HashKey)
	userID, err := auth.storage.GetUserID(ctx, user)
	if err != nil {
		return "", fmt.Errorf("could not get user id: %w", err)
	}
	tokenClaim := domain.TokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(auth.config.TokenTTLSeconds) * time.Second).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		UserID: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaim)
	signedToken, err := token.SignedString([]byte(auth.config.TokenKey))
	if err != nil {
		return "", fmt.Errorf("could not sign token: %w", err)
	}
	return signedToken, nil
}

func (auth *AuthService) GetUserID(accessToken string) (int, error) {
	token, err := jwt.ParseWithClaims(accessToken, &domain.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(auth.config.TokenKey), nil
	})
	if err != nil {
		return 0, fmt.Errorf("could not parse token: %w", err)
	}
	claims, ok := token.Claims.(*domain.TokenClaims)
	if !ok {
		return 0, errors.New("token is not valid")
	}
	if claims.StandardClaims.ExpiresAt < time.Now().Unix() {
		return 0, errors.New("token expired")
	}
	return claims.UserID, nil
}
