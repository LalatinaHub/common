package subconverter

import (
	"encoding/base64"
	"strings"

	"github.com/LalatinaHub/common/proxy"
)

// ToRaw returns a slice of proxy URLs as strings.
func (s *Subconverter) ToRaw() []string {
	var lines []string
	for i := range s.nodes {
		node := &s.nodes[i]
		if node.Raw != "" {
			lines = append(lines, node.Raw)
			continue
		}

		urlStr, err := proxy.FormatString(node)
		if err == nil && urlStr != "" {
			lines = append(lines, urlStr)
		}
	}
	return lines
}

// ToRawString returns proxy URLs joined by newlines.
func (s *Subconverter) ToRawString() string {
	return strings.Join(s.ToRaw(), "\n")
}

// ToBase64 returns standard Base64-encoded string of the raw lines,
// formatted for mobile clients like v2rayNG, Shadowrocket, and NekoBox.
func (s *Subconverter) ToBase64() string {
	raw := s.ToRawString()
	return base64.StdEncoding.EncodeToString([]byte(raw))
}
