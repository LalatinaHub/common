// Package proxy provides parser utilities for various VPN proxy URLs (Shadowsocks, VMess, VLESS, Trojan).
// It has zero heavy dependencies and maps directly to the model.ProxyNode structure.
package proxy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/LalatinaHub/common/model"
)

// Parser provides native proxy URL parsing without external dependencies.
type Parser struct{}

// NewParser creates a new Parser instance.
func NewParser() *Parser {
	return &Parser{}
}

// DecodeIfBase64 checks if the input is a base64-encoded proxy URL and decodes it.
// If the input is already a plaintext proxy URL or cannot be decoded to one, it is returned as is.
func DecodeIfBase64(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "vless://") ||
		strings.HasPrefix(lower, "vmess://") ||
		strings.HasPrefix(lower, "trojan://") ||
		strings.HasPrefix(lower, "ss://") ||
		strings.HasPrefix(lower, "shadowsocks://") {
		return trimmed
	}

	isProxyURI := func(s string) bool {
		low := strings.ToLower(s)
		return strings.HasPrefix(low, "vless://") ||
			strings.HasPrefix(low, "vmess://") ||
			strings.HasPrefix(low, "trojan://") ||
			strings.HasPrefix(low, "ss://") ||
			strings.HasPrefix(low, "shadowsocks://")
	}

	for _, enc := range []*base64.Encoding{
		base64.StdEncoding,
		base64.URLEncoding,
		base64.RawStdEncoding,
		base64.RawURLEncoding,
	} {
		if decoded, err := enc.DecodeString(trimmed); err == nil {
			str := strings.TrimSpace(string(decoded))
			if isProxyURI(str) {
				return str
			}
		}
	}

	return trimmed
}

// Parse parses a proxy URL string into a ProxyNode.
// Supported schemes are: ss://, vmess://, vless://, and trojan://.
// If the proxyURL is base64-encoded, it will be automatically decoded.
func (p *Parser) Parse(proxyURL string) (*model.ProxyNode, error) {
	proxyURL = DecodeIfBase64(proxyURL)
	if strings.HasPrefix(proxyURL, "ss://") {
		return p.parseShadowsocks(proxyURL)
	} else if strings.HasPrefix(proxyURL, "vmess://") {
		return p.parseVMess(proxyURL)
	} else if strings.HasPrefix(proxyURL, "vless://") {
		return p.parseVLESS(proxyURL)
	} else if strings.HasPrefix(proxyURL, "trojan://") {
		return p.parseTrojan(proxyURL)
	}
	return nil, fmt.Errorf("unsupported proxy protocol: %s", proxyURL)
}

func (p *Parser) parseShadowsocks(proxyURL string) (*model.ProxyNode, error) {
	node := &model.ProxyNode{VPN: "shadowsocks"}
	content := strings.TrimPrefix(proxyURL, "ss://")

	parts := strings.SplitN(content, "#", 2)
	if len(parts) == 2 {
		node.Remark, _ = url.QueryUnescape(parts[1])
		content = parts[0]
	}

	var queryStr string
	if qIdx := strings.Index(content, "?"); qIdx != -1 {
		queryStr = content[qIdx+1:]
		content = content[:qIdx]
	}
	content = strings.TrimSuffix(content, "/")

	var methodPassword, serverPort string

	// Check if the URL has an @ outside base64 (e.g. ss://base64(userinfo)@host:port)
	if atIndex := strings.LastIndex(content, "@"); atIndex != -1 {
		userInfo := content[:atIndex]
		serverPort = content[atIndex+1:]

		// userInfo can be base64-encoded or raw
		if decoded, err := base64.RawURLEncoding.DecodeString(userInfo); err == nil {
			methodPassword = string(decoded)
		} else if decoded, err := base64.StdEncoding.DecodeString(userInfo); err == nil {
			methodPassword = string(decoded)
		} else {
			methodPassword = userInfo
		}
	} else {
		// Try decoding the whole content as base64 (e.g. ss://base64(userinfo@host:port))
		decoded, err := base64.RawURLEncoding.DecodeString(content)
		if err != nil {
			decoded, err = base64.StdEncoding.DecodeString(content)
			if err != nil {
				decoded = []byte(content)
			}
		}

		decodedStr := string(decoded)
		atIndex := strings.LastIndex(decodedStr, "@")
		if atIndex == -1 {
			return nil, fmt.Errorf("invalid shadowsocks URL format")
		}

		methodPassword = decodedStr[:atIndex]
		serverPort = decodedStr[atIndex+1:]
	}

	colonIndex := strings.Index(methodPassword, ":")
	if colonIndex == -1 {
		return nil, fmt.Errorf("invalid method:password format")
	}

	node.Method = methodPassword[:colonIndex]
	node.Password = methodPassword[colonIndex+1:]

	serverPortParts := strings.Split(serverPort, ":")
	if len(serverPortParts) != 2 {
		return nil, fmt.Errorf("invalid server:port format")
	}

	node.Server = serverPortParts[0]
	port, err := strconv.Atoi(serverPortParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}
	node.ServerPort = port

	if queryStr != "" {
		if q, err := url.ParseQuery(queryStr); err == nil {
			if pluginVal := q.Get("plugin"); pluginVal != "" {
				pluginParts := strings.SplitN(pluginVal, ";", 2)
				node.Plugin = pluginParts[0]
				if len(pluginParts) > 1 {
					node.PluginOpts = pluginParts[1]
				}
			}
		}
	}

	node.Raw = proxyURL

	return node, nil
}

