package httputils

import "net/http"

type Middleware func(http.Handler) http.Handler

// Router is a custom router that supports global and route-specific middleware.
type Router struct {
	mux                      *http.ServeMux
	authenticationMiddleware []Middleware // Middleware applied to authenticated routes
	globalMiddlewares        []Middleware // Middleware applied to all routes
}

// NewRouter creates a new Router instance.
func NewRouter(authenticationMiddleware ...Middleware) *Router {
	return &Router{
		mux:                      http.NewServeMux(),
		authenticationMiddleware: authenticationMiddleware,
		globalMiddlewares:        []Middleware{},
	}
}

// Use adds global middleware that will be applied to all routes.
func (r *Router) Use(middlewares ...Middleware) {
	r.globalMiddlewares = append(r.globalMiddlewares, middlewares...)
}

func (r *Router) AddStaticRoute(path string, handler http.Handler) {
	r.mux.Handle(path, handler)
}

// AddRoute adds a route with optional route-specific middleware.
func (r *Router) AddRoute(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	// Chain the middleware with the handler
	wrappedHandler := chainMiddleware(handler, middlewares...)
	r.mux.Handle(path, wrappedHandler)
}

func (r *Router) AddAuthenticatedRoute(path string, handler http.HandlerFunc, middlewares ...Middleware) {
	// Combine global middleware with route-specific middleware
	allMiddlewares := append([]Middleware{}, r.authenticationMiddleware...)
	allMiddlewares = append(allMiddlewares, middlewares...)
	// Chain the middleware with the handler
	wrappedHandler := chainMiddleware(handler, allMiddlewares...)
	r.mux.Handle(path, wrappedHandler)
}

// chainMiddleware chains multiple middleware functions together.
func chainMiddleware(handler http.Handler, middlewares ...Middleware) http.Handler {
	// Start with the final handler and wrap it with each middleware in reverse order
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}

func (r *Router) BuildHandler() http.Handler {
	return chainMiddleware(r.mux, r.globalMiddlewares...)
}
