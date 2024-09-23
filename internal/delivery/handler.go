package delivery

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mihailtudos/gophermart/internal/service"
)

type Handler struct {
	services *service.Services
}

func NewHandler(s *service.Services) *chi.Mux {
	h := &Handler{
		services: s,
	}

	router := chi.NewMux()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "HEAD", "OPTION"},
		AllowedHeaders:   []string{"User-Agent", "Content-Type", "Accept", "Accept-Encoding", "Accept-Language", "Cache-Control", "Connection", "DNT", "Host", "Origin", "Pragma", "Referer"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	userHandler := NewUserHandler(*h.services.UserService)
	authHandler := NewAuthHanler(h.services.UserService)

	router.Route("/api/user", func(r chi.Router) {
		r.Mount("/", userHandler)                 // Mount the general user routes here
		r.Post("/login", authHandler.Login)       // Add specific login route
		r.Post("/register", authHandler.Register) // Add specific register route
	})

	return router
}
