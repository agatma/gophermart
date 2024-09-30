package domain

import "github.com/dgrijalva/jwt-go"

type TokenClaims struct {
	jwt.StandardClaims
	UserID int `json:"user_id"`
}

type Token struct {
	Token string `json:"token"`
}
