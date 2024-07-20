package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/RIBorisov/gophermart/internal/models/balance"
	"github.com/RIBorisov/gophermart/internal/service"
	"github.com/RIBorisov/gophermart/internal/storage"
)

func CurrentBalance(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		current, err := svc.GetBalance(ctx)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotExists) {
				http.Error(w, "Balance info not found", http.StatusNotFound)
				return
			}
			svc.Log.Err("failed get current balance", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err = json.NewEncoder(w).Encode(current); err != nil {
			svc.Log.Err("failed encode response", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}

func BalanceWithdraw(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req balance.WithdrawRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			svc.Log.Err("failed decode request into struct", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		defer func() {
			if err = r.Body.Close(); err != nil {
				svc.Log.Err("failed close request body", err)
				http.Error(w, "", http.StatusInternalServerError)
			}
		}()

		if err = service.ValidateLuhn(req.Order); err != nil {
			http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
			return
		}

		err = svc.BalanceWithdraw(ctx, req)
		if err != nil {
			if errors.Is(err, storage.ErrInsufficientFunds) {
				http.Error(w, "You have insufficient funds", http.StatusPaymentRequired)
				return
			}
			svc.Log.Err("failed make balance withdraw", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
