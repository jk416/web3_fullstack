package token

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenMaker interface {
	CreateToken(userID uint) (string, error)
	VerifyToken(token string) (*Claims, error)
}
