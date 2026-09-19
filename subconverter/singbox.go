package subconverter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LalatinaHub/common/model"
)

// ToSingboxMap converts the internal proxy nodes into a complete sing-box configuration map.
// Profile can be "standard", "sfa" (Sing-Box For Android), or "bfr".
// If template is "cf", Cloudflare Trojan UDP relay configuration will be injected.
func (s *Subconverter) ToSingboxMap(profile string, template string) (map[string]any, error) {
	nodeOutbounds := make([]map[string]any, 0, len(s.nodes))
	nodeTags := make([]string, 0, len(s.nodes))

	for i, node := range s.nodes {
		ob := nodeToSingboxOutbound(&node, i+1)
		nodeOutbounds = append(nodeOutbounds, ob)
		nodeTags = append(nodeTags, ob["tag"].(string))
	}

	var outbounds []map[string]any

	switch profile {
	case "sfa", "bfr":
		// SFA / BFR Profile matching legacy mobile sing-box client configurations
		outbounds = append(outbounds,
			map[string]any{
				"type":      "selector",
				"tag":       "Internet",
				"outbounds": append([]string{"Best Latency"}, nodeTags...),
				"default":   "Best Latency",
			},
			map[string]any{
				"type":      "selector",
				"tag":       "Lock Region ID",
				"outbounds": append([]string{"Best Latency"}, nodeTags...),
			},
			map[string]any{
				"type":      "urltest",
				"tag":       "Best Latency",
				"outbounds": nodeTags,
				"url":       "https://cp.cloudflare.com/generate_204",
				"interval":  "3m",
				"tolerance": 50,
			},
		)
	default:
		// Standard modern sing-box v1.14+ profile
		outbounds = append(outbounds,
			map[string]any{
				"type":      "selector",
				"tag":       "select",
				"outbounds": append([]string{"auto"}, nodeTags...),
				"default":   "auto",
			},
			map[string]any{
				"type":      "urltest",
				"tag":       "auto",
				"outbounds": nodeTags,
				"url":       "https://cp.cloudflare.com/generate_204",
				"interval":  "3m",
				"tolerance": 50,
			},
		)
	}

	// Append individual node outbounds
	outbounds = append(outbounds, nodeOutbounds...)

	// Append standard system outbounds
	outbounds = append(outbounds,
		map[string]any{
			"type": "direct",
			"tag":  "direct",
		},
		map[string]any{
			"type": "block",
			"tag":  "block",
		},
	)

	dnsServers := []map[string]any{
		{
			"type":   "udp",
			"tag":    "remote-dns",
			"server": "1.1.1.1",
			"detour": "select",
		},
		{
			"type":   "udp",
			"tag":    "direct-dns",
			"server": "223.5.5.5",
			"detour": "direct",
		},
	}

	routeRules := []map[string]any{
		{
			"action": "sniff",
		},
		{
			"type": "logical",
			"mode": "or",
			"rules": []map[string]any{
				{"protocol": "dns"},
				{"port": 53},
			},
			"action": "hijack-dns",
		},
		{
			"action":        "route",
			"outbound":      "direct",
			"ip_is_private": true,
		},
	}

	// Post-processing template "cf"
	if template == "cf" && len(nodeOutbounds) > 0 {
		lastServer := nodeOutbounds[len(nodeOutbounds)-1]["server"].(string)
		udpOutbound := map[string]any{
			"type":        "trojan",
			"tag":         "Trojan UDP",
			"server":      lastServer,
			"server_port": 443,
			"password":    "t.me/foolvpn",
			"tls": map[string]any{
				"enabled":     true,
				"server_name": "id1.foolvpn.me",
				"insecure":    true,
			},
			"transport": map[string]any{
				"type": "ws",
				"path": "/trojan-udp",
				"headers": map[string]any{
					"Host": "id1.foolvpn.me",
				},
			},
		}
		outbounds = append(outbounds, udpOutbound)

		// Set DNS detour to Trojan UDP for remote-dns only
		for i := range dnsServers {
			if dnsServers[i]["tag"] == "remote-dns" {
				dnsServers[i]["detour"] = "Trojan UDP"
			}
		}

		// Insert UDP route rule
		routeRules = append(routeRules, map[string]any{
			"action":   "route",
			"network":  []string{"udp"},
			"outbound": "Trojan UDP",
		})
	}

	config := map[string]any{
		"log": map[string]any{
			"disabled": false,
			"level":    "warn",
		},
		"dns": map[string]any{
			"servers": dnsServers,
			"final":   "remote-dns",
		},
		"inbounds": []map[string]any{
			{
				"type":        "mixed",
				"tag":         "mixed-in",
				"listen":      "127.0.0.1",
				"listen_port": 2080,
			},
		},
		"outbounds": outbounds,
		"route": map[string]any{
			"rules":                 routeRules,
			"final":                 "select",
			"auto_detect_interface": true,
		},
	}

	return config, nil
}

