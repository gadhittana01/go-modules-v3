package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type TokenClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

type TokenClient interface {
	GenerateToken(req GenerateTokenReq) (GenerateTokenResp, error)
	ValidateToken(tokenString string) (*TokenClaims, error)
}

type tokenClient struct {
	secret      string
	expiryHours int
}

type GenerateTokenReq struct {
	UserID   string
	Username string
}

type GenerateTokenResp struct {
	Token    string
	ExpToken int64
}

type TokenPairResp struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // seconds until access token expires
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
		"user_id":  req.UserID,
		"username": req.Username,
		"exp":      expToken,
		"iat":      time.Now().Unix(),
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

// ValidateToken validates a JWT token and returns the claims
func (t *tokenClient) ValidateToken(tokenString string) (*TokenClaims, error) {
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
		userID, ok := claims["user_id"].(string)
		if !ok {
			return nil, errors.New("invalid user_id in token claims")
		}

		username, ok := claims["username"].(string)
		if !ok {
			return nil, errors.New("invalid username in token claims")
		}

		return &TokenClaims{
			UserID:   userID,
			Username: username,
		}, nil
	}

	return nil, errors.New("invalid token")
}

// Global token client instance
var globalTokenClient TokenClient

// SetGlobalTokenClient sets the global token client
func SetGlobalTokenClient(client TokenClient) {
	globalTokenClient = client
}

// ValidateToken validates a JWT token using the global token client
func ValidateToken(tokenString string) (*TokenClaims, error) {
	if globalTokenClient == nil {
		return nil, errors.New("token client not initialized")
	}
	return globalTokenClient.ValidateToken(tokenString)
}

// GenerateToken generates a JWT token using the global token client
func GenerateToken(req GenerateTokenReq) (GenerateTokenResp, error) {
	if globalTokenClient == nil {
		return GenerateTokenResp{}, errors.New("token client not initialized")
	}
	return globalTokenClient.GenerateToken(req)
}

// Redis-based token management
type RedisTokenManager struct {
	redisClient *redis.Client
	secret      string
	expiryHours int
}

// NewRedisTokenManager creates a new Redis-based token manager
func NewRedisTokenManager(redisClient *redis.Client, secret string, expiryHours int) *RedisTokenManager {
	return &RedisTokenManager{
		redisClient: redisClient,
		secret:      secret,
		expiryHours: expiryHours,
	}
}

// StoreToken stores a JWT token in Redis with user_id as key
func (rtm *RedisTokenManager) StoreToken(ctx context.Context, userID, token string) error {
	key := fmt.Sprintf("token:%s", userID)
	expiration := time.Duration(rtm.expiryHours) * time.Hour
	return rtm.redisClient.Set(ctx, key, token, expiration).Err()
}

// ValidateToken validates a JWT token by checking Redis
func (rtm *RedisTokenManager) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	// First, parse the JWT token to get user_id
	claims, err := rtm.parseJWTToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT token: %w", err)
	}

	// Check if token exists in Redis
	key := fmt.Sprintf("token:%s", claims.UserID)
	storedToken, err := rtm.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("token not found in Redis - user may have logged out")
		}
		return nil, fmt.Errorf("Redis error: %w", err)
	}

	// Compare tokens
	if storedToken != tokenString {
		return nil, errors.New("token mismatch - invalid session")
	}

	return claims, nil
}

// RevokeToken removes a token from Redis (for logout)
func (rtm *RedisTokenManager) RevokeToken(ctx context.Context, userID string) error {
	key := fmt.Sprintf("token:%s", userID)
	return rtm.redisClient.Del(ctx, key).Err()
}

// parseJWTToken parses a JWT token and returns claims
func (rtm *RedisTokenManager) parseJWTToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(rtm.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		if !ok {
			return nil, errors.New("invalid user_id in token claims")
		}

		username, ok := claims["username"].(string)
		if !ok {
			return nil, errors.New("invalid username in token claims")
		}

		return &TokenClaims{
			UserID:   userID,
			Username: username,
		}, nil
	}

	return nil, errors.New("invalid token")
}

// Global Redis token manager instance
var globalRedisTokenManager *RedisTokenManager

