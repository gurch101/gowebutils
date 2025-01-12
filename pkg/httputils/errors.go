package httputils

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/validation"
)

func logError(r *http.Request, err error) {
	slog.ErrorContext(
		r.Context(),
		err.Error(),
		"request_method", r.Method,
		"request_url", r.URL.String(),
		"stack", debug.Stack(),
	)
}

func errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	// Write the response using the writeJSON() helper. If this happens to return an error
	// then log it, and fall back to sending the client an empty response with a 500 Internal
	// Server Error status code
	err := WriteJSON(w, status, map[string]any{"errors": message}, nil)
	if err != nil {
		logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)

		_, err = w.Write([]byte("internal server error"))
		if err != nil {
			logError(r, err)
		}
	}
}

// serverErrorResponse method is used when our application encounters an unexpected problem
// at runtime. it logs the detailed error message and returns a 500 Internal Server Error.
func ServerErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	logError(r, err)

	message := "the server encountered a problem and could not process your request"
	errorResponse(w, r, http.StatusInternalServerError, message)
}

// UnprocessableEntityResponse method is used to send a 422 Unprocessable Entity status code.
func UnprocessableEntityResponse(w http.ResponseWriter, r *http.Request, err error) {
	errorResponse(w, r, http.StatusUnprocessableEntity, err.Error())
}

// BadRequestResponse sends a JSON-formatted error message with 400 Bad Request status code.
func BadRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// FailedValidationResponse sends JSON-formatted error message to client with 400 Bad Request status code.
func FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors []validation.Error) {
	errorResponse(w, r, http.StatusBadRequest, errors)
}

// NotFoundResponse method is used to send a 404 Not Found status code.
func NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	errorResponse(w, r, http.StatusNotFound, message)
}

func EditConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	errorResponse(w, r, http.StatusConflict, message)
}

func RateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	errorResponse(w, r, http.StatusTooManyRequests, message)
}

func UnauthorizedResponse(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api") {
		message := "You must be authenticated to access this resource"
		errorResponse(w, r, http.StatusUnauthorized, message)
	} else {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}
}

func HandleErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr validation.Error

	switch {
	case errors.As(err, &validationErr):
		FailedValidationResponse(w, r, []validation.Error{validationErr})
	case errors.Is(err, dbutils.ErrRecordNotFound):
		NotFoundResponse(w, r)
	case errors.Is(err, dbutils.ErrEditConflict):
		EditConflictResponse(w, r)
	default:
		ServerErrorResponse(w, r, err)
	}
}
