package repository_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LalatinaHub/common/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKVRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewKVRepository(db)
	ctx := context.Background()

	t.Run("success fetching all KV", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "key", "value"}).
			AddRow(1, "cf_token", "token123").
			AddRow(2, "max_clients", 100)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, key, value FROM kv;")).
			WillReturnRows(rows)

		kv, err := repo.GetAll(ctx)
		require.NoError(t, err)
		assert.Equal(t, "token123", kv["cf_token"])
		assert.Equal(t, int64(100), kv["max_clients"])
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, key, value FROM kv;")).
			WillReturnError(assert.AnError)

		kv, err := repo.GetAll(ctx)
		assert.Error(t, err)
		assert.Nil(t, kv)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
