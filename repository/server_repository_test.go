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

func TestServerRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewServerRepository(db)
	ctx := context.Background()

	t.Run("success fetching servers", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "code", "domain", "ip", "country", "users_count", "users_max",
		}).
			AddRow(1, "SG1", "sg1.example.com", "1.1.1.1", "SG", 5, 50).
			AddRow(2, "ID1", "id1.example.com", "2.2.2.2", "ID", 10, 50)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, code, domain, ip, country, users_count, users_max FROM servers;")).
			WillReturnRows(rows)

		servers, err := repo.GetAll(ctx)
		require.NoError(t, err)
		assert.Len(t, servers, 2)
		assert.Equal(t, "SG1", servers[0].Code)
		assert.Equal(t, "ID1", servers[1].Code)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, code, domain, ip, country, users_count, users_max FROM servers;")).
			WillReturnError(assert.AnError)

		servers, err := repo.GetAll(ctx)
		assert.Error(t, err)
		assert.Nil(t, servers)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
