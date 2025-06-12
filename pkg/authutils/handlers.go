package authutils

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/gurch101/gowebutils/pkg/httputils"
	"github.com/gurch101/gowebutils/pkg/parser"
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

// ErrInvalidInviteToken is returned when the invite token is invalid.
var ErrInvalidInviteToken = errors.New("invalid invite token")

type OidcController struct {
	oauth2Config             *oauth2Config
	getOrCreateUserFn        GetOrCreateUser
	sessionManager           *scs.SessionManager
	redirectURL              string
	secureStateSessionCookie bool
}

type oauth2Config struct {
	verifier        *oidc.IDTokenVerifier
	registrationURL string
	logoutURL       string
	postLogoutURL   string
	config          *oauth2.Config
}

// createOauthConfig creates an oauth2.Config object for the given idp URL.
// discoveryURL is the base URL that exposes /.well-known/openid-configuration
// spURL should be the host URL of your app.
func createOauthConfig(
	clientID,
	clientSecret,
	discoveryURL,
	registrationURL,
	logoutURL,
	postLogoutURL,
	authCallbackURL string,
) (*oauth2Config, error) {
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

	return &oauth2Config{
		verifier:        verifier,
		registrationURL: registrationURL,
		logoutURL:       logoutURL,
		postLogoutURL:   postLogoutURL,
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  authCallbackURL,
			Scopes:       []string{oidc.ScopeOpenID, "email"},
			Endpoint:     provider.Endpoint(),
		},
	}, nil
}

type GetOrCreateUser func(ctx context.Context, email string, inviteTokenPayload map[string]any) (User, error)

func CreateOidcController(
	sessionManager *scs.SessionManager,
	getOrCreateUserFn GetOrCreateUser,
) *OidcController {
	///nolint: exhaustruct
	gob.Register(User{})

	config, err := createOauthConfig(
		parser.ParseEnvStringPanic("OIDC_CLIENT_ID"),
		parser.ParseEnvStringPanic("OIDC_CLIENT_SECRET"),
		parser.ParseEnvStringPanic("OIDC_DISCOVERY_URL"),
		parser.ParseEnvStringPanic("REGISTRATION_URL"),
		parser.ParseEnvStringPanic("LOGOUT_URL"),
		parser.ParseEnvStringPanic("POST_LOGOUT_REDIRECT_URL"),
		parser.ParseEnvStringPanic("HOST")+"/auth/callback",
	)
	if err != nil {
		slog.Error("failed to create oauth config", "error", err)
		panic(err)
	}

	redirectURL := parser.ParseEnvString("REDIRECT_URL", "/")

	secureSessionCookie := false

	_, err = os.Stat("./tls/cert.pem")
	if err == nil {
		secureSessionCookie = true
	}

	return &OidcController{sessionManager: sessionManager, getOrCreateUserFn: getOrCreateUserFn, oauth2Config: config, redirectURL: redirectURL, secureStateSessionCookie: secureSessionCookie}
}

func NewOidcController(
	sessionManager *scs.SessionManager,
	fn GetOrCreateUser,
	config *oauth2Config,
) *OidcController {
	return &OidcController{sessionManager: sessionManager, getOrCreateUserFn: fn, oauth2Config: config}
}

func (c *OidcController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	c.redirectToAuthURL(w, r, nil)
}

func (c *OidcController) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	payload := map[string]any{
		"state": uuid.New().String(),
	}

	defaultInvite := ""

	invite := parser.ParseQSString(r.URL.Query(), "invite", &defaultInvite)
	if *invite != "" {
		_, err := VerifyInviteToken(*invite)
		if err != nil {
			httputils.BadRequestResponse(w, r, ErrInvalidInviteToken)

			return
		}

		payload["invite"] = *invite
	}

	state, err := Encrypt(payload)
	if err != nil {
		httputils.ServerErrorResponse(w, r, fmt.Errorf("failed to encrypt state: %w", err))
	}

	//nolint: exhaustruct
	http.SetCookie(w, &http.Cookie{
		Name:     "state",
		Value:    state,
		Quoted:   false,
		HttpOnly: true,
		Secure:   c.secureStateSessionCookie,
		Path:     "/",
		// use lax since we are using a third-party for auth
		SameSite: http.SameSiteLaxMode,
	})

	registrationURL, err := httputils.GetURL(c.oauth2Config.registrationURL, map[string]string{
		"client_id":     c.oauth2Config.config.ClientID,
		"response_type": "code",
		"redirect_uri":  c.oauth2Config.config.RedirectURL,
		"state":         state,
	})
	if err != nil {
		httputils.ServerErrorResponse(w, r, fmt.Errorf("failed to get registration url: %w", err))

		return
	}

	http.Redirect(w, r, registrationURL, http.StatusTemporaryRedirect)
}

func (c *OidcController) AuthCallback(w http.ResponseWriter, r *http.Request) {
	state, err := verifyState(w, r, c.secureStateSessionCookie)
	if err != nil {
		slog.Info("failed to verify state", "error", err)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)

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

	var payload map[string]any

	invite, ok := state["invite"].(string)
	if ok {
		payload, err = VerifyInviteToken(invite)
		if err != nil {
			httputils.ServerErrorResponse(w, r, fmt.Errorf("failed to verify invite token: %w", err))

			return
		}
	}

	user, err := c.getOrCreateUserFn(r.Context(), claims.Email, payload)
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

	http.Redirect(w, r, c.redirectURL, http.StatusTemporaryRedirect)
}

func (c *OidcController) LogoutHandler(w http.ResponseWriter, r *http.Request) {
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

func (c *OidcController) redirectToAuthURL(w http.ResponseWriter, r *http.Request, payload map[string]any) {
	if payload == nil {
		payload = map[string]any{}
	}

	payload["state"] = uuid.New().String()

	state, err := Encrypt(payload)
	if err != nil {
		httputils.ServerErrorResponse(w, r, fmt.Errorf("failed to encrypt state: %w", err))
	}

	//nolint: exhaustruct
	http.SetCookie(w, &http.Cookie{
		Name:     "state",
		Value:    state,
		Quoted:   false,
		HttpOnly: true,
		Secure:   c.secureStateSessionCookie,
		Path:     "/",
		// use lax since we are using a third-party for auth
		SameSite: http.SameSiteLaxMode,
	})

	url := c.oauth2Config.config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func verifyState(w http.ResponseWriter, r *http.Request, secureCookie bool) (map[string]any, error) {
	//nolint: exhaustruct
	cookie := &http.Cookie{
		Name:     "state",
		Value:    "",
		Quoted:   false,
		Expires:  time.Now().Add(-1 * time.Hour),
		Path:     "/",
		Secure:   secureCookie,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)

	state := r.URL.Query().Get("state")
	if state == "" {
		return nil, fmt.Errorf("no state query param: %w", ErrInvalidState)
	}

	cookie, err := r.Cookie("state")
	if err != nil {
		return nil, fmt.Errorf("failed to get state cookie: %w", err)
	}

	if state != cookie.Value {
		return nil, fmt.Errorf("state mismatch: %w", ErrInvalidState)
	}

	payload, err := Decrypt(state)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt state: %w", err)
	}

	return payload, nil
}

func (c *OidcController) exchangeCodeForIDToken(r *http.Request) (*oidc.IDToken, error) {
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
