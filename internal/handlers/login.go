package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/RIBorisov/gophermart/internal/models/login"
	"github.com/RIBorisov/gophermart/internal/models/register"
	"github.com/RIBorisov/gophermart/internal/service"
	"github.com/RIBorisov/gophermart/internal/storage"
)

func Login(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		response := login.Response{
			Success: true,
			Details: "Successfully logged in",
		}
		var user *register.Request

		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			svc.Log.Err("failed decode register request", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		if err := r.Body.Close(); err != nil {
			svc.Log.Err("failed close request body", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		if err := user.Validate(); err != nil {
			http.Error(w, "Please, check if login and password provided", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		authToken, err := svc.LoginUser(ctx, user)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotExists) || errors.Is(err, service.ErrIncorrectPassword) {
				http.Error(w, "Invalid login and (or) password", http.StatusUnauthorized)
				return
			} else {
				svc.Log.Err("failed login user", err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Authorization", "Bearer "+authToken)
		w.WriteHeader(http.StatusOK)

		if err = json.NewEncoder(w).Encode(response); err != nil {
			svc.Log.Err("failed encode response", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}
