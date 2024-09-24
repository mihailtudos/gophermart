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
	"github.com/mihailtudos/gophermart/internal/validator"
	"github.com/mihailtudos/gophermart/pkg/helpers"
)

const (
	ContentTypeHeaderName = "Content-Type"
	PlainTextContentType  = "text/plain"
)

type userHandler struct {
	UserManager
}

func NewUserHandler(um UserManager) *chi.Mux {
	uh := userHandler{um}

	router := chi.NewMux()

	// user protected routes
	router.Group(func(r chi.Router) {
		r.Use(middleware.Authenticated(um))
		r.Post("/orders", uh.registerOrder)
		r.Get("/orders", uh.getOrders)
		r.Get("/balance", uh.getBalance)
		r.Post("/balance/withdraw", uh.withrawalPoints)
		r.Get("/withdrawals", uh.getWithrawals)
	})

	return router
}

func (uh *userHandler) registerOrder(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(ContentTypeHeaderName) != PlainTextContentType {
		ErrorResponse(w, r, http.StatusBadRequest, nil)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}
	input := string(body)

	v := validator.New()
	v.Check(validator.IsValidOrderNumber(input), "order number", "invalid order number format")
	if !v.Valid() {
		FailedValidationResponse(w, r, v.Errors)
		return
	}

	user := helpers.ContextGetUser(r)
	order := domain.Order{
		OrderNumber: input,
		UserID:      user.ID,
		OrderStatus: domain.OrderStatusNew,
	}

	_, err = uh.RegisterOrder(r.Context(), order)

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

	w.WriteHeader(http.StatusAccepted)
}

func (uh *userHandler) getOrders(w http.ResponseWriter, r *http.Request) {
	user := helpers.ContextGetUser(r)

	orders, err := uh.GetUserOrders(r.Context(), user.ID)
	if err != nil {
		ServerErrorResponse(w, r, err)
		return
	}

	fmt.Printf("%#v", orders)
	fmt.Println("order cOunt", len(orders))

	if len(orders) == 0 {
		_, err := helpers.WriteJSON(w, http.StatusNoContent, nil, nil)
		if err != nil {
			ServerErrorResponse(w, r, err)
		}

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
	if _, err := w.Write(js); err != nil {
		ServerErrorResponse(w, r, err)
		return
	}
}

func (uh *userHandler) getBalance(w http.ResponseWriter, r *http.Request) {
	user := helpers.ContextGetUser(r)

	balance, err := uh.GetUserBalance(r.Context(), user.ID)
	if err != nil {
		ServerErrorResponse(w, r, err)
		return
	}

	if err := helpers.WriteUnwrappedJSON(w, http.StatusOK, balance, nil); err != nil {
		ServerErrorResponse(w, r, err)
		return
	}
}

func (uh *userHandler) withrawalPoints(w http.ResponseWriter, r *http.Request) {
	user := helpers.ContextGetUser(r)
	var withdrawalsRequest domain.Withdrawal
	if err := helpers.ReadJSON(w, r, &withdrawalsRequest); err != nil {
		ServerErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(validator.IsValidOrderNumber(withdrawalsRequest.Order), "order", "invalid order number")

	if !v.Valid() {
		FailedValidationResponse(w, r, v.Errors)
		return
	}
	withdrawalsRequest.UserID = user.ID

	_, err := uh.WithdrawalPoints(r.Context(), withdrawalsRequest)
	if err != nil {
		if errors.Is(err, postgres.ErrInsufficientPoints) {
			ErrorResponse(w, r, http.StatusPaymentRequired, "insufficient points")
			return
		}

		ServerErrorResponse(w, r, err)
		return
	}
}

func (uh *userHandler) getWithrawals(w http.ResponseWriter, r *http.Request) {
	user := helpers.ContextGetUser(r)
	withdrawals, err := uh.GetWithdrawals(r.Context(), user.ID)
	if err != nil {
		ServerErrorResponse(w, r, err)
		return
	}

	if len(withdrawals) == 0 {
		ErrorResponse(w, r, http.StatusNoContent, "no withdrawal records")
		return
	}

	if err := helpers.WriteUnwrappedJSON(w, http.StatusOK, withdrawals, nil); err != nil {
		ServerErrorResponse(w, r, err)
		return
	}
}
