package gor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type repository[T any, ID any] struct {
	db *sql.DB
}

func (r *repository[T, ID]) Create(ctx context.Context, t T) (T, error) {
	row := r.db.QueryRowContext(ctx, `INSERT INTO user (name, email) returning id, name, email`)
	if row == nil {
		return t, errors.New("repository.Create: row is nil")
	}
	if row.Err() != nil {
		return t, fmt.Errorf("repository.Create: %w", row.Err())
	}

	err := row.Scan(t)
	if err != nil {
		return t, fmt.Errorf("repository.Create row scan: %w", err)
	}

	return t, nil
}
