package subconverter

import (
	"fmt"
	"strings"

	"github.com/LalatinaHub/common/model"
	"gopkg.in/yaml.v3"
)

// ToClashMap converts the internal proxy nodes into a Clash / Mihomo configuration map.
// If template is "cf", Cloudflare Trojan UDP relay configuration will be injected.
func (s *Subconverter) ToClashMap(template string) (map[string]any, error) {
	proxies := make([]map[string]any, 0, len(s.nodes))
	proxyNames := make([]string, 0, len(s.nodes))

	for i, node := range s.nodes {
		pMap := nodeToClashProxy(&node, i+1)
		proxies = append(proxies, pMap)
		proxyNames = append(proxyNames, pMap["name"].(string))
	}

	proxyGroups := []map[string]any{
		{
			"name": "PROXIES",
			"type": "select",
			"proxies": append([]string{
				"AUTO-FALLBACK",
				"LOAD-BALANCE",
			}, proxyNames...),
		},
		{
			"name":      "AUTO-FALLBACK",
			"type":      "url-test",
			"url":       "http://cp.cloudflare.com/generate_204",
			"interval":  300,
			"tolerance": 50,
			"proxies":   proxyNames,
		},
		{
			"name":      "LOAD-BALANCE",
			"type":      "load-balance",
			"strategy":  "round-robin",
			"url":       "http://cp.cloudflare.com/generate_204",
			"interval":  300,
			"proxies":   proxyNames,
		},
	}

	rules := []string{
		"GEOIP,LAN,DIRECT,no-resolve",
		"MATCH,PROXIES",
	}

	// Post-processing template "cf"
	if template == "cf" && len(proxies) > 0 {
		lastServer := proxies[len(proxies)-1]["server"].(string)
		udpProxy := map[string]any{
			"name":             "Trojan UDP",
			"type":             "trojan",
			"server":           lastServer,
			"port":             443,
			"password":         "t.me/foolvpn",
			"sni":              "id1.foolvpn.me",
			"skip-cert-verify": true,
			"network":          "ws",
			"ws-opts": map[string]any{
				"path": "/trojan-udp",
				"headers": map[string]any{
					"Host": "id1.foolvpn.me",
				},
			},
			"udp": true,
		}
		proxies = append(proxies, udpProxy)

		// Insert UDP rule before final MATCH rule
		udpRule := "NETWORK,UDP,Trojan UDP"
		rules = append([]string{udpRule}, rules...)
	}

	config := map[string]any{
		"port":                7890,
		"socks-port":          7891,
		"allow-lan":           false,
		"mode":                "rule",
		"log-level":           "info",
		"external-controller": "127.0.0.1:9090",
		"proxies":             proxies,
		"proxy-groups":        proxyGroups,
		"rules":               rules,
	}

	return config, nil
}

// ToClash serializes the proxy nodes into a valid Clash / Mihomo YAML string.
func (s *Subconverter) ToClash(template string) (string, error) {
	config, err := s.ToClashMap(template)
	if err != nil {
		return "", err
	}

	bytes, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Clash YAML: %w", err)
	}

	return string(bytes), nil
}

func nodeToClashProxy(node *model.ProxyNode, index int) map[string]any {
	name := strings.TrimSpace(node.Remark)
	if name == "" {
		name = fmt.Sprintf("%s-%d", node.VPN, index)
	}

	p := map[string]any{
		"name":   name,
		"server": node.Server,
		"port":   node.ServerPort,
		"udp":    true,
	}

	switch strings.ToLower(node.VPN) {
	case "shadowsocks", "ss":
		p["type"] = "ss"
		cipher := node.Method
		if cipher == "" {
			cipher = "aes-256-gcm"
		}
		p["cipher"] = cipher
		p["password"] = node.Password
		if node.Plugin != "" {
			p["plugin"] = node.Plugin
			if node.PluginOpts != "" {
				p["plugin-opts"] = parsePluginOpts(node.PluginOpts)
			}
		}

	case "vmess":
		p["type"] = "vmess"
		p["uuid"] = node.UUID
		p["alterId"] = node.AlterID
		cipher := node.Security
		if cipher == "" {
			cipher = "auto"
		}
		p["cipher"] = cipher
		p["tls"] = node.TLS
		p["skip-cert-verify"] = true
		if node.SNI != "" {
			p["servername"] = node.SNI
		} else if node.Host != "" {
			p["servername"] = node.Host
		}

		applyTransportClash(p, node)

	case "vless":
		p["type"] = "vless"
		p["uuid"] = node.UUID
		p["tls"] = node.TLS
		p["skip-cert-verify"] = true
		if node.SNI != "" {
			p["servername"] = node.SNI
		} else if node.Host != "" {
			p["servername"] = node.Host
		}

		applyTransportClash(p, node)

	case "trojan":
		p["type"] = "trojan"
		p["password"] = node.Password
		p["skip-cert-verify"] = true
		if node.SNI != "" {
			p["sni"] = node.SNI
		} else if node.Host != "" {
			p["sni"] = node.Host
		}

		applyTransportClash(p, node)
	}

	return p
}

func applyTransportClash(p map[string]any, node *model.ProxyNode) {
	transport := strings.ToLower(node.Transport)
	if transport == "" {
		transport = "tcp"
	}
	p["network"] = transport

	switch transport {
	case "ws":
		wsOpts := map[string]any{}
		if node.Path != "" {
			wsOpts["path"] = node.Path
		}
		if node.Host != "" {
			wsOpts["headers"] = map[string]any{
				"Host": node.Host,
			}
		}
		if len(wsOpts) > 0 {
			p["ws-opts"] = wsOpts
		}
	case "grpc":
		if node.ServiceName != "" {
			p["grpc-opts"] = map[string]any{
				"grpc-service-name": node.ServiceName,
			}
		}
	}
}

func parsePluginOpts(opts string) map[string]any {
	result := make(map[string]any)
	pairs := strings.Split(opts, ";")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		} else if len(kv) == 1 && kv[0] != "" {
			result[kv[0]] = true
		}
	}
	return result
}
