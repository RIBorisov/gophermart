package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/RIBorisov/gophermart/internal/config"
	"github.com/RIBorisov/gophermart/internal/logger"
	"github.com/RIBorisov/gophermart/internal/models/register"
	"github.com/RIBorisov/gophermart/internal/storage"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type Service struct {
	Log     *logger.Log
	Storage storage.Store
	Config  *config.Config
}

func (s *Service) RegisterUser(ctx context.Context, user *register.Request) (string, error) {
	encrypted, err := encrypt(s.Config.Secret.SecretKey, user.Password)
	if err != nil {
		return "", fmt.Errorf("failed encrypt user password: %w", err)
	}
	user.Password = encrypted

	userID, err := s.Storage.Register(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed register user: %w", err)
	}
	authToken, err := buildJWTString(s.Config.Secret.SecretKey, userID)
	if err != nil {
		return "", fmt.Errorf("failed generate authToken: %w", err)
	}

	return authToken, nil
}

func encrypt(secret, data string) (string, error) {
	// хешируем пароль с ключом
	h := sha256.New()
	h.Write([]byte(data + secret))
	hashedPassword := h.Sum(nil)
	// кодируем в base64 строку для хранения в бд
	encodedPassword := base64.StdEncoding.EncodeToString(hashedPassword)

	//err := decrypt(secret, encodedPassword, data)
	//if err != nil {
	//	return "", fmt.Errorf("failed decode: %w", err)
	//}
	return encodedPassword, nil
}

func decrypt(secret, encodedData, password string) error {
	// decode encoded data
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("failed to decode encoded data: %w", err)
	}

	// resolving hash
	h := sha256.New()
	h.Write([]byte(password + secret))
	expectedHash := h.Sum(nil)

	// compare hashes, if !ok => user passed invalid password
	if !hmac.Equal(decodedBytes, expectedHash) {
		return storage.ErrIncorrectPassword
	}
	return nil
}

type сlaims struct {
	jwt.RegisteredClaims
	UserID string
}

const tokenExp = time.Hour * 720

func buildJWTString(secretKey string, userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, сlaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		UserID: userID,
	})
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to create token string: %w", err)
	}
	return tokenString, nil
}

//func getUserID(tokenString, secretKey string, log *logger.Log) string {
//	claims := &Claims{}
//	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
//		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
//			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
//		}
//		return []byte(secretKey), nil
//	})
//	if err != nil {
//		log.Err("failed parse with claims tokenString: ", err)
//		return ""
//	}
//	if !token.Valid {
//		log.Err("Token is not valid: ", token)
//		return ""
//	}
//
//	return claims.UserID
//}
