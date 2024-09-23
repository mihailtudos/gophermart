package middleware

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/mihailtudos/gophermart/internal/domain"
	"github.com/mihailtudos/gophermart/internal/repository/postgres"
	"github.com/mihailtudos/gophermart/internal/service/auth"
	"github.com/mihailtudos/gophermart/pkg/helpers"
)

const (
	authTokenType      = "Bearer"
	authHeaderParts    = 2
	authTokenTypeIndex = 0
	authTokenIndex     = 1
)

type authService interface {
	VerifyToken(ctx context.Context, token string) (string, error)
	GetUserByID(ctx context.Context, ID string) (domain.User, error)
}

var bearerRegex = regexp.MustCompile(`^Bearer\s[\w-]*\.[\w-]*\.[\w-]*$`)

func Authenticated(us authService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// avoid any sort of chaching
			w.Header().Add("Vary", "Authorization")
			authHeader := r.Header.Get("Authorization")

			// Check if the Authorization header is present
			if authHeader == "" {
				http.Error(w, "Authorization header missing", http.StatusUnauthorized)
				return
			}

			// Ensure the header starts with "Bearer" and is followed by a valid JWT format using regex
			if !bearerRegex.MatchString(authHeader) {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			headerParts := strings.Split(authHeader, " ")
			if headerParts[authTokenTypeIndex] != authTokenType || len(headerParts) != authHeaderParts {
				http.Error(w, "missing auth header", http.StatusUnauthorized)
				return
			}

			token := headerParts[authTokenIndex]
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
