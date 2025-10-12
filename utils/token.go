package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenClient interface {
	GenerateToken(req GenerateTokenReq) (GenerateTokenResp, error)
	ValidateToken(tokenString string) (*uuid.UUID, error)
}

type tokenClient struct {
	secret      string
	expiryHours int
}

type GenerateTokenReq struct {
	UserID string
}

type GenerateTokenResp struct {
	Token    string
	ExpToken int64
}

// NewToken creates a new token client
func NewToken(secret string, expiryHours int) TokenClient {
	return &tokenClient{
		secret:      secret,
		expiryHours: expiryHours,
	}
}

// GenerateToken generates a JWT token for a user
func (t *tokenClient) GenerateToken(req GenerateTokenReq) (GenerateTokenResp, error) {
	expTime := time.Now().Add(time.Hour * time.Duration(t.expiryHours))
	expToken := expTime.Unix()

	claims := jwt.MapClaims{
		"user_id": req.UserID,
		"exp":     expToken,
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(t.secret))
	if err != nil {
		return GenerateTokenResp{}, err
	}

	return GenerateTokenResp{
		Token:    tokenString,
		ExpToken: expToken,
	}, nil
}

// ValidateToken validates a JWT token and returns the user ID
func (t *tokenClient) ValidateToken(tokenString string) (*uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return nil, errors.New("invalid token claims")
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, errors.New("invalid user ID in token")
		}

		return &userID, nil
	}

	return nil, errors.New("invalid token")
}
