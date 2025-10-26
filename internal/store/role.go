package store

import (
	"context"
	"database/sql"
)

type Role struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Level int    `json:"level"`
}

type RoleStore struct {
	db *sql.DB
}

func (s *RoleStore) HasPermission(ctx context.Context, requiredRole string, userRoleLevel int) (bool, error) {
	query := `
	SELECT level
	FROM roles
	WHERE name = $1	
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	var role Role
	if err := s.db.QueryRowContext(ctx, query, requiredRole).Scan(&role.Level); err != nil {
		return false, err
	}

	return userRoleLevel >= role.Level, nil

}
