package main

import (
	"context"
	"fmt"

	"github.com/gurch101/gowebutils/pkg/authutils"
	"github.com/gurch101/gowebutils/pkg/dbutils"
)

func GetUserByID(ctx context.Context, db dbutils.DB, userid int64) (authutils.User, error) {
	var user authutils.User
	err := dbutils.GetByID(ctx, db, "users", userid, map[string]any{
		"id":        &user.ID,
		"tenant_id": &user.TenantID,
		"user_name": &user.UserName,
		"email":     &user.Email,
	})

	if err != nil {
		return authutils.User{}, fmt.Errorf("get user query failed: %w", err)
	}

	return user, nil
}
