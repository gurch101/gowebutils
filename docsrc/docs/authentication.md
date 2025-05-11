# Authentication

The `authutils` package provides comprehensive user authentication and registration functionality using OpenID Connect (OIDC) with AWS Cognito. This integration simplifies secure user management in your web application.

### Endpoints

- `/login`: Initiates the OIDC authentication flow.
- `/logout`: Initiates the OIDC logout flow.
- `/register`: Initiates the user registration flow.
- `/auth/callback`: Handles OIDC authentication responses (redirects to / on success).

### Initialization

1. Set up the following environment variables:

```sh
# Your application's host URL
export HOST="https://localhost:8080"
export OIDC_CLIENT_ID="your-client-id"
export OIDC_CLIENT_SECRET="your-client-secret"
# The base path for the /.well-known/openid-configuration endpoint
export OIDC_DISCOVERY_URL="https://cognito-idp.us-east-1.amazonaws.com/us-east-1_123456789"
# The IDP user signup URL
export REGISTRATION_URL="https://us-east-123456789.auth.us-east-1.amazoncognito.com/signup"
# The URL to redirect to after a successful login/registration. Defaults to /.
# For local dev, you can override this to redirect to your single-page app.
export REDIRECT_URL=/
# The IDP user logout URL
export LOGOUT_URL="https://us-east-123456789.auth.us-east-1.amazoncognito.com/logout"
# The URL to redirect to after the user has logged out
export POST_LOGOUT_REDIRECT_URL="${HOST}/static/bye.html"
```

2. Implement the following function which will be called whenever a user logs in or registers for your application:

```go
// The user returned by this function will be stored in the session store
func GetOrCreateUser(ctx context.Context, db dbutils.DB, email string, tokenPayload map[string]any) (authutils.User, error) {
  // return the valid user by email if it exists
  // otherwise, create a new user with the given email (brand new registration)
  // otherwise, create a new user with the given email + tokenPayload (invited user)
}
```

3. Implement the following function which will be called on every request to a protected route with a valid session:

```go
// Verifies that a user in the session is still valid
func GetUserExists(ctx context.Context, db dbutils.DB, user authutils.User) bool {
  // return true if the user is valid
}
```

4. When initializing your application, provide these functions:

```go
app, err := app.NewApp(
  app.WithGetOrCreateUserFn(GetOrCreateUser),
  app.WithGetUserExistsFn(GetUserExists),
)
```

### Invite User Flow

You can implement a user invitation flow with the following function. Invite tokens remain valid for 7 days.

```go
func InviteUser(
  _ context.Context,
  mailer mailutils.Mailer,
  hostName string,
  tenantID int64,
  username, email string,
) error {
  payload := map[string]any{
    "tenant_id": tenantID,
    "email":     email,
    "username":  username,
  }
  inviteToken, err := authutils.CreateInviteToken(payload)

  if err != nil {
    return err
  }

  mailer.Send(email, "invite.go.tmpl", map[string]string{
    "URL": fmt.Sprintf("%s/register?code=%s", hostName, inviteToken),
  })

  return nil
}
```
