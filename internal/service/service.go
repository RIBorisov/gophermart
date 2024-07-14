package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v4"

	"github.com/RIBorisov/gophermart/internal/config"
	"github.com/RIBorisov/gophermart/internal/errs"
	"github.com/RIBorisov/gophermart/internal/logger"
	accmodels "github.com/RIBorisov/gophermart/internal/models/accrual"
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

func encrypt(secret, data string) string {
	h := sha256.New()
	h.Write([]byte(data + secret))
	hashedPassword := h.Sum(nil)
	encodedPassword := base64.StdEncoding.EncodeToString(hashedPassword)

	return encodedPassword
}

func decryptAndCompare(secret, encodedData, password string) error {
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("failed to decode encoded data: %w", err)
	}

	h := sha256.New()
	h.Write([]byte(password + secret))
	expectedHash := h.Sum(nil)

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
	encrypted := encrypt(s.Config.Secret.SecretKey, user.Password)
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

	if err = decryptAndCompare(s.Config.Secret.SecretKey, fromDB.Password, user.Password); err != nil {
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

func (s *Service) GetUserOrders(ctx context.Context) ([]orders.Order, error) {
	list := make([]orders.Order, 0)
	raw, err := s.Storage.GetUserOrders(ctx)
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
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].UploadedAt.After(list[j].UploadedAt)
	})

	return list, nil
}

func (s *Service) GetBalance(ctx context.Context) (balance.Response, error) {
	raw, err := s.Storage.GetBalance(ctx)
	if err != nil {
		return balance.Response{}, fmt.Errorf("failed get balance from storage: %w", err)
	}

	return balance.Response{Current: raw.Current, Withdrawn: raw.Withdrawn}, nil
}

func (s *Service) BalanceWithdraw(ctx context.Context, withdraw balance.WithdrawRequest) error {
	if err := s.Storage.BalanceWithdraw(ctx, withdraw); err != nil {
		return fmt.Errorf("failed make balance withdraw request: %w", err)
	}
	return nil
}

func (s *Service) GetWithdrawals(ctx context.Context) ([]balance.Withdrawal, error) {
	raw, err := s.Storage.GetWithdrawals(ctx)
	if err != nil {
		return nil, err
	}

	if len(raw) == 0 {
		return nil, errs.ErrNoWithdrawals
	}

	wList := make([]balance.Withdrawal, 0)
	for _, row := range raw {
		fTime, err := time.Parse(time.RFC3339, row.ProcessedAt.Format(time.RFC3339))
		if err != nil {
			return nil, fmt.Errorf("failed parse time into RFC3339: %w", err)
		}
		wList = append(wList, balance.Withdrawal{Order: row.OrderID, Sum: row.Amount, ProcessedAt: fTime})
	}
	sort.SliceStable(wList, func(i, j int) bool {
		return wList[i].ProcessedAt.After(wList[j].ProcessedAt)
	})

	return wList, nil
}

func (s *Service) GetOrdersForProcessing(ctx context.Context) ([]string, error) {
	oList, err := s.Storage.GetOrdersList(ctx)
	if err != nil {
		return nil, err
	}

	return oList, nil
}

func (s *Service) FetchOrderInfo(
	ctx context.Context,
	client *resty.Client,
	orderNo string,
) (*accmodels.OrderInfoResponse, error) {
	var updatedInfo accmodels.OrderInfoResponse

	url := s.Config.Service.AccrualSystemAddress + s.Config.Service.AccrualOrderInfoRoute
	resp, err := client.R().
		SetContext(ctx).
		SetPathParam("orderID", orderNo).
		SetResult(&updatedInfo).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed make request to accrual: %w", err)
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("got unexpected request error: %s", resp.Error())
	}
	return &updatedInfo, nil
}

func (s *Service) UpdateOrder(ctx context.Context, data *accmodels.OrderInfoResponse) error {
	status, err := data.Status.ConvertToOrderStatus()
	if err != nil {
		return fmt.Errorf("failed convert to order status: %w", err)
	}

	updData := &orders.UpdateOrder{Status: status, Number: data.Order, Accrual: data.Accrual}

	if err = s.Storage.UpdateOrder(ctx, updData); err != nil {
		return fmt.Errorf("failed update order: '%v', details: %w", data.Order, err)
	}

	return nil
}
