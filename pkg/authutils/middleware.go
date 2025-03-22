package authutils

import (
	"context"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/httputils"
)

type getUserExistsFn func(ctx context.Context, db dbutils.DB, user User) bool

func GetSessionMiddleware(
	sessionManager *scs.SessionManager,
	userExistsFn getUserExistsFn,
	db dbutils.DB,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := sessionManager.Get(r.Context(), "user").(User)
			if !ok {
				httputils.UnauthorizedResponse(w, r)

				return
			}

			exists := userExistsFn(r.Context(), db, user)
			if !exists {
				httputils.UnauthorizedResponse(w, r)

				return
			}

			r = ContextSetUser(r, user)

			next.ServeHTTP(w, r)
		})
	}
}
