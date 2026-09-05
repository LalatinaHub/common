// Package repository provides data access interfaces and SQL implementations for domain entities.
package repository

import (
	"context"

	"github.com/LalatinaHub/common/model"
)

// UserRepository handles persistence queries for User entities.
type UserRepository interface {
	// GetActiveUsersGroupedByVPN fetches users where quota > 0 and expired date >= now.
	// Returns map keyed by VPN protocol ("trojan", "vmess", "vless").
	GetActiveUsersGroupedByVPN(ctx context.Context) (map[string][]model.User, error)

	// DeductQuota subtracts usedBytes from user's current quota.
	// Returns the remaining quota in bytes, whether quota is depleted (<= 0), and any error.
	DeductQuota(ctx context.Context, userID int64, usedBytes int64) (remainingBytes int64, isDepleted bool, err error)

	// DeductQuotaBatch updates quotas in a single transaction with prepared statements.
	// Returns IDs of users whose quota has become depleted (<= 0).
	DeductQuotaBatch(ctx context.Context, usages map[int64]int64) (depletedUserIDs []int64, err error)

	// CreateUser inserts a new user record into the users table and returns the generated ID.
	CreateUser(ctx context.Context, u *model.User) (int64, error)
}

// ServerRepository handles persistence queries for Server entities.
type ServerRepository interface {
	// GetAll fetches all cluster server records.
	GetAll(ctx context.Context) ([]model.Server, error)
}

// KVRepository handles key-value configuration queries.
type KVRepository interface {
	// GetAll fetches all key-value entries as a map.
	GetAll(ctx context.Context) (map[string]any, error)
}

// ProxyRepository handles proxy relay database queries.
type ProxyRepository interface {
	// GetRelays fetches Shadowsocks relay proxies filtered by excluded country codes
	// and capped at maxPerCountry per country code.
	GetRelays(ctx context.Context, excludedCountryCodes []string, maxPerCountry int) ([]model.ProxyNode, error)
}
