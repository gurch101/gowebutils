package authutils

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
)

const sessionTimeout = 12 * time.Hour

func CreateSessionManager(db *sql.DB) *scs.SessionManager {
	secureSessionCookie := false

	_, err := os.Stat("./tls/cert.pem")
	if err == nil {
		secureSessionCookie = true
	}

	sessionManager := scs.New()
	sessionManager.Store = sqlite3store.New(db)
	sessionManager.Lifetime = sessionTimeout
	sessionManager.Cookie.Secure = secureSessionCookie

	if secureSessionCookie {
		sessionManager.Cookie.SameSite = http.SameSiteStrictMode
	}

	return sessionManager
}
