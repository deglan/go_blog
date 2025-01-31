package store

import (
	"context"
	"database/sql"
)

type Role struct {
	Id          int64
	Name        string
	Level       int
	Description string
}

type RolesStore struct {
	db *sql.DB
}

func (s *RolesStore) GetByName(ctx context.Context, name string) (*Role, error) {
	query := `SELECT id, name, level, description FROM roles WHERE name = $1`

	role := &Role{}
	err := s.db.QueryRowContext(ctx, query, name).Scan(&role.Id, &role.Name, &role.Level, &role.Description)
	if err != nil {
		return nil, err
	}

	return role, nil
}
