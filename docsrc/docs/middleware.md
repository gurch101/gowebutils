# Middleware

`gowebutils` automatically applies the following middleware to all routes:

- RealIP - sets the remote address of the request to the X-Real-IP/X-Forwarded-For header
- RequestID - injects a RequestID into the request context
- RateLimitMiddleware - limits the number of requests per second
- RequestLogger - logs the request id, request method, request path, request status, request duration, and request size
- Recoverer - logs and recovers from panics and returns a 500 status code
- Compress - compresses the response body based on the Accept-Encoding header
- sessionManager.LoadAndSave - loads and saves session data for the request

Requests added as protected routes will have the following additional middleware applied:

- SessionMiddleware - ensure a valid session is present for the request, otherwise returns a 401 status code
- NoCache - sets the Cache-Control header to no-cache, no-store, must-revalidate

Requests can also be added as protected routes with additional permissions required which would apply the following middleware:

- PermissionsMiddleware - ensure the user has the required permissions for the request, otherwise returns a 403 status code

There is also optional middleware that can be used to add to your routes via `AddProtectedRouteWithMiddleware`:

- authutils.IsAdmin - ensures the user is an admin
