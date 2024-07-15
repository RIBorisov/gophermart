package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/RIBorisov/gophermart/internal/models/register"
	"github.com/RIBorisov/gophermart/internal/service"
	"github.com/RIBorisov/gophermart/internal/storage"
)

func Register(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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

		authToken, err := svc.RegisterUser(ctx, user)

		if err != nil {
			if errors.Is(err, storage.ErrUserExists) {
				http.Error(w, "User already exists", http.StatusConflict)
				return
			} else {
				svc.Log.Err("failed register user", err)
				http.Error(w, "", http.StatusInternalServerError)
			}
		}

		w.Header().Set("Authorization", "Bearer "+authToken)
		w.WriteHeader(http.StatusOK)

		response := register.Response{
			Success: true,
			Details: fmt.Sprintf("Successfully registered user with name '%s'", user.Login),
		}

		if err = json.NewEncoder(w).Encode(response); err != nil {
			svc.Log.Err("failed encode response", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}
