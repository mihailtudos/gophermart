package delivery

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mihailtudos/gophermart/internal/delivery/middleware"
	"github.com/mihailtudos/gophermart/internal/domain"
	"github.com/mihailtudos/gophermart/internal/repository/postgres"
	"github.com/mihailtudos/gophermart/internal/service"
	"github.com/mihailtudos/gophermart/internal/service/auth"
	"github.com/mihailtudos/gophermart/internal/validator"
	"github.com/mihailtudos/gophermart/pkg/helpers"
)

type userHandler struct {
	service service.UserService
}

func NewUserHandler(us service.UserService) *chi.Mux {
	uh := userHandler{service: us}

	router := chi.NewMux()
	router.Post("/register", uh.register)
	router.Post("/login", uh.login)

	router.Group(func(r chi.Router) {
		r.Use(middleware.Authenticated(us))
		r.Post("/orders", uh.registerOrder)
		r.Get("/orders", uh.getOrders)
	})

	return router
}

func (uh *userHandler) login(w http.ResponseWriter, r *http.Request) {
	var input auth.UserAuthInput
	if err := helpers.ReadJSON(w, r, &input); err != nil {
		ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	user := domain.User{Login: input.Login}
	user.Password.Set(input.Password)

	v := validator.New()
	domain.ValidateUser(v, &user)
	if !v.Valid() {
		ErrorResponse(w, r, http.StatusUnauthorized, "invalid credentials")
		return
	}

	user, err := uh.service.Login(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrNoRowsFound):
			ErrorResponse(w, r, http.StatusBadRequest, "incorrect credentials")
		default:
			ServerErrorResponse(w, r, err)
		}
		return
	}

	tokens, err := uh.service.GenerateUserTokens(r.Context(), user.ID)
	if err != nil {
		ServerErrorResponse(w, r,
			fmt.Errorf("failed to generate user tokens: %w", err))
		return
	}

	if err := uh.service.SetSessionToken(r.Context(), user.ID, tokens.RefreshToken); err != nil {
		ServerErrorResponse(w, r,
			fmt.Errorf("failed to set user session: %w", err))
		return
	}

	// Set the Authorization header
	helpers.SetAuthorizationHeaders(w, tokens)

	err = helpers.WriteJSON(w, http.StatusOK,
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

func (uh *userHandler) register(w http.ResponseWriter, r *http.Request) {
	var input auth.UserAuthInput
	if err := helpers.ReadJSON(w, r, &input); err != nil {
		ErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	user := domain.User{Login: input.Login}
	user.Password.Set(input.Password)

	v := validator.New()
	domain.ValidateUser(v, &user)
	if !v.Valid() {
		FailedValidationResponse(w, r, v.Errors)
		return
	}

	userID, err := uh.service.Register(r.Context(), user)
	if err != nil {
		if errors.Is(err, postgres.ErrDuplicateLogin) {
			ErrorResponse(w, r, http.StatusConflict, "login has been taken")
			return
		}

		ServerErrorResponse(w, r,
			fmt.Errorf("failed to register new user: %w", err))
		return
	}

	tokens, err := uh.service.GenerateUserTokens(r.Context(), userID)
	if err != nil {
		ServerErrorResponse(w, r,
			fmt.Errorf("failed to generate user tokens: %w", err))
		return
	}

	if err := uh.service.SetSessionToken(r.Context(), userID, tokens.RefreshToken); err != nil {
		ServerErrorResponse(w, r,
			fmt.Errorf("failed to set user session: %w", err))
		return
	}

	// Setting auth headers
	helpers.SetAuthorizationHeaders(w, tokens)

	// Respond with the tokens in the JSON body
	err = helpers.WriteJSON(w, http.StatusOK,
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

func (uh *userHandler) registerOrder(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		ErrorResponse(w, r, http.StatusBadRequest, nil)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	input := string(body)

	v := validator.New()
	v.Check(validator.IsValidOrderNumber(input), "order number", "invalid order number format")
	if !v.Valid() {
		FailedValidationResponse(w, r, v.Errors)
		return
	}

	user := helpers.ContextGetUser(r)
	_, err = uh.service.RegisterOrder(r.Context(), input, user.ID)

	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrOrderAlreadyExistsSameUser):
			w.WriteHeader(http.StatusOK)
			return
		case errors.Is(err, postgres.ErrOrderAlreadyAccepted):
			w.WriteHeader(http.StatusAccepted)
			return
		case errors.Is(err, postgres.ErrOrderAlreadyExistsDifferentUser):
			w.WriteHeader(http.StatusConflict)
			return
		default:
			ServerErrorResponse(w, r, err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (uh *userHandler) getOrders(w http.ResponseWriter, r *http.Request) {
	user := helpers.ContextGetUser(r)

	orders, err := uh.service.GetUserOrders(r.Context(), user.ID)
	if err != nil {
		ServerErrorResponse(w, r, err)
		return
	}

	fmt.Printf("%#v", orders)
	fmt.Println("order cOunt", len(orders))

	if len(orders) == 0 {
		helpers.WriteJSON(w, http.StatusNoContent, nil, nil)
		return
	}

	js, err := json.MarshalIndent(orders, "", "\t")
	if err != nil {
		ServerErrorResponse(w, r, err)
		return
	}
	js = append(js, '\n')
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(js)
}
