package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/RIBorisov/gophermart/internal/service"
	"github.com/RIBorisov/gophermart/internal/storage"
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
			if errors.Is(err, storage.ErrAnotherUserOrderCreated) {
				http.Error(w, storage.ErrAnotherUserOrderCreated.Error(), http.StatusConflict)
				return
			}
			if errors.Is(err, storage.ErrOrderCreatedAlready) {
				w.WriteHeader(http.StatusOK)
				return
			} else {
				svc.Log.Err("failed create order", err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
		}
		svc.Log.Info("Successfully loaded order", "order_id", string(orderNo))
		w.WriteHeader(http.StatusAccepted)
	}
}

func GetOrders(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		list, err := svc.GetUserOrders(ctx)
		if err != nil {
			svc.Log.Err("failed get orders", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		if len(list) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err = json.NewEncoder(w).Encode(list); err != nil {
			svc.Log.Err("failed encode response", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
