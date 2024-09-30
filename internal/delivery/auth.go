package delivery

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/mihailtudos/gophermart/internal/domain"
	"github.com/mihailtudos/gophermart/internal/repository/postgres"
	"github.com/mihailtudos/gophermart/internal/validator"
	"github.com/mihailtudos/gophermart/pkg/helpers"
)

type Auth interface {
	SetSessionToken(ctx context.Context, userID string, tokens string) error
	GenerateUserTokens(ctx context.Context, userID string) (domain.Tokens, error)
	Login(ctx context.Context, input domain.UserAuthInput) (domain.User, error)
	Register(ctx context.Context, user domain.User) (string, error)
}

type authHandler struct {
	Auth
}

func NewAuthHanler(auth Auth) *authHandler {
	return &authHandler{auth}
}

func (ah *authHandler) Signin(w http.ResponseWriter, r *http.Request) {
	var input domain.UserAuthInput
	if err := helpers.ReadJSON(w, r, &input); err != nil {
		ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	user := domain.User{Login: input.Login}
	if err := user.Password.Set(input.Password); err != nil {
		ServerErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	domain.ValidateUser(v, &user)
	if !v.Valid() {
		ErrorResponse(w, r, http.StatusUnauthorized, "invalid credentials")
		return
	}
	user, err := ah.Login(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrNoRowsFound):
			ErrorResponse(w, r, http.StatusBadRequest, "incorrect credentials")
		default:
			ServerErrorResponse(w, r, err)
		}
		return
	}

	tokens, err := ah.GenerateUserTokens(r.Context(), user.ID)
	if err != nil {
		ServerErrorResponse(w, r,
			fmt.Errorf("failed to generate user tokens: %w", err))
		return
	}

	if err := ah.SetSessionToken(r.Context(), user.ID, tokens.RefreshToken); err != nil {
		ServerErrorResponse(w, r,
			fmt.Errorf("failed to set user session: %w", err))
		return
	}

	// Set the Authorization header
	helpers.SetAuthorizationHeaders(w, tokens)

	_, err = helpers.WriteJSON(w, http.StatusOK,
		helpers.Envelope{
			"access_token":  tokens.AccessToken,
			"refresh_token": tokens.RefreshToken,
		},
		nil)

	if err != nil {
		ServerErrorResponse(w, r,
			fmt.Errorf("failed to write JSON response: %w", err))
	}
}

func (ah *authHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var input domain.UserAuthInput
	if err := helpers.ReadJSON(w, r, &input); err != nil {
		ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	user := domain.User{Login: input.Login}
	if err := user.Password.Set(input.Password); err != nil {
		ServerErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	domain.ValidateUser(v, &user)
	if !v.Valid() {
		FailedValidationResponse(w, r, v.Errors)
		return
	}

	userID, err := ah.Register(r.Context(), user)
	if err != nil {
		if errors.Is(err, postgres.ErrDuplicateLogin) {
			ErrorResponse(w, r, http.StatusConflict, "login has been taken")
			return
		}

		ServerErrorResponse(w, r,
			fmt.Errorf("failed to register new user: %w", err))
		return
	}

	tokens, err := ah.GenerateUserTokens(r.Context(), userID)
	if err != nil {
		ServerErrorResponse(w, r,
			fmt.Errorf("failed to generate user tokens: %w", err))
		return
	}

	if err := ah.SetSessionToken(r.Context(), userID, tokens.RefreshToken); err != nil {
		ServerErrorResponse(w, r,
			fmt.Errorf("failed to set user session: %w", err))
		return
	}

	// Setting auth headers
	helpers.SetAuthorizationHeaders(w, tokens)

	// Respond with the tokens in the JSON body
	_, err = helpers.WriteJSON(w, http.StatusOK,
		helpers.Envelope{
			"access_token":  tokens.AccessToken,
			"refresh_token": tokens.RefreshToken,
		},
		nil)

	if err != nil {
		ServerErrorResponse(w, r,
			fmt.Errorf("failed to write JSON response: %w", err))
	}
}