// ToSingbox serializes the configuration to a JSON string.
func (s *Subconverter) ToSingbox(profile string, template string) (string, error) {
	config, err := s.ToSingboxMap(profile, template)
	if err != nil {
		return "", err
	}

	bytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal sing-box JSON: %w", err)
	}

	return string(bytes), nil
}

// ToSFA generates Sing-Box For Android (SFA) JSON configuration string.
func (s *Subconverter) ToSFA(template string) (string, error) {
	return s.ToSingbox("sfa", template)
}

// ToBFR generates Box For Root (BFR) JSON configuration string.
func (s *Subconverter) ToBFR(template string) (string, error) {
	return s.ToSingbox("bfr", template)
}

func nodeToSingboxOutbound(node *model.ProxyNode, index int) map[string]any {
	tag := strings.TrimSpace(node.Remark)
	if tag == "" {
		tag = fmt.Sprintf("%s-%d", node.VPN, index)
	}

	ob := map[string]any{
		"tag":         tag,
		"server":      node.Server,
		"server_port": node.ServerPort,
	}

	switch strings.ToLower(node.VPN) {
	case "shadowsocks", "ss":
		ob["type"] = "shadowsocks"
		method := node.Method
		if method == "" {
			method = "aes-256-gcm"
		}
		ob["method"] = method
		ob["password"] = node.Password
		if node.Plugin != "" {
			ob["plugin"] = node.Plugin
			if node.PluginOpts != "" {
				ob["plugin_opts"] = node.PluginOpts
			}
		}

	case "vmess":
		ob["type"] = "vmess"
		ob["uuid"] = node.UUID
		security := node.Security
		if security == "" {
			security = "auto"
		}
		ob["security"] = security
		ob["alter_id"] = node.AlterID

		applySingboxTLS(ob, node)
		applySingboxTransport(ob, node)

	case "vless":
		ob["type"] = "vless"
		ob["uuid"] = node.UUID

		applySingboxTLS(ob, node)
		applySingboxTransport(ob, node)

	case "trojan":
		ob["type"] = "trojan"
		ob["password"] = node.Password

		applySingboxTLS(ob, node)
		applySingboxTransport(ob, node)
	}

	return ob
}

func applySingboxTLS(ob map[string]any, node *model.ProxyNode) {
	if node.TLS {
		serverName := node.SNI
		if serverName == "" {
			serverName = node.Host
		}
		if serverName == "" {
			serverName = node.Server
		}

		ob["tls"] = map[string]any{
			"enabled":     true,
			"server_name": serverName,
			"insecure":    true,
		}
	}
}

func applySingboxTransport(ob map[string]any, node *model.ProxyNode) {
	transport := strings.ToLower(node.Transport)
	switch transport {
	case "ws":
		headers := map[string]any{}
		if node.Host != "" {
			headers["Host"] = node.Host
		}
		ob["transport"] = map[string]any{
			"type":    "ws",
			"path":    node.Path,
			"headers": headers,
		}
	case "grpc":
		ob["transport"] = map[string]any{
			"type":         "grpc",
			"service_name": node.ServiceName,
		}
	}
}
