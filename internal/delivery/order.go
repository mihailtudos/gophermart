package delivery

import (
	"github.com/go-chi/chi/v5"
	"github.com/mihailtudos/gophermart/internal/service"
)

type orderHandler struct {
	service service.OrderService
}

func NewOrderHandler(os service.OrderService) *chi.Mux {
	_ = orderHandler{service: os}

	router := chi.NewMux()

	return router
}
