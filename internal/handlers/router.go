package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/RIBorisov/gophermart/internal/service"
)

func NewRouter(svc *service.Service) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Post("/api/user/register", Register(svc))
	router.Post("/api/user/login", Login(svc))
	router.Route("/api/user", func(r chi.Router) {
		//r.Use(мидлварь проверки авторизации)
		//POST / api / user / orders
		//GET / api / user / orders
		//GET / api / user / balance
		//POST / api / user / balance / withdraw
		//GET /api/user/withdrawals
	})

	//router.Get("/", DraftHandler(svc.Log))
	return router
}
