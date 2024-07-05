package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/RIBorisov/gophermart/internal/errs"
	"github.com/RIBorisov/gophermart/internal/service"
)

func CreateOrder(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		orderNo, err := io.ReadAll(r.Body)
		if err != nil {
			svc.Log.Err("failed read request body", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		if len(orderNo) == 0 {
			http.Error(w, "Empty request body, please provide order number", http.StatusBadRequest)
			return
		}

		if err = service.ValidateLuhn(string(orderNo)); err != nil {
			http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
			return
		}
		if err = svc.CreateOrder(ctx, string(orderNo)); err != nil {
			if errors.Is(err, errs.ErrAnotherUserOrderCreated) {
				http.Error(w, errs.ErrAnotherUserOrderCreated.Error(), http.StatusConflict)
				return
			}
			if errors.Is(err, errs.ErrOrderCreatedAlready) {
				w.WriteHeader(http.StatusOK)
				return
			} else {
				svc.Log.Err("failed create order", err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
		}
		svc.Log.Info("Successfully loaded order", "number", string(orderNo))
		w.WriteHeader(http.StatusAccepted)
	}
}
