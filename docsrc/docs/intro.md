---
sidebar_position: 1
---

# Introduction

`gowebutils` is a collection of packages that can be used to build web applications.

## Features

### HTTP

- sensible defaults for http/2 server with graceful shutdown
- utilities for handling JSON requests/responses, query string, and url path parameter parsing
- routing with `chi`
- middleware for logging, panic recovery, cors, session management, rate limiting, and gzip

### Validation

- simple validator for checking request errors

### Logging

- structured logging using `slog`

### Security

- Support for oauth2 authentication code flow. Tested with AWS Cognito.
- invite user flow support
- session management using `scs`

### Email

- templated email support using `gomail` and an `SMTP` provider of your choice.

### File Storage

- download/upload/delete files from s3

### Payments

- TODO: add payment service

### Database

- helpers for simple CRUD
- dynamic query builder
- database migrations with `gomigrate`
- database seeding
- optimal sqlite connection pooling built in

### Testing

- helpers for testing http endpoints
- in-memory database for testing

## Installation

In your go project directory, run:

```sh
go get github.com/gurch101/gowebutils

mkdir tls
cd tls && go run /usr/local/go/src/crypto/tls/generate_cert.go --rsa-bits=2048 --host=localhost
```
