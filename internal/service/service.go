package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/RIBorisov/gophermart/internal/config"
	"github.com/RIBorisov/gophermart/internal/errs"
	"github.com/RIBorisov/gophermart/internal/logger"
	"github.com/RIBorisov/gophermart/internal/models/balance"
	"github.com/RIBorisov/gophermart/internal/models/orders"
	"github.com/RIBorisov/gophermart/internal/models/register"
	"github.com/RIBorisov/gophermart/internal/storage"
)

type Service struct {
	Log     *logger.Log
	Storage storage.Store
	Config  *config.Config
}

func encrypt(secret, data string) (string, error) {
	// хешируем пароль с ключом
	h := sha256.New()
	h.Write([]byte(data + secret))
	hashedPassword := h.Sum(nil)
	// кодируем в base64 строку для хранения в бд
	encodedPassword := base64.StdEncoding.EncodeToString(hashedPassword)

	return encodedPassword, nil
}

func decryptAndCompare(secret, encodedData, password string) error {
	// декодируем закодированную строку
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("failed to decode encoded data: %w", err)
	}

	// вычисляем хеш
	h := sha256.New()
	h.Write([]byte(password + secret))
	expectedHash := h.Sum(nil)

	// сравниваем хеши, если !ok, то неверный пароль
	if !hmac.Equal(decodedBytes, expectedHash) {
		return errs.ErrIncorrectPassword
	}
	return nil
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func (s *Service) BuildJWTString(secretKey string, userID string) (string, error) {
	const tokenExp = time.Hour * 720

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
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

func (s *Service) RegisterUser(ctx context.Context, user *register.Request) (string, error) {
	encrypted, err := encrypt(s.Config.Secret.SecretKey, user.Password)
	if err != nil {
		return "", fmt.Errorf("failed encrypt user password: %w", err)
	}
	user.Password = encrypted

	userID, err := s.Storage.SaveUser(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed register user: %w", err)
	}

	authToken, err := s.BuildJWTString(s.Config.Secret.SecretKey, userID)
	if err != nil {
		return "", fmt.Errorf("failed generate authorization token: %w", err)
	}

	return authToken, nil
}

func (s *Service) LoginUser(ctx context.Context, user *register.Request) (string, error) {
	fromDB, err := s.Storage.GetUser(ctx, user.Login)
	if err != nil {
		return "", fmt.Errorf("failed get user from DB: %w", err)
	}

	if err := decryptAndCompare(s.Config.Secret.SecretKey, fromDB.Password, user.Password); err != nil {
		return "", errs.ErrIncorrectPassword
	}
	authToken, err := s.BuildJWTString(s.Config.Secret.SecretKey, fromDB.ID)
	if err != nil {
		return "", fmt.Errorf("failed generate authToken: %w", err)
	}

	return authToken, nil
}

func (s *Service) CreateOrder(ctx context.Context, orderNo string) error {
	if err := s.Storage.SaveOrder(ctx, orderNo); err != nil {
		return fmt.Errorf("failed save order: %w", err)
	}

	return nil
}

func (s *Service) GetOrders(ctx context.Context) ([]orders.Order, error) {
	var list []orders.Order
	raw, err := s.Storage.GetOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed get orders from storage: %w", err)
	}
	for _, o := range raw {
		list = append(list, orders.Order{
			Number:     o.OrderID,
			Status:     o.Status,
			Accrual:    o.Bonus,
			UploadedAt: o.UploadedAt,
		})
	}

	return list, nil
}

func (s *Service) GetBalance(ctx context.Context) (balance.Response, error) {
	raw, err := s.Storage.GetBalance(ctx)
	if err != nil {
		return balance.Response{}, fmt.Errorf("failed get balance from storage: %w", err)
	}

	return balance.Response{Current: raw.Current, Withdrawn: raw.Withdrawn}, nil
}