// SetGlobalRedisTokenManager sets the global Redis token manager
func SetGlobalRedisTokenManager(manager *RedisTokenManager) {
	globalRedisTokenManager = manager
}

// ValidateTokenWithRedis validates a JWT token using Redis
func ValidateTokenWithRedis(ctx context.Context, tokenString string) (*TokenClaims, error) {
	if globalRedisTokenManager == nil {
		return nil, errors.New("Redis token manager not initialized")
	}
	return globalRedisTokenManager.ValidateToken(ctx, tokenString)
}

// StoreTokenInRedis stores a token in Redis
func StoreTokenInRedis(ctx context.Context, userID, token string) error {
	if globalRedisTokenManager == nil {
		return errors.New("Redis token manager not initialized")
	}
	return globalRedisTokenManager.StoreToken(ctx, userID, token)
}

// RevokeTokenFromRedis removes a token from Redis
func RevokeTokenFromRedis(ctx context.Context, userID string) error {
	if globalRedisTokenManager == nil {
		return errors.New("Redis token manager not initialized")
	}
	return globalRedisTokenManager.RevokeToken(ctx, userID)
}

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(req GenerateTokenReq) (TokenPairResp, error) {
	if globalRedisTokenManager == nil {
		return TokenPairResp{}, errors.New("Redis token manager not initialized")
	}

	// Access token: 15 minutes
	accessExpTime := time.Now().Add(15 * time.Minute)
	accessExpToken := accessExpTime.Unix()
	accessClaims := jwt.MapClaims{
		"user_id":  req.UserID,
		"username": req.Username,
		"exp":      accessExpToken,
		"iat":      time.Now().Unix(),
		"type":     "access",
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(globalRedisTokenManager.secret))
	if err != nil {
		return TokenPairResp{}, err
	}

	// Refresh token: 7 days
	refreshExpTime := time.Now().Add(7 * 24 * time.Hour)
	refreshExpToken := refreshExpTime.Unix()
	refreshClaims := jwt.MapClaims{
		"user_id":  req.UserID,
		"username": req.Username,
		"exp":      refreshExpToken,
		"iat":      time.Now().Unix(),
		"type":     "refresh",
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(globalRedisTokenManager.secret))
	if err != nil {
		return TokenPairResp{}, err
	}

	return TokenPairResp{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    900, // 15 minutes in seconds
	}, nil
}

// StoreRefreshTokenInRedis stores a refresh token in Redis
func StoreRefreshTokenInRedis(ctx context.Context, userID, token string) error {
	if globalRedisTokenManager == nil {
		return errors.New("Redis token manager not initialized")
	}
	key := fmt.Sprintf("refresh_token:%s", userID)
	expiration := 7 * 24 * time.Hour // 7 days
	return globalRedisTokenManager.redisClient.Set(ctx, key, token, expiration).Err()
}

// ValidateRefreshToken validates a refresh token
func ValidateRefreshToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	if globalRedisTokenManager == nil {
		return nil, errors.New("Redis token manager not initialized")
	}

	// First, parse the JWT token to get user_id and check if it's a refresh token
	claims, err := globalRedisTokenManager.parseJWTToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT token: %w", err)
	}

	// Check if token exists in Redis
	key := fmt.Sprintf("refresh_token:%s", claims.UserID)
	storedToken, err := globalRedisTokenManager.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("refresh token not found - user may have logged out")
		}
		return nil, fmt.Errorf("Redis error: %w", err)
	}

	// Compare tokens
	if storedToken != tokenString {
		return nil, errors.New("refresh token mismatch - invalid session")
	}

	return claims, nil
}

// RevokeRefreshTokenFromRedis removes a refresh token from Redis
func RevokeRefreshTokenFromRedis(ctx context.Context, userID string) error {
	if globalRedisTokenManager == nil {
		return errors.New("Redis token manager not initialized")
	}
	key := fmt.Sprintf("refresh_token:%s", userID)
	return globalRedisTokenManager.redisClient.Del(ctx, key).Err()
}

// RevokeAllTokens removes both access and refresh tokens
func RevokeAllTokens(ctx context.Context, userID string) error {
	if err := RevokeTokenFromRedis(ctx, userID); err != nil {
		return err
	}
	return RevokeRefreshTokenFromRedis(ctx, userID)
}
