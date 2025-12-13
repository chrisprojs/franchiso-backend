package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type JWTClaims struct {
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// generate JWT function for access or refresh token
func GenerateJWT(userID, role, tokenType string) (string, error) {
	var expiresAt time.Time
	if tokenType == "access" {
		expiresAt = time.Now().Add(180 * time.Minute) // access token 15 minutes
	} else if tokenType == "refresh" {
		expiresAt = time.Now().Add(7 * 24 * time.Hour) // refresh token 7 days
	} else {
		return "", errors.New("token type tidak valid")
	}

	claims := JWTClaims{
		UserID:    userID,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// validasi JWT function to check token type
func ValidateJWT(tokenString, expectedType string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token tidak valid")
	}
	if claims.TokenType != expectedType {
		return nil, errors.New("token type tidak sesuai")
	}
	return claims, nil
}
