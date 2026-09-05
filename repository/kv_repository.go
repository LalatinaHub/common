package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type kvRepo struct {
	db *sql.DB
}

// NewKVRepository returns a KVRepository implementation using *sql.DB.
func NewKVRepository(db *sql.DB) KVRepository {
	return &kvRepo{db: db}
}

func (r *kvRepo) GetAll(ctx context.Context) (map[string]any, error) {
	result := make(map[string]any)

	rows, err := r.db.QueryContext(ctx, "SELECT id, key, value FROM kv;")
	if err != nil {
		return nil, fmt.Errorf("failed to query kv: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id    int64
			key   string
			value any
		)

		err := rows.Scan(&id, &key, &value)
		if err != nil {
			return nil, fmt.Errorf("failed to scan kv: %w", err)
		}

		result[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating kv: %w", err)
	}

	return result, nil
}
