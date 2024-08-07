package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/RIBorisov/gophermart/internal/config"
	"github.com/RIBorisov/gophermart/internal/logger"
	"github.com/RIBorisov/gophermart/internal/service"
	"github.com/RIBorisov/gophermart/internal/storage"
	"github.com/RIBorisov/gophermart/internal/storage/mocks"
)

func TestCreateOrder(t *testing.T) {
	const (
		route = "/api/user/orders"
		POST  = http.MethodPost
	)
	log := &logger.Log{}
	log.Initialize("DEBUG")
	cfg, err := config.LoadConfig()
	assert.NoError(t, err)

	tests := []struct {
		name           string
		orderNo        string
		callTimes      int
		wantStatusCode int
		wantError      error
	}{
		{
			name:           "Positive #1",
			orderNo:        "7177570715",
			callTimes:      1,
			wantStatusCode: http.StatusAccepted,
			wantError:      nil,
		},
		{
			name:           "Negative #1 (Not Luhn)",
			orderNo:        "0123456789",
			callTimes:      0,
			wantStatusCode: http.StatusUnprocessableEntity,
			wantError:      nil,
		},
		{
			name:           "Negative #2",
			orderNo:        "",
			callTimes:      0,
			wantStatusCode: http.StatusBadRequest,
			wantError:      nil,
		},
		{
			name:           "Negative #3",
			orderNo:        "7177570715",
			callTimes:      1,
			wantStatusCode: http.StatusConflict,
			wantError:      storage.ErrAnotherUserOrderCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockStore(ctrl)
			mockStore.EXPECT().SaveOrder(gomock.Any(), tt.orderNo).Times(tt.callTimes).Return(tt.wantError)

			svc := &service.Service{Config: cfg, Log: log, Storage: mockStore}

			handler := CreateOrder(svc)
			assert.NoError(t, err)

			req, err := http.NewRequest(POST, route, strings.NewReader(tt.orderNo))
			assert.NoError(t, err)

			w := httptest.NewRecorder()

			handler(w, req)
			resp := w.Result()
			assert.NoError(t, resp.Body.Close())
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
		})
	}
}

func TestGetOrders(t *testing.T) {
	const (
		route = "/api/user/orders"
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
		wantResponse   interface{}
		wantError      error
	}{
		{
			name:           "Positive #1",
			callTimes:      1,
			wantStatusCode: http.StatusOK,
			wantResponse: []storage.OrderEntity{
				{Status: "PROCESSED", OrderID: "1761025707", UserID: "123", Bonus: 150},
				{Status: "PROCESSED", OrderID: "4657676856", UserID: "123", Bonus: 250},
				{Status: "PROCESSED", OrderID: "2075656310", UserID: "123", Bonus: 115.55}},
			wantError: nil,
		},
		{
			name:           "Positive #2",
			callTimes:      1,
			wantStatusCode: http.StatusNoContent,
			wantResponse:   []storage.OrderEntity{},
			wantError:      nil,
		},
		{
			name:           "Negative #1",
			callTimes:      1,
			wantStatusCode: http.StatusInternalServerError,
			wantResponse: []storage.OrderEntity{
				{Status: "PROCESSED", OrderID: "1761025707", UserID: "123", Bonus: 150},
			},
			wantError: errors.New("unexpected error"),
		},
	}
	for _, tt := range tests { //nolint:dupl // duplicate code block
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockStore(ctrl)
			mockStore.EXPECT().
				GetUserOrders(gomock.Any()).Times(tt.callTimes).
				Return(tt.wantResponse, tt.wantError)

			svc := &service.Service{Config: cfg, Log: log, Storage: mockStore}
			handler := GetOrders(svc)
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
