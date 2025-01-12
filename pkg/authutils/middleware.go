package authutils

import (
	"context"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/gurch101/gowebutils/pkg/httputils"
)

type getUserExistsFn[T any] func(ctx context.Context, user T) bool

func GetSessionMiddleware[T any](
	sessionManager *scs.SessionManager,
	userExistsFn getUserExistsFn[T],
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := sessionManager.Get(r.Context(), "user").(T)
			if !ok {
				httputils.UnauthorizedResponse(w, r)

				return
			}

			exists := userExistsFn(r.Context(), user)
			if !exists {
				httputils.UnauthorizedResponse(w, r)

				return
			}

			r = ContextSetUser(r, user)

			next.ServeHTTP(w, r)
		})
	}
}
