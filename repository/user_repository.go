package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/LalatinaHub/common/model"
)

type userRepo struct {
	db *sql.DB
}

// NewUserRepository returns a UserRepository implementation using *sql.DB.
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) GetActiveUsersGroupedByVPN(ctx context.Context) (map[string][]model.User, error) {
	result := make(map[string][]model.User)
	now := time.Now()

	rows, err := r.db.QueryContext(ctx, "SELECT id, token, password, expired, server_code, quota, relay, adblock, vpn FROM users;")
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			u          model.User
			expiredStr string
			adblockVal any
		)

		err := rows.Scan(
			&u.ID,
			&u.Token,
			&u.Password,
			&expiredStr,
			&u.ServerCode,
			&u.Quota,
			&u.Relay,
			&adblockVal,
			&u.VPN,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if b, ok := adblockVal.(bool); ok {
			u.Adblock = b
		} else if i, ok := adblockVal.(int64); ok {
			u.Adblock = i > 0
		}

		parsedExpired, err := time.Parse("2006-01-02", expiredStr)
		if err == nil {
			u.Expired = parsedExpired
		}

		if u.IsActive(now) {
			result[u.VPN] = append(result[u.VPN], u)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return result, nil
}

func (r *userRepo) DeductQuota(ctx context.Context, userID int64, usedBytes int64) (int64, bool, error) {
	var currentQuota int64
	row := r.db.QueryRowContext(ctx, "SELECT quota FROM users WHERE id = ?;", userID)
	if err := row.Scan(&currentQuota); err != nil {
		return 0, false, fmt.Errorf("failed to fetch user %d quota: %w", userID, err)
	}

	usedMB := usedBytes / 1_000_000
	if usedMB > 0 {
		currentQuota -= usedMB
		_, err := r.db.ExecContext(ctx, "UPDATE users SET quota = ? WHERE id = ?;", currentQuota, userID)
		if err != nil {
			return currentQuota, false, fmt.Errorf("failed to update quota for user %d: %w", userID, err)
		}
	}

	return currentQuota, currentQuota <= 0, nil
}

// DeductQuotaBatch updates multiple users' quotas in a single transaction using prepared statements.
// Returns IDs of users whose quota has become depleted (<= 0).
func (r *userRepo) DeductQuotaBatch(ctx context.Context, usages map[int64]int64) ([]int64, error) {
	if len(usages) == 0 {
		return nil, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare update statement
	updateStmt, err := tx.PrepareContext(ctx, "UPDATE users SET quota = quota - ? WHERE id = ?;")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer updateStmt.Close()

	// Prepare select statement to check resulting quota
	selectStmt, err := tx.PrepareContext(ctx, "SELECT quota FROM users WHERE id = ?;")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select statement: %w", err)
	}
	defer selectStmt.Close()

	var depletedUserIDs []int64

	for userID, usedBytes := range usages {
		usedMB := usedBytes / 1_000_000
		if usedMB <= 0 {
			continue
		}

		// Update quota
		if _, err := updateStmt.ExecContext(ctx, usedMB, userID); err != nil {
			return nil, fmt.Errorf("failed to update quota for user %d: %w", userID, err)
		}

		// Check if quota is depleted
		var currentQuota int64
		if err := selectStmt.QueryRowContext(ctx, userID).Scan(&currentQuota); err != nil {
			// User might not exist or query failed, skip
			continue
		}

		if currentQuota <= 0 {
			depletedUserIDs = append(depletedUserIDs, userID)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return depletedUserIDs, nil
}

func (r *userRepo) CreateUser(ctx context.Context, u *model.User) (int64, error) {
	expiredStr := u.Expired.Format("2006-01-02")
	adblockVal := 0
	if u.Adblock {
		adblockVal = 1
	}

	res, err := r.db.ExecContext(ctx,
		"INSERT INTO users (token, password, expired, server_code, quota, relay, adblock, vpn) VALUES (?, ?, ?, ?, ?, ?, ?, ?);",
		u.Token, u.Password, expiredStr, u.ServerCode, u.Quota, u.Relay, adblockVal, u.VPN,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	u.ID = id
	return id, nil
}
