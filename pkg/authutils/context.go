package authutils

import (
	"context"
	"net/http"
)

type contextKey string

const userContextKey = contextKey("user")

// The ContextSetUser() method returns a new copy of the request with the provided
// User struct added to the context. Note that we use our userContextKey constant as the
// key.
func ContextSetUser[T any](r *http.Request, user T) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)

	return r.WithContext(ctx)
}

// The ContextGetUser() retrieves the User struct from the request context. The only
// time that we'll use this helper is when we logically expect there to be User struct
// value in the context, and if it doesn't exist it will firmly be an 'unexpected' error.
// As we discussed earlier in the book, it's OK to panic in those circumstances.
func ContextGetUser[T any](r *http.Request) T {
	user, ok := r.Context().Value(userContextKey).(T)
	if !ok {
		panic("missing or invalid user value in request context")
	}

	return user
}
