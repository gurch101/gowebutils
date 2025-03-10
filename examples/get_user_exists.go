package main

import (
	"context"

	"github.com/gurch101/gowebutils/pkg/authutils"
	"github.com/gurch101/gowebutils/pkg/dbutils"
)

// GetUserExists checks if a user with the given ID exists in the database.
func GetUserExists(ctx context.Context, db dbutils.DB, user authutils.User) bool {
	return dbutils.Exists(ctx, db, "users", user.ID)
}
