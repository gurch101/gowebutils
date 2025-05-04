package authutils

import "github.com/gurch101/gowebutils/pkg/collectionutils"

// User represents a user in the system.
type User struct {
	ID          int64
	TenantID    int64
	UserName    string
	Email       string
	IsAdmin     bool
	Permissions map[string]bool
}

// HasPermission checks if the user has a specific permission.
func (u *User) HasPermission(permissionCode string) bool {
	permission, exists := u.Permissions[permissionCode]

	return exists && permission
}

// HasAnyPermission checks if the user has any of the given permissions.
func (u *User) HasAnyPermission(permissionCodes ...string) bool {
	return collectionutils.Contains(permissionCodes, u.HasPermission)
}

// HasAllPermissions checks if the user has all of the given permissions.
func (u *User) HasAllPermissions(permissionCodes ...string) bool {
	return collectionutils.ContainsAll(permissionCodes, u.HasPermission)
}
