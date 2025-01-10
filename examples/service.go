package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/mailutils"
	"github.com/gurch101/gowebutils/pkg/stringutils"
)

type User struct {
	ID       int64
	TenantID int64
	UserName string
	Email    string
}

type AuthService struct {
	DB       *sql.DB
	mailer   *mailutils.Mailer
	hostName string
}

func NewAuthService(db *sql.DB, mailer *mailutils.Mailer, hostName string) *AuthService {
	return &AuthService{DB: db, mailer: mailer, hostName: hostName}
}

func (a *AuthService) GetUserExists(ctx context.Context, user User) bool {
	return dbutils.Exists(ctx, a.DB, "users", user.ID)
}

func (a *AuthService) GetUser(ctx context.Context, userid int64) (User, error) {
	var user User
	err := dbutils.GetByID(ctx, a.DB, "users", userid, map[string]any{
		"id":        &user.ID,
		"tenant_id": &user.TenantID,
		"user_name": &user.UserName,
		"email":     &user.Email,
	})

	if err != nil {
		return User{}, fmt.Errorf("get user query failed: %w", err)
	}

	return user, nil
}

func (a *AuthService) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := dbutils.NewQueryBuilder(a.DB).Select("id,tenant_id,user_name,email").From("users").Where("email = ?", email).QueryRow(&user.ID, &user.TenantID, &user.UserName, &user.Email)

	if err != nil {
		return User{}, fmt.Errorf("get user by email failed: %w", err)
	}

	return user, nil
}

func (a *AuthService) RegisterUser(ctx context.Context, username, email, inviteToken string) (User, error) {
	var userID *int64
	var err error

	if inviteToken != "" {
		// TODO
		// userID, err = a.registerInvitedUser(ctx, username, email, inviteToken)

		// if err != nil {
		// 	return User{}, fmt.Errorf("failed to register invited user: %w", err)
		// }
	} else {
		userID, err = a.registerNewUser(ctx, username, email)

		if err != nil {
			return User{}, fmt.Errorf("failed to register new user: %w", err)
		}
	}

	user, err := a.GetUser(ctx, *userID)

	if err != nil {
		return User{}, fmt.Errorf("failed to get user during register user: %w", err)
	}

	return user, nil
}

// func (a *AuthService) registerInvitedUser(ctx context.Context, username, email, inviteToken string) (*int64, error) {
// }

func (a *AuthService) registerNewUser(ctx context.Context, username, email string) (*int64, error) {
	var userID *int64
	err := dbutils.WithTransaction(ctx, a.DB, func(tx *sql.Tx) error {
		tenantID, err := dbutils.Insert(ctx, tx, "tenants", map[string]any{
			"tenant_name":   uuid.New().String(),
			"contact_email": email,
			"plan":          "free",
		})

		if err != nil {
			return fmt.Errorf("failed to create tenant: %w", err)
		}

		userID, err = dbutils.Insert(ctx, tx, "users", map[string]any{
			"tenant_id": tenantID,
			"user_name": username,
			"email":     email,
		})

		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		return nil
	})

	if err != nil || userID == nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	return userID, nil
}

func (a *AuthService) InviteUser(ctx context.Context, tenantID int64, username, email string) error {
	a.mailer.Send(email, "invite.go.tmpl", map[string]string{
		"URL": fmt.Sprintf("%s/register?code=%s", a.hostName, stringutils.NewUUID()),
	})

	return nil
}

func (a *AuthService) GetOrCreateUser(ctx context.Context, email string) (User, error) {
	user, err := a.GetUserByEmail(ctx, email)
	slog.InfoContext(ctx, "getOrCreateUser", "user exists?", err != nil)
	if err != nil {
		if errors.Is(err, dbutils.ErrRecordNotFound) {
			user, err := a.RegisterUser(ctx, stringutils.NewUUID(), email, "")
			if err != nil {
				return User{}, err
			}
			slog.InfoContext(ctx, "getOrCreateUser", "user created", user)
			return user, nil
		}
		return User{}, err
	}
	return user, nil
}
