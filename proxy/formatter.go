package proxy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/LalatinaHub/common/model"
)

// Formatter serializes a model.ProxyNode into a valid proxy URL.
type Formatter struct{}

// NewFormatter creates a new Formatter instance.
func NewFormatter() *Formatter {
	return &Formatter{}
}

// Format converts a model.ProxyNode into a *url.URL structure.
func (f *Formatter) Format(node *model.ProxyNode) (*url.URL, error) {
	if node == nil {
		return nil, fmt.Errorf("proxy node is nil")
	}

	vpn := strings.ToLower(strings.TrimSpace(node.VPN))
	switch vpn {
	case "shadowsocks", "ss":
		return f.formatShadowsocks(node)
	case "vmess":
		return f.formatVMess(node)
	case "vless":
		return f.formatVLESS(node)
	case "trojan":
		return f.formatTrojan(node)
	default:
		return nil, fmt.Errorf("unsupported proxy protocol: %s", node.VPN)
	}
}

// FormatString converts a model.ProxyNode into its string URL representation.
func (f *Formatter) FormatString(node *model.ProxyNode) (string, error) {
	u, err := f.Format(node)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Format is a package-level helper that converts a model.ProxyNode into a *url.URL.
func Format(node *model.ProxyNode) (*url.URL, error) {
	return NewFormatter().Format(node)
}

// FormatString is a package-level helper that converts a model.ProxyNode into its string URL representation.
func FormatString(node *model.ProxyNode) (string, error) {
	return NewFormatter().FormatString(node)
}

func (f *Formatter) formatShadowsocks(node *model.ProxyNode) (*url.URL, error) {
	if node.Server == "" {
		return nil, fmt.Errorf("shadowsocks server host is empty")
	}
	port := node.ServerPort
	if port <= 0 {
		port = 8388
	}

	method := node.Method
	if method == "" {
		method = "aes-256-gcm"
	}
	credRaw := fmt.Sprintf("%s:%s", method, node.Password)
	cred := base64.RawURLEncoding.EncodeToString([]byte(credRaw))

	rawURL := fmt.Sprintf("ss://%s@%s:%d", cred, node.Server, port)
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to build shadowsocks URL: %w", err)
	}

	if node.Plugin != "" {
		q := u.Query()
		if node.PluginOpts != "" {
			q.Set("plugin", fmt.Sprintf("%s;%s", node.Plugin, node.PluginOpts))
		} else {
			q.Set("plugin", node.Plugin)
		}
		u.RawQuery = q.Encode()
	}

	if node.Remark != "" {
		u.Fragment = node.Remark
	}

	return u, nil
}

func (f *Formatter) formatVMess(node *model.ProxyNode) (*url.URL, error) {
	if node.Server == "" {
		return nil, fmt.Errorf("vmess server host is empty")
	}
	port := node.ServerPort
	if port <= 0 {
		port = 443
	}

	tlsStr := ""
	if node.TLS {
		tlsStr = "tls"
	}

	scy := node.Security
	if scy == "" {
		scy = "auto"
	}

	transport := node.Transport
	if transport == "" {
		transport = "tcp"
	}

	params := map[string]any{
		"v":    "2",
		"ps":   node.Remark,
		"add":  node.Server,
		"port": port,
		"id":   node.UUID,
		"aid":  node.AlterID,
		"scy":  scy,
		"net":  transport,
		"type": "none",
		"host": node.Host,
		"path": node.Path,
		"tls":  tlsStr,
		"sni":  node.SNI,
	}

	jsonBytes, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal vmess config to JSON: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(jsonBytes)
	u, err := url.Parse("vmess://" + encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to build vmess URL: %w", err)
	}

	return u, nil
}

func (f *Formatter) formatVLESS(node *model.ProxyNode) (*url.URL, error) {
	if node.Server == "" {
		return nil, fmt.Errorf("vless server host is empty")
	}
	port := node.ServerPort
	if port <= 0 {
		port = 443
	}

	baseURL := fmt.Sprintf("vless://%s@%s:%d", node.UUID, node.Server, port)
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to build vless URL: %w", err)
	}

	q := u.Query()
	if node.Transport != "" {
		q.Set("type", node.Transport)
	}
	if node.Security != "" {
		q.Set("security", node.Security)
	} else if node.TLS {
		q.Set("security", "tls")
	}
	if node.Host != "" {
		q.Set("host", node.Host)
	}
	if node.Path != "" {
		q.Set("path", node.Path)
	}
	if node.SNI != "" {
		q.Set("sni", node.SNI)
	}
	if node.ServiceName != "" {
		q.Set("serviceName", node.ServiceName)
	}

	u.RawQuery = q.Encode()
	if node.Remark != "" {
		u.Fragment = node.Remark
	}

	return u, nil
}

func (f *Formatter) formatTrojan(node *model.ProxyNode) (*url.URL, error) {
	if node.Server == "" {
		return nil, fmt.Errorf("trojan server host is empty")
	}
	port := node.ServerPort
	if port <= 0 {
		port = 443
	}

	baseURL := fmt.Sprintf("trojan://%s@%s:%d", url.QueryEscape(node.Password), node.Server, port)
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to build trojan URL: %w", err)
	}

	q := u.Query()
	if node.Transport != "" {
		q.Set("type", node.Transport)
	}
	if node.Security != "" {
		q.Set("security", node.Security)
	} else if node.TLS {
		q.Set("security", "tls")
	}
	if node.Host != "" {
		q.Set("host", node.Host)
	}
	if node.Path != "" {
		q.Set("path", node.Path)
	}
	if node.SNI != "" {
		q.Set("sni", node.SNI)
	}

	u.RawQuery = q.Encode()
	if node.Remark != "" {
		u.Fragment = node.Remark
	}

	return u, nil
}
