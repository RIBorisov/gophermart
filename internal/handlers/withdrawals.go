package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/RIBorisov/gophermart/internal/service"
)

func Withdrawals(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wList, err := svc.GetWithdrawals(r.Context())
		if err != nil {
			if errors.Is(err, service.ErrNoWithdrawals) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			svc.Log.Err("failed get withdrawals list", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(wList); err != nil {
			svc.Log.Err("failed encode withdrawals response", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
