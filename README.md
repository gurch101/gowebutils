# Go Web Application Skeleton

### Features

##### HTTP

- structured logging
- middleware for logging, panic recovery, cors, session management, rate limiting, and gzip
- error response handling
- sensible defaults for http server with graceful shutdown
- utilities for handling JSON requests/responses, query string and url path parameter parsing
- https and http/2 out-of-the-box

##### Security

- Support for oauth2 authentication code flow. Tested with AWS Cognito.
- invite user flow support
- TODO: session and bearer token authentication
- TODO: Role-based access control

##### Email

- templated email sending

##### File Storage

- TODO: add file storage service

##### Payments

- TODO: add payment service

##### Database

- helpers for simple CRUD
- dynamic query builder
- database migrations
- database seeding

##### Validation

- simple validator for checking request errors

##### Testing

- helpers for testing http endpoints
- in-memory database for testing

### Installation

1. run `./setup.sh`
2. Create a cognito user pool and register an app. Set env vars in .env
3. run `make dev/run`

### Production

shutdown the server by running `pkill -SIGTERM webapp`
