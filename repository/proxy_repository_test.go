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

func TestProxyRepository_GetRelays(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewProxyRepository(db)
	ctx := context.Background()

	cols := []string{
		"id", "server", "ip", "server_port", "uuid", "password", "security",
		"alter_id", "method", "plugin", "plugin_opts", "host", "tls",
		"transport", "path", "service_name", "insecure", "sni", "remark",
		"conn_mode", "country_code", "region", "org", "vpn", "raw",
	}

	t.Run("filters excluded country and caps max per country", func(t *testing.T) {
		rows := sqlmock.NewRows(cols).
			AddRow(1, "s1.com", "1.1.1.1", 8388, "", "p1", "", 0, "aes-128-gcm", "", "", "", false, "tcp", "", "", false, "", "SG1", "direct", "SG", "Asia", "Org1", "shadowsocks", "raw1").
			AddRow(2, "s2.com", "1.1.1.2", 8388, "", "p2", "", 0, "aes-128-gcm", "", "", "", false, "tcp", "", "", false, "", "SG2", "direct", "SG", "Asia", "Org2", "shadowsocks", "raw2").
			AddRow(3, "s3.com", "1.1.1.3", 8388, "", "p3", "", 0, "aes-128-gcm", "", "", "", false, "tcp", "", "", false, "", "SG3", "direct", "SG", "Asia", "Org3", "shadowsocks", "raw3").
			AddRow(4, "s4.com", "2.2.2.1", 8388, "", "p4", "", 0, "aes-128-gcm", "", "", "", false, "tcp", "", "", false, "", "ID1", "direct", "ID", "Asia", "Org4", "shadowsocks", "raw4").
			AddRow(5, "s5.com", "3.3.3.1", 8388, "", "p5", "", 0, "aes-128-gcm", "", "", "", false, "tcp", "", "", false, "", "US1", "direct", "US", "Americas", "Org5", "shadowsocks", "raw5")

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, server, ip, server_port, uuid, password, security, alter_id, method, plugin, plugin_opts, host, tls, transport, path, service_name, insecure, sni, remark, conn_mode, country_code, region, org, vpn, raw FROM proxies WHERE vpn = 'shadowsocks' AND method NOT LIKE '202%';")).
			WillReturnRows(rows)

		// Exclude "ID", and max 2 per country
		relays, err := repo.GetRelays(ctx, []string{"ID"}, 2)
		require.NoError(t, err)

		// Should have 2 SG, 0 ID (excluded), 1 US => total 3
		assert.Len(t, relays, 3)
		assert.Equal(t, "SG1", relays[0].Remark)
		assert.Equal(t, "SG2", relays[1].Remark)
		assert.Equal(t, "US1", relays[2].Remark)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, server, ip, server_port, uuid, password, security, alter_id, method, plugin, plugin_opts, host, tls, transport, path, service_name, insecure, sni, remark, conn_mode, country_code, region, org, vpn, raw FROM proxies WHERE vpn = 'shadowsocks' AND method NOT LIKE '202%';")).
			WillReturnError(assert.AnError)

		relays, err := repo.GetRelays(ctx, []string{"ID"}, 2)
		assert.Error(t, err)
		assert.Nil(t, relays)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
