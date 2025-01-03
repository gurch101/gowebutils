# Go Web Application Skeleton

### Features

##### HTTP

- structured logging
- middleware for logging, panic recovery, and rate limiting
- error response handling
- sensible defaults for http server with graceful shutdown
- utilities for handling JSON requests/responses, query string and url path parameter parsing

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
2. run `make dev/run`

### Production

shutdown the server by running `pkill -SIGTERM webapp`
