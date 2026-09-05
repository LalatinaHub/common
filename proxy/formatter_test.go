package proxy_test

import (
	"testing"

	"github.com/LalatinaHub/common/model"
	"github.com/LalatinaHub/common/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatter_FormatShadowsocks(t *testing.T) {
	node := &model.ProxyNode{
		VPN:        "shadowsocks",
		Server:     "ss.example.com",
		ServerPort: 8388,
		Method:     "aes-128-gcm",
		Password:   "secretpass",
		Remark:     "SS-Node-1",
		Plugin:     "obfs-local",
		PluginOpts: "obfs=tls;obfs-host=example.com",
	}

	rawURL, err := proxy.FormatString(node)
	require.NoError(t, err)
	assert.Contains(t, rawURL, "ss://")
	assert.Contains(t, rawURL, "@ss.example.com:8388")
	assert.Contains(t, rawURL, "#SS-Node-1")
	assert.Contains(t, rawURL, "plugin=obfs-local%3Bobfs%3Dtls%3Bobfs-host%3Dexample.com")

	// Round-trip test
	parser := proxy.NewParser()
	parsedNode, err := parser.Parse(rawURL)
	require.NoError(t, err)
	assert.Equal(t, node.Server, parsedNode.Server)
	assert.Equal(t, node.ServerPort, parsedNode.ServerPort)
	assert.Equal(t, node.Method, parsedNode.Method)
	assert.Equal(t, node.Password, parsedNode.Password)
	assert.Equal(t, node.Remark, parsedNode.Remark)
}

func TestFormatter_FormatVMess(t *testing.T) {
	node := &model.ProxyNode{
		VPN:        "vmess",
		Server:     "vmess.example.com",
		ServerPort: 443,
		UUID:       "11111111-2222-3333-4444-555555555555",
		AlterID:    0,
		Security:   "auto",
		Transport:  "ws",
		Host:       "cdn.example.com",
		Path:       "/vmess-ws",
		TLS:        true,
		SNI:        "sni.example.com",
		Remark:     "VMess-SG",
	}

	rawURL, err := proxy.FormatString(node)
	require.NoError(t, err)
	assert.Contains(t, rawURL, "vmess://")

	// Round-trip test
	parser := proxy.NewParser()
	parsedNode, err := parser.Parse(rawURL)
	require.NoError(t, err)
	assert.Equal(t, node.Server, parsedNode.Server)
	assert.Equal(t, node.ServerPort, parsedNode.ServerPort)
	assert.Equal(t, node.UUID, parsedNode.UUID)
	assert.Equal(t, node.Transport, parsedNode.Transport)
	assert.Equal(t, node.Host, parsedNode.Host)
	assert.Equal(t, node.Path, parsedNode.Path)
	assert.Equal(t, node.TLS, parsedNode.TLS)
	assert.Equal(t, node.SNI, parsedNode.SNI)
	assert.Equal(t, node.Remark, parsedNode.Remark)
}

func TestFormatter_FormatVLESS(t *testing.T) {
	node := &model.ProxyNode{
		VPN:         "vless",
		Server:      "vless.example.com",
		ServerPort:  443,
		UUID:        "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		Transport:   "grpc",
		Security:    "tls",
		TLS:         true,
		Host:        "host.example.com",
		Path:        "/vless-path",
		SNI:         "vless.example.com",
		ServiceName: "vless-grpc",
		Remark:      "VLESS-ID",
	}

	rawURL, err := proxy.FormatString(node)
	require.NoError(t, err)
	assert.Contains(t, rawURL, "vless://aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee@vless.example.com:443")
	assert.Contains(t, rawURL, "#VLESS-ID")

	// Round-trip test
	parser := proxy.NewParser()
	parsedNode, err := parser.Parse(rawURL)
	require.NoError(t, err)
	assert.Equal(t, node.Server, parsedNode.Server)
	assert.Equal(t, node.ServerPort, parsedNode.ServerPort)
	assert.Equal(t, node.UUID, parsedNode.UUID)
	assert.Equal(t, node.Transport, parsedNode.Transport)
	assert.Equal(t, node.Security, parsedNode.Security)
	assert.Equal(t, node.TLS, parsedNode.TLS)
	assert.Equal(t, node.Host, parsedNode.Host)
	assert.Equal(t, node.Path, parsedNode.Path)
	assert.Equal(t, node.SNI, parsedNode.SNI)
	assert.Equal(t, node.ServiceName, parsedNode.ServiceName)
	assert.Equal(t, node.Remark, parsedNode.Remark)
}

func TestFormatter_FormatTrojan(t *testing.T) {
	node := &model.ProxyNode{
		VPN:        "trojan",
		Server:     "trojan.example.com",
		ServerPort: 443,
		Password:   "p@ssw0rd!#",
		Transport:  "ws",
		Security:   "tls",
		TLS:        true,
		Host:       "trojan.example.com",
		Path:       "/trojan-ws",
		SNI:        "trojan.example.com",
		Remark:     "Trojan-Premium",
	}

	rawURL, err := proxy.FormatString(node)
	require.NoError(t, err)
	assert.Contains(t, rawURL, "trojan://")
	assert.Contains(t, rawURL, "@trojan.example.com:443")
	assert.Contains(t, rawURL, "#Trojan-Premium")

	// Round-trip test
	parser := proxy.NewParser()
	parsedNode, err := parser.Parse(rawURL)
	require.NoError(t, err)
	assert.Equal(t, node.Server, parsedNode.Server)
	assert.Equal(t, node.ServerPort, parsedNode.ServerPort)
	assert.Equal(t, node.Password, parsedNode.Password)
	assert.Equal(t, node.Transport, parsedNode.Transport)
	assert.Equal(t, node.Security, parsedNode.Security)
	assert.Equal(t, node.TLS, parsedNode.TLS)
	assert.Equal(t, node.Host, parsedNode.Host)
	assert.Equal(t, node.Path, parsedNode.Path)
	assert.Equal(t, node.SNI, parsedNode.SNI)
	assert.Equal(t, node.Remark, parsedNode.Remark)
}

func TestFormatter_ValidationErrors(t *testing.T) {
	t.Run("nil node", func(t *testing.T) {
		_, err := proxy.Format(nil)
		assert.Error(t, err)
	})

	t.Run("unsupported protocol", func(t *testing.T) {
		node := &model.ProxyNode{
			VPN:    "wireguard",
			Server: "1.1.1.1",
		}
		_, err := proxy.Format(node)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported proxy protocol")
	})

	t.Run("empty server", func(t *testing.T) {
		node := &model.ProxyNode{
			VPN: "trojan",
		}
		_, err := proxy.Format(node)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server host is empty")
	})
}
