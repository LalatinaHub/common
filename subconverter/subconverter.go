package subconverter

import (
	"errors"
	"strings"

	"github.com/LalatinaHub/common/model"
	"github.com/LalatinaHub/common/proxy"
)

// DefaultUDPAccount is the predefined Trojan UDP relay node used for Cloudflare UDP detours.
const DefaultUDPAccount = "trojan://t.me%2Ffoolvpn@172.67.73.39:443?path=%2Ftrojan-udp&security=tls&host=id1.foolvpn.me&type=ws&sni=id1.foolvpn.me#Trojan%20UDP"

// Subconverter transforms a collection of ProxyNode objects into various client configurations
// (Clash Meta YAML, sing-box JSON, SFA/BFR profiles, Raw URLs, Base64).
type Subconverter struct {
	nodes  []model.ProxyNode
	parser *proxy.Parser
}

// New creates a new Subconverter with a slice of ProxyNode.
func New(nodes []model.ProxyNode) *Subconverter {
	return &Subconverter{
		nodes:  nodes,
		parser: proxy.NewParser(),
	}
}

// NewFromRaw parses a raw configuration string (newline or comma-separated proxy URLs)
// and constructs a Subconverter instance.
func NewFromRaw(rawConfig string) (*Subconverter, error) {
	parser := proxy.NewParser()
	cleaned := strings.ReplaceAll(rawConfig, ",", "\n")
	lines := strings.Split(cleaned, "\n")

	var nodes []model.ProxyNode
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}

		node, err := parser.Parse(trimmed)
		if err == nil && node != nil {
			nodes = append(nodes, *node)
		}
	}

	if len(nodes) == 0 {
		return nil, errors.New("no valid proxy configurations found in input")
	}

	return &Subconverter{
		nodes:  nodes,
		parser: parser,
	}, nil
}

// Nodes returns the underlying slice of ProxyNode entities.
func (s *Subconverter) Nodes() []model.ProxyNode {
	return s.nodes
}

// NodeCount returns the number of proxy nodes.
func (s *Subconverter) NodeCount() int {
	return len(s.nodes)
}
