package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/mihailtudos/gophermart/internal/repository/postgres"
	"github.com/mihailtudos/gophermart/internal/service"
	"github.com/mihailtudos/gophermart/internal/service/auth"
	"github.com/mihailtudos/gophermart/pkg/helpers"
)

func Authenticated(us service.UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// avoid any sort of chaching
			w.Header().Add("Vary", "Authorization")
			authorizationHeader := r.Header.Get("Authorization")
			if authorizationHeader == "" {
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			headerParts := strings.Split(authorizationHeader, " ")
			if headerParts[0] != "Bearer" || len(headerParts) != 2 {
				http.Error(w, "missing auth header", http.StatusUnauthorized)
				return
			}

			token := headerParts[1]
			id, err := us.VerifyToken(r.Context(), token)
			if err != nil {
				switch {
				case errors.Is(err, auth.ErrTokenExpired):
					http.Error(w, "token expired", http.StatusUnauthorized)
					return
				case errors.Is(err, auth.ErrInvalidToken):
					http.Error(w, "invalid token", http.StatusUnauthorized)
					return
				default:
					http.Error(w, "we encountered an issue, try later", http.StatusInternalServerError)
					return
				}
			}

			user, err := us.GetUserByID(r.Context(), id)
			if err != nil {
				switch {
				case errors.Is(err, postgres.ErrNoRowsFound):
					http.Error(w, "token expired", http.StatusUnauthorized)
					return
				default:
					http.Error(w, "we encountered an issue, try later", http.StatusInternalServerError)
					return
				}
			}

			r = helpers.ContextSetUser(r, user)
			next.ServeHTTP(w, r)
		})
	}
}
