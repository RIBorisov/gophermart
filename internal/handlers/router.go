package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	myMW "github.com/RIBorisov/gophermart/internal/middleware"
	"github.com/RIBorisov/gophermart/internal/service"
)

func NewRouter(svc *service.Service) *chi.Mux {

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/api/user/register", Register(svc))
	router.Post("/api/user/login", Login(svc))
	router.Route("/api/user", func(r chi.Router) {
		r.Use(myMW.CheckAuth(svc).Middleware)
		r.Post("/orders", CreateOrder(svc))
		r.Get("/orders", GetOrders(svc))
		r.Get("/balance", CurrentBalance(svc))
		r.Post("/balance/withdraw", BalanceWithdraw(svc))
		//GET /api/user/withdrawals
	})

	//router.Get("/", DraftHandler(svc.Log))
	return router
}
