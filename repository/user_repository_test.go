package repository_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LalatinaHub/common/model"
	"github.com/LalatinaHub/common/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_GetActiveUsersGroupedByVPN(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	t.Run("success fetching active users grouped", func(t *testing.T) {
		futureDate := time.Now().AddDate(0, 1, 0).Format("2006-01-02")
		pastDate := time.Now().AddDate(0, -1, 0).Format("2006-01-02")

		rows := sqlmock.NewRows([]string{
			"id", "token", "password", "expired", "server_code", "quota", "relay", "adblock", "vpn",
		}).
			AddRow(1, "token1", "pass1", futureDate, "ID", 1000, 1, 1, "vmess").
			AddRow(2, "token2", "pass2", futureDate, "SG", 500, 0, 0, "trojan").
			AddRow(3, "token3", "pass3", pastDate, "US", 500, 0, 0, "vmess"). // Expired
			AddRow(4, "token4", "pass4", futureDate, "US", 0, 0, 0, "vless")   // Zero quota

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, token, password, expired, server_code, quota, relay, adblock, vpn FROM users;")).
			WillReturnRows(rows)

		usersMap, err := repo.GetActiveUsersGroupedByVPN(ctx)
		require.NoError(t, err)

		assert.Len(t, usersMap["vmess"], 1)
		assert.Equal(t, int64(1), usersMap["vmess"][0].ID)
		assert.True(t, usersMap["vmess"][0].Adblock)

		assert.Len(t, usersMap["trojan"], 1)
		assert.Equal(t, int64(2), usersMap["trojan"][0].ID)
		assert.False(t, usersMap["trojan"][0].Adblock)

		assert.Empty(t, usersMap["vless"])
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error handling", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, token, password, expired, server_code, quota, relay, adblock, vpn FROM users;")).
			WillReturnError(assert.AnError)

		usersMap, err := repo.GetActiveUsersGroupedByVPN(ctx)
		assert.Error(t, err)
		assert.Nil(t, usersMap)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_DeductQuota(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	t.Run("success deducting quota and remaining", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT quota FROM users WHERE id = ?;")).
			WithArgs(int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"quota"}).AddRow(100))

		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET quota = ? WHERE id = ?;")).
			WithArgs(int64(90), int64(1)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		remaining, depleted, err := repo.DeductQuota(ctx, 1, 10_000_000) // 10MB
		require.NoError(t, err)
		assert.Equal(t, int64(90), remaining)
		assert.False(t, depleted)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("depleted quota <= 0", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT quota FROM users WHERE id = ?;")).
			WithArgs(int64(2)).
			WillReturnRows(sqlmock.NewRows([]string{"quota"}).AddRow(5))

		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET quota = ? WHERE id = ?;")).
			WithArgs(int64(-5), int64(2)).
			WillReturnResult(sqlmock.NewResult(2, 1))

		remaining, depleted, err := repo.DeductQuota(ctx, 2, 10_000_000) // 10MB
		require.NoError(t, err)
		assert.Equal(t, int64(-5), remaining)
		assert.True(t, depleted)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_DeductQuotaBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	t.Run("empty usages returns immediately", func(t *testing.T) {
		depleted, err := repo.DeductQuotaBatch(ctx, map[int64]int64{})
		require.NoError(t, err)
		assert.Nil(t, depleted)
	})

	t.Run("batch transaction with depleted user", func(t *testing.T) {
		usages := map[int64]int64{
			1: 10_000_000, // 10MB
		}

		mock.ExpectBegin()
		mock.ExpectPrepare(regexp.QuoteMeta("UPDATE users SET quota = quota - ? WHERE id = ?;"))
		mock.ExpectPrepare(regexp.QuoteMeta("SELECT quota FROM users WHERE id = ?;"))

		// Exec update
		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET quota = quota - ? WHERE id = ?;")).
			WithArgs(int64(10), int64(1)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Select remaining quota
		mock.ExpectQuery(regexp.QuoteMeta("SELECT quota FROM users WHERE id = ?;")).
			WithArgs(int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"quota"}).AddRow(0))

		mock.ExpectCommit()

		depleted, err := repo.DeductQuotaBatch(ctx, usages)
		require.NoError(t, err)
		assert.Equal(t, []int64{1}, depleted)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepository_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	u := &model.User{
		Token:      "new-token",
		Password:   "new-pass",
		Expired:    time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		ServerCode: "SG",
		Quota:      5000,
		Relay:      "relay-1",
		Adblock:    true,
		VPN:        "vmess",
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (token, password, expired, server_code, quota, relay, adblock, vpn) VALUES (?, ?, ?, ?, ?, ?, ?, ?);")).
		WithArgs("new-token", "new-pass", "2026-12-31", "SG", int64(5000), "relay-1", 1, "vmess").
		WillReturnResult(sqlmock.NewResult(100, 1))

	id, err := repo.CreateUser(ctx, u)
	require.NoError(t, err)
	assert.Equal(t, int64(100), id)
	assert.Equal(t, int64(100), u.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
