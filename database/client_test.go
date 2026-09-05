package database_test

import (
	"context"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LalatinaHub/common/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_MissingEnv(t *testing.T) {
	database.ResetDBInstance()
	defer database.ResetDBInstance()

	os.Unsetenv("TURSO_DATABASE_URL")
	os.Unsetenv("TURSO_AUTH_TOKEN")

	db, err := database.GetDB()
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "TURSO_DATABASE_URL environment variable not set")
}

func TestDatabase_SetDB_Ping_InitIndexes_Close(t *testing.T) {
	database.ResetDBInstance()
	defer database.ResetDBInstance()

	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer mockDB.Close()

	database.SetDB(mockDB)

	ctx := context.Background()

	t.Run("ping success", func(t *testing.T) {
		mock.ExpectPing()
		err := database.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("init indexes success", func(t *testing.T) {
		indexes := []string{
			`CREATE INDEX IF NOT EXISTS idx_users_expired ON users(expired);`,
			`CREATE INDEX IF NOT EXISTS idx_users_quota ON users(quota);`,
			`CREATE INDEX IF NOT EXISTS idx_users_server_code ON users(server_code);`,
			`CREATE INDEX IF NOT EXISTS idx_users_vpn ON users(vpn);`,
		}

		for _, idx := range indexes {
			mock.ExpectExec(regexp.QuoteMeta(idx)).WillReturnResult(sqlmock.NewResult(0, 0))
		}

		err := database.InitIndexes(ctx)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("close success", func(t *testing.T) {
		mock.ExpectClose()
		err := database.Close()
		assert.NoError(t, err)
	})
}
