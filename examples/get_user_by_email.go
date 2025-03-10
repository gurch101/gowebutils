package main

import (
	"context"
	"errors"

	"github.com/gurch101/gowebutils/pkg/authutils"
	"github.com/gurch101/gowebutils/pkg/dbutils"
)

var ErrUserNotFound = errors.New("user not found")

func GetUserByEmail(ctx context.Context, db dbutils.DB, email string) (authutils.User, error) {
	var user authutils.User
	err := dbutils.NewQueryBuilder(db).
		Select("id,tenant_id,user_name,email").
		From("users").
		Where("email = ?", email).
		QueryRowContext(ctx, &user.ID, &user.TenantID, &user.UserName, &user.Email)

	if err != nil {
		return authutils.User{}, ErrUserNotFound
	}

	return user, nil
}
