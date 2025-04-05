# Routing

`gowebutils` uses the `chi` router to handle routing. The `App` object has two methods for adding routes:

`AddProtectedRoute(method string, path string, handler http.HandlerFunc)` - adds a route that requires a valid session to access

`AddPublicRoute(method string, path string, handler http.HandlerFunc)` - adds a route that does not require a valid session to access

See middleware for more information on the middlware applied to each route.
