package authutils

import (
	"context"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

type getUserExistsFn[T any] func(ctx context.Context, user T) (bool, error)

func GetSessionMiddleware[T any](
	sessionManager *scs.SessionManager,
	userExistsFn getUserExistsFn[T],
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := sessionManager.Get(r.Context(), "user").(T)
			if !ok {
				// TODO unauthenticated error
				return
			}

			exists, err := userExistsFn(r.Context(), user)
			if err != nil || !exists {
				// TODO unauthenticated error
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
