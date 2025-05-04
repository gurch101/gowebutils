package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gurch101/gowebutils/pkg/authutils"
	"github.com/gurch101/gowebutils/pkg/dbutils"
	"github.com/gurch101/gowebutils/pkg/stringutils"
)

var ErrInvalidTenantID = errors.New("invalid tenant_id in token payload")

func GetOrCreateUser(ctx context.Context, db dbutils.DB, email string, tokenPayload map[string]any) (authutils.User, error) {
	if tokenPayload != nil {
		_, ok := tokenPayload["tenant_id"].(float64)
		if !ok {
			return authutils.User{}, ErrInvalidTenantID
		}

		//nolint: err113
		return authutils.User{}, fmt.Errorf("TODO")
	} else {
		user, err := GetUserByEmail(ctx, db, email)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				user, err := RegisterUser(ctx, db, stringutils.NewUUID(), email, "")
				if err != nil {
					return authutils.User{}, err
				}

				return user, nil
			}

			return authutils.User{}, err
		}

		return user, nil
	}
}
