package testutils

import "net/http"

func StubAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Stub middleware that does nothing
		next.ServeHTTP(w, r)
	})
}
