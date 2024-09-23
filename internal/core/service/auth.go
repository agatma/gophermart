package service

import (
	"context"
	"errors"
	"gophermart/internal/config"
	"gophermart/internal/core/domain"
	"gophermart/internal/shared-kernel/hash"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type AuthStorage interface {
	CreateUser(ctx context.Context, user *domain.UserIn) error
	GetUserId(ctx context.Context, user *domain.UserIn) (int, error)
}

type AuthService struct {
	storage AuthStorage
	config  *config.Config
}

func newAuthService(storage AuthStorage, config *config.Config) *AuthService {
	return &AuthService{storage: storage, config: config}
}

func (auth *AuthService) CreateUser(ctx context.Context, user *domain.UserIn) error {
	user.Password = hash.Encode([]byte(user.Password), auth.config.HashKey)
	return auth.storage.CreateUser(ctx, user)
}

func (auth *AuthService) CreateToken(ctx context.Context, user *domain.UserIn) (string, error) {
	user.Password = hash.Encode([]byte(user.Password), auth.config.HashKey)
	userId, err := auth.storage.GetUserId(ctx, user)
	if err != nil {
		return "", err
	}
	tokenClaim := domain.TokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(auth.config.TokenTTLSeconds) * time.Second).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		UserID: userId,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaim)
	return token.SignedString([]byte(auth.config.TokenKey))
}

func (auth *AuthService) GetUserId(accessToken string) (int, error) {
	token, err := jwt.ParseWithClaims(accessToken, &domain.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(auth.config.TokenKey), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*domain.TokenClaims)
	if !ok {
		return 0, errors.New("token is not valid")
	}
	if claims.StandardClaims.ExpiresAt < time.Now().Unix() {
		return 0, errors.New("token has expired")
	}
	return claims.UserID, nil
}
