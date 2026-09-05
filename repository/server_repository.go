package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/LalatinaHub/common/model"
)

type serverRepo struct {
	db *sql.DB
}

// NewServerRepository returns a ServerRepository implementation using *sql.DB.
func NewServerRepository(db *sql.DB) ServerRepository {
	return &serverRepo{db: db}
}

func (r *serverRepo) GetAll(ctx context.Context) ([]model.Server, error) {
	var servers []model.Server

	rows, err := r.db.QueryContext(ctx, "SELECT id, code, domain, ip, country, users_count, users_max FROM servers;")
	if err != nil {
		return nil, fmt.Errorf("failed to query servers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var s model.Server
		err := rows.Scan(
			&s.ID,
			&s.Code,
			&s.Domain,
			&s.IP,
			&s.Country,
			&s.UsersCount,
			&s.UsersMax,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan server: %w", err)
		}
		servers = append(servers, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating servers: %w", err)
	}

	return servers, nil
}