func (p *Parser) parseVMess(proxyURL string) (*model.ProxyNode, error) {
	node := &model.ProxyNode{VPN: "vmess"}
	content := strings.TrimPrefix(proxyURL, "vmess://")

	decoded, err := base64.RawURLEncoding.DecodeString(content)
	if err != nil {
		decoded, err = base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode vmess base64: %w", err)
		}
	}

	var vmessConfig struct {
		Add  string `json:"add"`
		Port any    `json:"port"`
		ID   string `json:"id"`
		Aid  any    `json:"aid"`
		Net  string `json:"net"`
		Type string `json:"type"`
		Host string `json:"host"`
		Path string `json:"path"`
		TLS  string `json:"tls"`
		PS   string `json:"ps"`
		SNI  string `json:"sni"`
	}

	if err := json.Unmarshal(decoded, &vmessConfig); err != nil {
		return nil, fmt.Errorf("failed to parse vmess JSON: %w", err)
	}

	node.Server = vmessConfig.Add
	node.UUID = vmessConfig.ID
	node.Transport = vmessConfig.Net
	node.Host = vmessConfig.Host
	node.Path = vmessConfig.Path
	node.Remark = vmessConfig.PS
	node.SNI = vmessConfig.SNI

	if vmessConfig.TLS == "tls" {
		node.TLS = true
	}

	switch v := vmessConfig.Port.(type) {
	case string:
		port, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %w", err)
		}
		node.ServerPort = port
	case float64:
		node.ServerPort = int(v)
	case int:
		node.ServerPort = v
	}

	switch v := vmessConfig.Aid.(type) {
	case string:
		aid, err := strconv.Atoi(v)
		if err == nil {
			node.AlterID = aid
		}
	case float64:
		node.AlterID = int(v)
	case int:
		node.AlterID = v
	}

	node.Raw = proxyURL
	return node, nil
}

func (p *Parser) parseVLESS(proxyURL string) (*model.ProxyNode, error) {
	node := &model.ProxyNode{VPN: "vless"}
	content := strings.TrimPrefix(proxyURL, "vless://")

	parts := strings.SplitN(content, "#", 2)
	if len(parts) == 2 {
		node.Remark, _ = url.QueryUnescape(parts[1])
		content = parts[0]
	}

	parts = strings.SplitN(content, "?", 2)
	baseURL := parts[0]
	var params url.Values
	if len(parts) == 2 {
		params, _ = url.ParseQuery(parts[1])
	}

	atIndex := strings.Index(baseURL, "@")
	if atIndex == -1 {
		return nil, fmt.Errorf("invalid vless URL format")
	}

	node.UUID = baseURL[:atIndex]
	serverPort := baseURL[atIndex+1:]

	colonIndex := strings.LastIndex(serverPort, ":")
	if colonIndex == -1 {
		return nil, fmt.Errorf("invalid server:port format")
	}

	node.Server = serverPort[:colonIndex]
	port, err := strconv.Atoi(serverPort[colonIndex+1:])
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}
	node.ServerPort = port

	if params != nil {
		node.Transport = params.Get("type")
		node.Security = params.Get("security")
		node.Host = params.Get("host")
		node.Path = params.Get("path")
		node.SNI = params.Get("sni")
		node.ServiceName = params.Get("serviceName")

		if params.Get("security") == "tls" {
			node.TLS = true
		}
	}

	node.Raw = proxyURL
	return node, nil
}

func (p *Parser) parseTrojan(proxyURL string) (*model.ProxyNode, error) {
	node := &model.ProxyNode{VPN: "trojan"}
	content := strings.TrimPrefix(proxyURL, "trojan://")

	parts := strings.SplitN(content, "#", 2)
	if len(parts) == 2 {
		node.Remark, _ = url.QueryUnescape(parts[1])
		content = parts[0]
	}

	parts = strings.SplitN(content, "?", 2)
	baseURL := parts[0]
	var params url.Values
	if len(parts) == 2 {
		params, _ = url.ParseQuery(parts[1])
	}

	atIndex := strings.Index(baseURL, "@")
	if atIndex == -1 {
		return nil, fmt.Errorf("invalid trojan URL format")
	}

	unescapedPass, err := url.QueryUnescape(baseURL[:atIndex])
	if err == nil {
		node.Password = unescapedPass
	} else {
		node.Password = baseURL[:atIndex]
	}
	serverPort := baseURL[atIndex+1:]

	colonIndex := strings.LastIndex(serverPort, ":")
	if colonIndex == -1 {
		return nil, fmt.Errorf("invalid server:port format")
	}

	node.Server = serverPort[:colonIndex]
	port, err := strconv.Atoi(serverPort[colonIndex+1:])
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}
	node.ServerPort = port

	if params != nil {
		node.Transport = params.Get("type")
		node.Security = params.Get("security")
		node.Host = params.Get("host")
		node.Path = params.Get("path")
		node.SNI = params.Get("sni")

		if params.Get("security") == "tls" || params.Get("tls") == "1" {
			node.TLS = true
		}
	}

	node.Raw = proxyURL
	return node, nil
}
