package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gurch101/gowebutils/pkg/authutils"
	"github.com/gurch101/gowebutils/pkg/dbutils"
)

func RegisterUser(ctx context.Context, db dbutils.DB, username, email, inviteToken string) (authutils.User, error) {
	var userID *int64

	var err error

	if inviteToken != "" {

	} else {
		userID, err = registerNewUser(ctx, db, username, email)

		if err != nil {
			return authutils.User{}, fmt.Errorf("failed to register new user: %w", err)
		}
	}

	user, err := GetUserByID(ctx, db, *userID)

	if err != nil {
		return authutils.User{}, fmt.Errorf("failed to get user during register user: %w", err)
	}

	return user, nil
}

func registerNewUser(ctx context.Context, db dbutils.DB, username, email string) (*int64, error) {
	var userID *int64

	err := dbutils.WithTransaction(ctx, db, func(tx dbutils.DB) error {
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
