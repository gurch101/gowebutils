package parser

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// ErrInvalidPathParam is returned when the path parameter is invalid.
var ErrInvalidPathParam = errors.New("invalid path param")

// Retrieve the "id" URL parameter from the current request context, then convert it to
// an integer and return it. If the operation isn't successful, return 0 and an error.
func ReadIDPathParam(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: id: %w", ErrInvalidPathParam, err)
	}

	if id < 0 {
		return 0, ErrInvalidPathParam
	}

	return id, nil
}
