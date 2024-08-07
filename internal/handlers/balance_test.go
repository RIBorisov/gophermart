package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/RIBorisov/gophermart/internal/config"
	"github.com/RIBorisov/gophermart/internal/logger"
	"github.com/RIBorisov/gophermart/internal/models/balance"
	"github.com/RIBorisov/gophermart/internal/service"
	"github.com/RIBorisov/gophermart/internal/storage"
	"github.com/RIBorisov/gophermart/internal/storage/mocks"
)

func TestCurrentBalance(t *testing.T) {
	const (
		route = "/api/user/balance"
		GET   = http.MethodGet
	)
	log := &logger.Log{}
	log.Initialize("DEBUG")
	cfg, err := config.LoadConfig()
	assert.NoError(t, err)

	tests := []struct {
		name           string
		callTimes      int
		wantStatusCode int
		wantResponse   *storage.BalanceEntity
		wantError      error
	}{
		{
			name:           "Positive #1",
			callTimes:      1,
			wantStatusCode: http.StatusOK,
			wantResponse: &storage.BalanceEntity{
				Current:   100.15,
				Withdrawn: 100.85,
			},
			wantError: nil,
		},
		{
			name:           "Negative #1",
			callTimes:      1,
			wantStatusCode: http.StatusNotFound,
			wantResponse:   &storage.BalanceEntity{},
			wantError:      storage.ErrUserNotExists,
		},
		{
			name:           "Negative #2",
			callTimes:      1,
			wantStatusCode: http.StatusInternalServerError,
			wantResponse:   &storage.BalanceEntity{},
			wantError:      errors.New("unexpected error"),
		},
	}
	for _, tt := range tests { //nolint:dupl // duplicate code block
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockStore(ctrl)
			mockStore.EXPECT().GetBalance(gomock.Any()).Times(tt.callTimes).Return(tt.wantResponse, tt.wantError)

			svc := &service.Service{Config: cfg, Log: log, Storage: mockStore}
			handler := CurrentBalance(svc)
			assert.NoError(t, err)

			req, err := http.NewRequest(GET, route, http.NoBody)
			assert.NoError(t, err)

			w := httptest.NewRecorder()

			handler(w, req)
			resp := w.Result()
			assert.NoError(t, resp.Body.Close())
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
		})
	}
}

func TestBalanceWithdraw(t *testing.T) {
	const (
		route = "/api/user/balance/withdraw"
		POST  = http.MethodPost
	)
	log := &logger.Log{}
	log.Initialize("DEBUG")
	cfg, err := config.LoadConfig()
	assert.NoError(t, err)

	tests := []struct {
		name           string
		callTimes      int
		wantStatusCode int
		body           balance.WithdrawRequest
		wantError      error
	}{
		{
			name:           "Positive #1",
			callTimes:      1,
			body:           balance.WithdrawRequest{Order: "3682151158", Sum: 15.19},
			wantStatusCode: http.StatusOK,
			wantError:      nil,
		},
		{
			name:           "Negative #1",
			callTimes:      1,
			body:           balance.WithdrawRequest{Order: "3682151158", Sum: 9999.115},
			wantStatusCode: http.StatusPaymentRequired,
			wantError:      storage.ErrInsufficientFunds,
		},
		{
			name:           "Negative #2",
			callTimes:      1,
			body:           balance.WithdrawRequest{Order: "3682151158", Sum: 9999.115},
			wantStatusCode: http.StatusInternalServerError,
			wantError:      storage.ErrGetUserFromContext,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockStore(ctrl)
			mockStore.EXPECT().BalanceWithdraw(gomock.Any(), tt.body).Times(tt.callTimes).Return(tt.wantError)

			svc := &service.Service{Config: cfg, Log: log, Storage: mockStore}
			handler := BalanceWithdraw(svc)
			assert.NoError(t, err)

			reqBody, err := json.Marshal(tt.body)
			assert.NoError(t, err)

			req, err := http.NewRequest(POST, route, bytes.NewBuffer(reqBody))
			assert.NoError(t, err)

			w := httptest.NewRecorder()

			handler(w, req)
			resp := w.Result()
			assert.NoError(t, resp.Body.Close())
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
		})
	}
}
