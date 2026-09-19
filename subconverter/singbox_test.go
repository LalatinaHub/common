package subconverter

import (
	"encoding/json"
	"testing"

	"github.com/LalatinaHub/common/model"
)

func TestToSingboxMap_DirectDNS(t *testing.T) {
	s := &Subconverter{
		nodes: []model.ProxyNode{
			{
				VPN:        "vmess",
				Server:     "example.com",
				ServerPort: 443,
				UUID:       "test-uuid",
				Remark:     "Test Node",
			},
		},
	}

	config, err := s.ToSingboxMap("standard", "")
	if err != nil {
		t.Fatalf("ToSingboxMap failed: %v", err)
	}

	// Verify DNS configuration
	dnsConfig, ok := config["dns"].(map[string]any)
	if !ok {
		t.Fatal("DNS config not found")
	}

	servers, ok := dnsConfig["servers"].([]map[string]any)
	if !ok {
		t.Fatal("DNS servers not found")
	}

	// Check direct-dns server
	var directDNS map[string]any
	for _, server := range servers {
		if server["tag"] == "direct-dns" {
			directDNS = server
			break
		}
	}

	if directDNS == nil {
		t.Fatal("direct-dns server not found")
	}

	// Verify direct-dns uses proper format
	if dnsType, ok := directDNS["type"].(string); !ok || dnsType != "udp" {
		t.Errorf("direct-dns should have 'type': 'udp', got: %v", directDNS["type"])
	}

	if server, ok := directDNS["server"].(string); !ok || server == "" {
		t.Error("direct-dns should have 'server' field")
	}

	// Verify direct-dns does NOT have detour field (uses direct connection by default)
	if _, hasDetour := directDNS["detour"]; hasDetour {
		t.Error("direct-dns should not have 'detour' field (uses direct connection by default)")
	}
}

func TestToSingboxMap_CFTemplate(t *testing.T) {
	s := &Subconverter{
		nodes: []model.ProxyNode{
			{
				VPN:        "vmess",
				Server:     "example.com",
				ServerPort: 443,
				UUID:       "test-uuid",
				Remark:     "Test Node",
			},
		},
	}

	config, err := s.ToSingboxMap("standard", "cf")
	if err != nil {
		t.Fatalf("ToSingboxMap with CF template failed: %v", err)
	}

	// Verify DNS configuration
	dnsConfig, ok := config["dns"].(map[string]any)
	if !ok {
		t.Fatal("DNS config not found")
	}

	servers, ok := dnsConfig["servers"].([]map[string]any)
	if !ok {
		t.Fatal("DNS servers not found")
	}

	// Check both DNS servers
	dnsServers := make(map[string]interface{})
	for _, server := range servers {
		tag := server["tag"].(string)
		detour, hasDetour := server["detour"]
		if hasDetour {
			dnsServers[tag] = detour.(string)
		} else {
			dnsServers[tag] = nil
		}
	}

	// remote-dns should use Trojan UDP
	if dnsServers["remote-dns"] != "Trojan UDP" {
		t.Errorf("remote-dns should use 'Trojan UDP', got: %v", dnsServers["remote-dns"])
	}

	// direct-dns should not have detour (uses direct connection by default)
	if dnsServers["direct-dns"] != nil {
		t.Errorf("direct-dns should not have detour field, got: %v", dnsServers["direct-dns"])
	}

	// Verify Trojan UDP outbound exists
	outbounds, ok := config["outbounds"].([]map[string]any)
	if !ok {
		t.Fatal("Outbounds not found")
	}

	var trojanUDP map[string]any
	for _, outbound := range outbounds {
		if outbound["tag"] == "Trojan UDP" {
			trojanUDP = outbound
			break
		}
	}

	if trojanUDP == nil {
		t.Fatal("Trojan UDP outbound not found")
	}
}

func TestToSingbox_JSONFormat(t *testing.T) {
	s := &Subconverter{
		nodes: []model.ProxyNode{
			{
				VPN:        "vmess",
				Server:     "example.com",
				ServerPort: 443,
				UUID:       "test-uuid",
				Remark:     "Test Node",
			},
		},
	}

	jsonStr, err := s.ToSingbox("standard", "")
	if err != nil {
		t.Fatalf("ToSingbox failed: %v", err)
	}

	// Verify it's valid JSON
	var result map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Generated config is not valid JSON: %v", err)
	}

	t.Logf("Generated config:\n%s", jsonStr)
}
