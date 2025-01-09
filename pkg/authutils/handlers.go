package authutils

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"golang.org/x/oauth2"
)

// ErrAuthCodeNotFound is returned when the authorization code is not found in the request.
var ErrAuthCodeNotFound = errors.New("authorization code not found")

// ErrInvalidState is returned when the state is invalid.
var ErrInvalidState = errors.New("invalid state")

// ErrTokenExchangeFailed is returned when the token exchange fails.
var ErrTokenExchangeFailed = errors.New("token exchange failed")

// ErrNoCode is returned when the code is not found in the request.
var ErrNoCode = errors.New("no code")

// ErrNoIDToken is returned when the id token is not found in the response.
var ErrNoIDToken = errors.New("no id token")

type OidcController[T any] struct {
	oauth2Config      *Oauth2Config
	getOrCreateUserFn GetOrCreateUser[T]
	sessionManager    *scs.SessionManager
}

type Oauth2Config struct {
	verifier        *oidc.IDTokenVerifier
	registrationURL string
	logoutURL       string
	postLogoutURL   string
	config          *oauth2.Config
}

// CreateOauthConfig creates an oauth2.Config object for the given idp URL.
// discoveryURL is the base URL that exposes /.well-known/openid-configuration
// spURL should be the host URL of your app.
func CreateOauthConfig(
	clientID,
	clientSecret,
	discoveryURL,
	registrationURL,
	logoutURL,
	postLogoutURL,
	redirectURL string,
) (*Oauth2Config, error) {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, discoveryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create oauth config: %w", err)
	}

	//nolint: exhaustruct
	verifier := provider.Verifier(&oidc.Config{
		ClientID:                   clientID,
		SkipClientIDCheck:          false,
		SkipExpiryCheck:            false,
		SkipIssuerCheck:            false,
		InsecureSkipSignatureCheck: false,
		SupportedSigningAlgs:       []string{"RS256"},
	})

	return &Oauth2Config{
		verifier:        verifier,
		registrationURL: registrationURL,
		logoutURL:       logoutURL,
		postLogoutURL:   postLogoutURL,
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{oidc.ScopeOpenID, "email"},
			Endpoint:     provider.Endpoint(),
		},
	}, nil
}

type GetOrCreateUser[T any] func(ctx context.Context, username, email string) (T, error)

func NewOidcController[T any](
	sessionManager *scs.SessionManager,
	fn GetOrCreateUser[T],
	config *Oauth2Config,
) *OidcController[T] {
	return &OidcController[T]{sessionManager: sessionManager, getOrCreateUserFn: fn, oauth2Config: config}
}

func (c *OidcController[T]) RegisterRoutes(router *httputils.Router) {
	router.AddRoute("GET /login", c.loginHandler)
	router.AddRoute("GET /register", c.registerHandler)
	router.AddRoute("GET /auth/callback", c.authCallback)
	router.AddRoute("GET /logout", c.logoutHandler)
}

func (c *OidcController[T]) RedirectToAuthURL(w http.ResponseWriter, r *http.Request, destURL string) {
	state := uuid.New().String()

	if destURL != "" {
		state = fmt.Sprintf("%s?dest=%s", state, url.QueryEscape(destURL))
	}

	//nolint: exhaustruct
	http.SetCookie(w, &http.Cookie{
		Name:     "state",
		Value:    state,
		Quoted:   false,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		// use lax since we are using a third-party for auth
		SameSite: http.SameSiteLaxMode,
	})

	url := c.oauth2Config.config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (c *OidcController[T]) loginHandler(w http.ResponseWriter, r *http.Request) {
	c.RedirectToAuthURL(w, r, "")
}

func (c *OidcController[T]) registerHandler(w http.ResponseWriter, r *http.Request) {
	registrationURL, err := httputils.GetURL(c.oauth2Config.registrationURL, map[string]string{
		"client_id":     c.oauth2Config.config.ClientID,
		"response_type": "code",
		"redirect_uri":  c.oauth2Config.config.RedirectURL,
	})
	if err != nil {
		httputils.ServerErrorResponse(w, r, fmt.Errorf("failed to get registration url: %w", err))

		return
	}

	http.Redirect(w, r, registrationURL, http.StatusTemporaryRedirect)
}

func (c *OidcController[T]) authCallback(w http.ResponseWriter, r *http.Request) {
	//nolint: exhaustruct
	cookie := &http.Cookie{
		Name:     "state",
		Value:    "",
		Quoted:   false,
		Expires:  time.Now().Add(-1 * time.Hour),
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)

	err := verifyState(r)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)

		return
	}

	idToken, err := c.exchangeCodeForIDToken(r)
	if err != nil {
		httputils.ServerErrorResponse(w, r, err)

		return
	}

	// Extract custom claims
	var claims struct {
		Email string `json:"email"`
	}

	if err := idToken.Claims(&claims); err != nil {
		httputils.ServerErrorResponse(w, r, fmt.Errorf("failed to parse claims: %w", err))

		return
	}

	slog.InfoContext(r.Context(), "claims", "claims", claims)

	user, err := c.getOrCreateUserFn(r.Context(), uuid.New().String(), claims.Email)
	if err != nil {
		httputils.ServerErrorResponse(w, r, fmt.Errorf("failed to get or create user: %w", err))

		return
	}

	err = c.sessionManager.RenewToken(r.Context())
	if err != nil {
		httputils.ServerErrorResponse(w, r, fmt.Errorf("failed to renew session: %w", err))

		return
	}

	c.sessionManager.Put(r.Context(), "user", user)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (c *OidcController[T]) logoutHandler(w http.ResponseWriter, r *http.Request) {
	err := c.sessionManager.RenewToken(r.Context())
	if err != nil {
		httputils.ServerErrorResponse(w, r, fmt.Errorf("failed to renew session: %w", err))

		return
	}

	// Remove the user from the session data so that the user is
	// 'logged out'.
	c.sessionManager.Remove(r.Context(), "user")

	logoutURL, err := httputils.GetURL(c.oauth2Config.logoutURL, map[string]string{
		"client_id":  c.oauth2Config.config.ClientID,
		"logout_uri": c.oauth2Config.postLogoutURL,
	})
	if err != nil {
		httputils.ServerErrorResponse(w, r, fmt.Errorf("failed to get logout url: %w", err))

		return
	}

	http.Redirect(w, r, logoutURL, http.StatusSeeOther)
}

func verifyState(r *http.Request) error {
	state := r.URL.Query().Get("state")
	if state == "" {
		return fmt.Errorf("no state query param: %w", ErrInvalidState)
	}

	cookie, err := r.Cookie("state")
	if err != nil {
		return fmt.Errorf("failed to get state cookie: %w", err)
	}

	if state != cookie.Value {
		return fmt.Errorf("state mismatch: %w", ErrInvalidState)
	}

	return nil
}

func (c *OidcController[T]) exchangeCodeForIDToken(r *http.Request) (*oidc.IDToken, error) {
	code := r.URL.Query().Get("code")

	if code == "" {
		return nil, ErrNoCode
	}

	rawToken, err := c.oauth2Config.config.Exchange(r.Context(), code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	rawIDToken, ok := rawToken.Extra("id_token").(string)
	if !ok {
		return nil, ErrNoIDToken
	}

	idToken, err := c.oauth2Config.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify id token: %w", err)
	}

	return idToken, nil
}
