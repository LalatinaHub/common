package model

// ProxyNode represents a generic relay proxy node from external database or parsed protocol URL.
// It contains all configuration properties required to reconstruct or convert the node to various protocols
// (such as Shadowsocks, VMess, VLESS, and Trojan).
type ProxyNode struct {
	ID          int64  `json:"id"`
	Server      string `json:"server"`
	IP          string `json:"ip"`
	ServerPort  int    `json:"server_port"`
	UUID        string `json:"uuid"`
	Password    string `json:"password"`
	Security    string `json:"security"`
	AlterID     int    `json:"alter_id"`
	Method      string `json:"method"`
	Plugin      string `json:"plugin"`
	PluginOpts  string `json:"plugin_opts"`
	Host        string `json:"host"`
	TLS         bool   `json:"tls"`
	Transport   string `json:"transport"`
	Path        string `json:"path"`
	ServiceName string `json:"service_name"`
	Insecure    bool   `json:"insecure"`
	SNI         string `json:"sni"`
	Remark      string `json:"remark"`
	ConnMode    string `json:"conn_mode"`
	CountryCode string `json:"country_code"`
	Region      string `json:"region"`
	Org         string `json:"org"`
	VPN         string `json:"vpn"`
	Raw         string `json:"raw"`
}
