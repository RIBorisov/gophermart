package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/RIBorisov/gophermart/internal/config"
	"github.com/RIBorisov/gophermart/internal/logger"
	"github.com/RIBorisov/gophermart/internal/service"
	"github.com/RIBorisov/gophermart/internal/storage"
	"github.com/RIBorisov/gophermart/internal/storage/mocks"
)

func TestWithdrawals(t *testing.T) {
	const (
		route = "/api/user/withdrawals"
		GET   = http.MethodGet
	)
	log := &logger.Log{}
	log.Initialize("DEBUG")
	cfg, err := config.LoadConfig()
	assert.NoError(t, err)
	now := time.Now()
	tests := []struct {
		name           string
		callTimes      int
		wantStatusCode int
		wantResponse   []storage.WithdrawalsEntity
		wantError      error
	}{
		{
			name:           "Positive #1",
			callTimes:      1,
			wantStatusCode: http.StatusOK,
			wantResponse: []storage.WithdrawalsEntity{
				{
					ProcessedAt: now,
					UserID:      "123",
					OrderID:     "5116141762",
					Amount:      150.99,
				},
				{
					ProcessedAt: now,
					UserID:      "123",
					OrderID:     "5830317037",
					Amount:      15.75,
				},
			},
			wantError: nil,
		},
		{
			name:           "Negative #1",
			callTimes:      1,
			wantStatusCode: http.StatusNoContent,
			wantResponse:   nil,
			wantError:      service.ErrNoWithdrawals,
		},
		{
			name:           "Negative #2",
			callTimes:      1,
			wantStatusCode: http.StatusInternalServerError,
			wantResponse:   nil,
			wantError:      errors.New("unexpected error"),
		},
	}
	for _, tt := range tests { //nolint:dupl // duplicate code block
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockStore(ctrl)
			mockStore.EXPECT().GetWithdrawals(gomock.Any()).Times(tt.callTimes).Return(tt.wantResponse, tt.wantError)

			svc := &service.Service{Config: cfg, Log: log, Storage: mockStore}
			handler := Withdrawals(svc)
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
