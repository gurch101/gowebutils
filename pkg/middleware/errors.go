package middleware

import (
	"log/slog"
	"net/http"

	"gurch101.github.io/go-web/pkg/parser"
	"gurch101.github.io/go-web/pkg/validation"
)

func logError(r *http.Request, err error) {
	slog.Error(err.Error(), "request_method", r.Method, "request_url", r.URL.String())
}

func errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	// Write the response using the writeJSON() helper. If this happens to return an error
	// then log it, and fall back to sending the client an empty response with a 500 Internal
	// Server Error status code
	err := parser.WriteJSON(w, status, map[string]any{"errors": message}, nil)
	if err != nil {
		logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
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
	errorResponse(w, r, http.StatusBadRequest, err)
}

// failedValidationResponse sends JSON-formatted error message to client with 400 Bad Request status code.
func FailedValidationResponse(w http.ResponseWriter, r *http.Request, errors []validation.ValidationError) {
	errorResponse(w, r, http.StatusBadRequest, errors)
}
