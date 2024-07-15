package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/RIBorisov/gophermart/internal/config"
	"github.com/RIBorisov/gophermart/internal/errs"
	"github.com/RIBorisov/gophermart/internal/logger"
	"github.com/RIBorisov/gophermart/internal/service"
	"github.com/RIBorisov/gophermart/internal/storage/mocks"
)

func TestLoginHandler(t *testing.T) {
	const route = "/api/user/login"
	log := &logger.Log{}
	log.Initialize("DEBUG")
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Err("failed load config", err)
	}

	tests := []struct {
		name           string
		method         string
		callTimes      int
		body           map[string]string
		wantStatusCode int
		wantError      error
	}{
		{
			name:      "Positive #1",
			method:    http.MethodPost,
			callTimes: 1,
			body: map[string]string{
				"login":    "Oleg",
				"password": "1kOp0x,^",
			},
			wantStatusCode: http.StatusOK,
			wantError:      nil,
		},
		{
			name:      "Negative #1",
			method:    http.MethodPost,
			callTimes: 0,
			body: map[string]string{
				"login":    "",
				"password": "111",
			},
			wantStatusCode: http.StatusBadRequest,
			wantError:      nil,
		},
		{
			name:      "Negative #2",
			method:    http.MethodPost,
			callTimes: 1,
			body: map[string]string{
				"login":    "Oleg",
				"password": "111",
			},
			wantStatusCode: http.StatusConflict,
			wantError:      errs.ErrUserExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockStore(ctrl)
			svc := &service.Service{Config: cfg, Log: log, Storage: mockStore}

			mockStore.EXPECT().
				SaveUser(gomock.Any(), gomock.Any()).
				Times(tt.callTimes).
				Return("", tt.wantError)

			handler := Register(svc)
			reqBody, err := json.Marshal(tt.body)
			assert.NoError(t, err)

			req, err := http.NewRequest(tt.method, route, bytes.NewBuffer(reqBody))
			assert.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler(w, req)
			resp := w.Result()
			assert.NoError(t, resp.Body.Close())
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
		})
	}
}
