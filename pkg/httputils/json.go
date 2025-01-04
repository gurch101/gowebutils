package httputils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ErrInvalidJSON is returned when the body is not valid JSON.
var ErrInvalidJSON = errors.New("invalid JSON")

// ReadJSON decodes request Body into corresponding Go type. It triages for any potential errors
// and returns corresponding appropriate errors.
func ReadJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Use http.MaxBytesReader() to limit the size of the request body to 1MB to prevent
	// any potential nefarious DoS attacks.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Initialize the json.Decoder, and call the DisallowUnknownFields() method on it
	// before decoding. So, if the JSON from the client includes any field which
	// cannot be mapped to the target destination, the decoder will return an error
	// instead of just ignoring the field.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the request body into the destination.
	if err := dec.Decode(dst); err != nil {
		return handleDecodeError(err, maxBytes)
	}

	if err := ensureSingleJSONValue(dec); err != nil {
		return err
	}

	return nil
}

// handleDecodeError handles errors returned by json.Decoder.Decode and returns custom errors.
func handleDecodeError(err error, maxBytes int) error {
	var syntaxError *json.SyntaxError

	var unmarshalTypeError *json.UnmarshalTypeError

	var invalidUnmarshalError *json.InvalidUnmarshalError

	switch {
	case errors.As(err, &syntaxError):
		return fmt.Errorf("%w: body contains badly-formed JSON at (character %d)", ErrInvalidJSON, syntaxError.Offset)

	case errors.Is(err, io.ErrUnexpectedEOF):
		return fmt.Errorf("%w: body contains badly-formed JSON", ErrInvalidJSON)

	case errors.As(err, &unmarshalTypeError):
		return handleUnmarshalTypeError(unmarshalTypeError)

	case errors.Is(err, io.EOF):
		return fmt.Errorf("%w: body must not be empty", ErrInvalidJSON)

	case strings.HasPrefix(err.Error(), "json: unknown field "):
		fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")

		return fmt.Errorf("%w: body contains unknown key %s", ErrInvalidJSON, fieldName)

	case err.Error() == "http: request body too large":
		return fmt.Errorf("%w: body must not be larger than %d bytes", ErrInvalidJSON, maxBytes)

	case errors.As(err, &invalidUnmarshalError):
		panic(err)

	default:
		return err
	}
}

// handleUnmarshalTypeError handles json.UnmarshalTypeError and returns a custom error.
func handleUnmarshalTypeError(err *json.UnmarshalTypeError) error {
	if err.Field != "" {
		return fmt.Errorf("%w: body contains incorrect JSON type for field %q", ErrInvalidJSON, err.Field)
	}

	return fmt.Errorf("%w: body contains incorrect JSON type (at character %d)", ErrInvalidJSON, err.Offset)
}

// ensureSingleJSONValue ensures the request body contains only a single JSON value.
func ensureSingleJSONValue(dec *json.Decoder) error {
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("%w: body must only contain a single JSON value", ErrInvalidJSON)
	}

	return nil
}

// WriteJSON marshals data structure to encoded JSON response and writes it to the response body.
func WriteJSON(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	// Use the json.MarshalIndent() function so that whitespace is added to the encoded JSON. Use
	// no line prefix and tab indents for each element.
	jsonPayload, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal indent in write json: %w", err)
	}

	// Append a newline to make it easier to view in terminal applications.
	jsonPayload = append(jsonPayload, '\n')

	// At this point, we know that we won't encounter any more errors before writing the response,
	// so it's safe to add any headers that we want to include. We loop through the header map
	// and add each header to the http.ResponseWriter header map. Note that it's OK if the
	// provided header map is nil. Go doesn't through an error if you try to range over (
	// or generally, read from) a nil map
	for key, value := range headers {
		w.Header()[key] = value
	}

	SetJSONContentTypeResponseHeader(w)
	w.WriteHeader(status)

	if _, err := w.Write(jsonPayload); err != nil {
		return fmt.Errorf("failed to write json: %w", err)
	}

	return nil
}
