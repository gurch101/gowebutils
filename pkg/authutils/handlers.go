package authutils

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"gurch101.github.io/go-web/pkg/httputils"
)

// ErrAuthCodeNotFound is returned when the authorization code is not found in the request.
var ErrAuthCodeNotFound = errors.New("authorization code not found")

// ErrInvalidState is returned when the state is invalid.
var ErrInvalidState = errors.New("invalid state")

// ErrTokenExchangeFailed is returned when the token exchange fails.
var ErrTokenExchangeFailed = errors.New("token exchange failed")

type OidcController struct {
	DB           *sql.DB
	oauth2Config *Oauth2Config
}

type Oauth2Config struct {
	verifier *oidc.IDTokenVerifier
	config   *oauth2.Config
}

func formatCallbackURL(host string) string {
	host = strings.TrimSuffix(host, "/")

	return host + "/auth/callback"
}

// CreateOauthConfig creates an oauth2.Config object for the given idp URL.
// for azure, the should https://<tenant-id>.b2clogin.com/<tenant-id>.onmicrosoft.com/<policy-name>/v2.0/
// spURL should be the host URL of your app.
func CreateOauthConfig(clientID, clientSecret, discoveryURL, issuerURL, spURL string) (*Oauth2Config, error) {
	ctx := context.Background()
	if issuerURL != "" {
		ctx = oidc.InsecureIssuerURLContext(ctx, issuerURL)
	}

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
		verifier: verifier,
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  formatCallbackURL(spURL),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
			Endpoint:     provider.Endpoint(),
		},
	}, nil
}

func NewOidcController(db *sql.DB, config *Oauth2Config) *OidcController {
	return &OidcController{DB: db, oauth2Config: config}
}

func (c *OidcController) RegisterRoutes(router *httputils.Router) {
	router.AddRoute("GET /login", c.loginHandler)
	router.AddRoute("GET /auth/callback", c.callbackHandler)
}

func (c *OidcController) RedirectToAuthURL(w http.ResponseWriter, r *http.Request, destURL string) {
	state := uuid.New().String()

	if destURL != "" {
		state = fmt.Sprintf("%s?dest=%s", state, url.QueryEscape(destURL))
	}

	// TODO: make secure
	//nolint: exhaustruct
	http.SetCookie(w, &http.Cookie{
		Name:     "state",
		Value:    state,
		Quoted:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	url := c.oauth2Config.config.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (c *OidcController) loginHandler(w http.ResponseWriter, r *http.Request) {
	c.RedirectToAuthURL(w, r, "")
}

// need to use this since the built-in exchange requires an access_token.
func (c *OidcController) exchange(ctx context.Context, code string) (string, error) {
	if code == "" {
		return "", ErrAuthCodeNotFound
	}

	tokenURL := c.oauth2Config.config.Endpoint.TokenURL
	// Exchange the authorization code for an id token
	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("client_id", c.oauth2Config.config.ClientID)
	form.Add("redirect_uri", c.oauth2Config.config.RedirectURL)
	form.Add("code", code)
	form.Add("scope", strings.Join(c.oauth2Config.config.Scopes, " "))
	form.Add("client_secret", c.oauth2Config.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(httputils.ContentTypeHeader, "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to exchange token: %w", err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if err != nil {
			if closeErr != nil {
				slog.ErrorContext(ctx, "failed to close response body", "error", closeErr)
			}

			return
		}

		err = closeErr
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %s", ErrTokenExchangeFailed, string(body))
	}

	var tokenResponse struct {
		IDToken string `json:"id_token"` //nolint:tagliatelle
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	return tokenResponse.IDToken, nil
}

func (c *OidcController) callbackHandler(w http.ResponseWriter, r *http.Request) {
	// Verify state and errors.
	state := r.URL.Query().Get("state")
	cookie, err := r.Cookie("state")

	if err != nil || cookie.Value != state {
		httputils.BadRequestResponse(w, r, ErrInvalidState)

		return
	}

	code := r.URL.Query().Get("code")

	// changed from context.Background() to r.Context()
	rawIDToken, err := c.exchange(r.Context(), code)
	if err != nil {
		httputils.ServerErrorResponse(w, r, fmt.Errorf("failed to exchange token: %w", err))

		return
	}

	// Parse and verify ID Token payload.
	idToken, err := c.oauth2Config.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		httputils.ServerErrorResponse(w, r, fmt.Errorf("failed to verify id token: %w", err))

		return
	}

	// Extract custom claims
	var claims struct {
		Oid string `json:"oid"`
	}

	if err := idToken.Claims(&claims); err != nil {
		httputils.ServerErrorResponse(w, r, fmt.Errorf("failed to parse claims: %w", err))

		return
	}
}
