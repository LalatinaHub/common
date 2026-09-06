package repository

import (
	"context"
	"database/sql"
	"fmt"
	"slices"

	"github.com/LalatinaHub/common/model"
	"github.com/LalatinaHub/common/proxy"
)

type proxyRepo struct {
	db *sql.DB
}

// NewProxyRepository returns a ProxyRepository implementation using *sql.DB.
func NewProxyRepository(db *sql.DB) ProxyRepository {
	return &proxyRepo{db: db}
}

func (r *proxyRepo) GetRelays(ctx context.Context, excludedCountryCodes []string, maxPerCountry int) ([]model.ProxyNode, error) {
	var (
		relays         []model.ProxyNode
		relayCodeCount = make(map[string]int)
	)

	query := "SELECT id, server, ip, server_port, uuid, password, security, alter_id, method, plugin, plugin_opts, host, tls, transport, path, service_name, insecure, sni, remark, conn_mode, country_code, region, org, vpn, raw FROM proxies WHERE vpn = 'shadowsocks' AND method NOT LIKE '202%';"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query proxies: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p model.ProxyNode
		err := rows.Scan(
			&p.ID,
			&p.Server,
			&p.IP,
			&p.ServerPort,
			&p.UUID,
			&p.Password,
			&p.Security,
			&p.AlterID,
			&p.Method,
			&p.Plugin,
			&p.PluginOpts,
			&p.Host,
			&p.TLS,
			&p.Transport,
			&p.Path,
			&p.ServiceName,
			&p.Insecure,
			&p.SNI,
			&p.Remark,
			&p.ConnMode,
			&p.CountryCode,
			&p.Region,
			&p.Org,
			&p.VPN,
			&p.Raw,
		)
		if err != nil {
			continue // Gracefully skip unparseable proxy row
		}

		p.Raw = proxy.DecodeIfBase64(p.Raw)

		if slices.Contains(excludedCountryCodes, p.CountryCode) {
			continue
		}

		if relayCodeCount[p.CountryCode] < maxPerCountry {
			relays = append(relays, p)
			relayCodeCount[p.CountryCode]++
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating proxies: %w", err)
	}

	return relays, nil
}
