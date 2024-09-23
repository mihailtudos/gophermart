package delivery

import (
	"fmt"
	"net/http"

	"github.com/mihailtudos/gophermart/internal/logger"
	"github.com/mihailtudos/gophermart/pkg/helpers"
)

func FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	ErrorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func ErrorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := helpers.Envelope{"error": message}
	_, err := helpers.WriteJSON(w, status, env, nil)
	if err != nil {
		logger.LogError(r.Context(), err, "failed to write json response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	logger.LogError(r.Context(), err, "ending request")

	message := "the server encountered a problem and could not process your request"
	ErrorResponse(w, r, http.StatusInternalServerError, message)
}

func NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	ErrorResponse(w, r, http.StatusInternalServerError, message)
}

func MethodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	ErrorResponse(w, r, http.StatusInternalServerError, message)
}
