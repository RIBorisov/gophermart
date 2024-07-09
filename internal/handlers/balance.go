package handlers

import (
	"encoding/json"
	"errors"
	"github.com/RIBorisov/gophermart/internal/errs"
	"github.com/RIBorisov/gophermart/internal/service"
	"net/http"
)

func CurrentBalance(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		current, err := svc.GetBalance(ctx)
		if err != nil {
			if errors.Is(err, errs.ErrUserNotExists) {
				http.Error(w, "Balance info not found", http.StatusNotFound)
				return
			}
			svc.Log.Err("failed get current balance", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(current); err != nil {
			svc.Log.Err("failed encode response", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// POST
func BalanceWithdraw(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//ctx := r.Context()
	}
}
