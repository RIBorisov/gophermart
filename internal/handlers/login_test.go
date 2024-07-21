package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/RIBorisov/gophermart/internal/config"
	"github.com/RIBorisov/gophermart/internal/logger"
	"github.com/RIBorisov/gophermart/internal/models/register"
	"github.com/RIBorisov/gophermart/internal/service"
	"github.com/RIBorisov/gophermart/internal/storage"
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
		callGetTimes   int
		callSaveTimes  int
		body           *register.Request
		wantStatusCode int
		wantError      error
		wantResponse   interface{}
	}{
		{
			name:           "Positive #1",
			method:         http.MethodPost,
			callGetTimes:   1,
			callSaveTimes:  1,
			body:           &register.Request{Login: "Vasiliy", Password: "pwd"},
			wantStatusCode: http.StatusOK,
			wantError:      nil,
			wantResponse: &storage.UserRow{
				ID:       "123",
				Login:    "Vasiliy",
				Password: "$2a$10$W5SAQxshIk4miQCHExdmgOwW6bPpWRhXhKTu7qHnJZ0Ye./Qt7u42",
			},
		},
		{
			name:           "Negative #1",
			method:         http.MethodPost,
			callGetTimes:   0,
			callSaveTimes:  0,
			body:           &register.Request{Login: "", Password: "password1"},
			wantStatusCode: http.StatusBadRequest,
			wantError:      nil,
		},
		{
			name:           "Negative #2",
			method:         http.MethodPost,
			callGetTimes:   1,
			callSaveTimes:  0,
			body:           &register.Request{Login: "UserNotRegistered", Password: "some-password1234"},
			wantStatusCode: http.StatusUnauthorized,
			wantError:      storage.ErrUserNotExists,
		},
		{
			name:           "Negative #3",
			method:         http.MethodPost,
			callGetTimes:   1,
			callSaveTimes:  0,
			body:           &register.Request{Login: "UserNotRegistered", Password: "some-password1234"},
			wantStatusCode: http.StatusInternalServerError,
			wantError:      errors.New("unexpected error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockStore(ctrl)
			mockStore.EXPECT().GetUser(gomock.Any(), gomock.Any()).
				Times(tt.callGetTimes).
				Return(tt.wantResponse, tt.wantError)

			mockStore.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Times(tt.callSaveTimes).Return("123", nil)

			svc := &service.Service{Config: cfg, Log: log, Storage: mockStore}

			if tt.callSaveTimes > 0 {
				_, err = mockStore.SaveUser(context.Background(), tt.body)
				assert.NoError(t, err)
			}

			handler := Login(svc)
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
