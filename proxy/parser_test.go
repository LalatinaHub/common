package proxy_test

import (
	"testing"

	"github.com/LalatinaHub/common/model"
	"github.com/LalatinaHub/common/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_ParseShadowsocks(t *testing.T) {
	parser := proxy.NewParser()

	tests := []struct {
		name    string
		url     string
		wantErr bool
		check   func(*testing.T, *model.ProxyNode)
	}{
		{
			name:    "valid shadowsocks with base64 full URL",
			url:     "ss://YWVzLTEyOC1nY206dGVzdHBhc3N3b3JkQGV4YW1wbGUuY29tOjgzODg=#TestServer",
			wantErr: false,
			check: func(t *testing.T, node *model.ProxyNode) {
				assert.Equal(t, "aes-128-gcm", node.Method)
				assert.Equal(t, "testpassword", node.Password)
				assert.Equal(t, "example.com", node.Server)
				assert.Equal(t, 8388, node.ServerPort)
				assert.Equal(t, "TestServer", node.Remark)
				assert.Equal(t, "shadowsocks", node.VPN)
			},
		},
		{
			name:    "valid shadowsocks with base64 userinfo only",
			url:     "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQ=@example.com:8388#TestServer2",
			wantErr: false,
			check: func(t *testing.T, node *model.ProxyNode) {
				assert.Equal(t, "aes-256-gcm", node.Method)
				assert.Equal(t, "password", node.Password)
				assert.Equal(t, "example.com", node.Server)
				assert.Equal(t, 8388, node.ServerPort)
				assert.Equal(t, "TestServer2", node.Remark)
			},
		},
		{
			name:    "valid shadowsocks without base64",
			url:     "ss://aes-256-cfb:password123@192.168.1.1:8388",
			wantErr: false,
			check: func(t *testing.T, node *model.ProxyNode) {
				assert.Equal(t, "aes-256-cfb", node.Method)
				assert.Equal(t, "password123", node.Password)
				assert.Equal(t, "192.168.1.1", node.Server)
				assert.Equal(t, 8388, node.ServerPort)
			},
		},
		{
			name:    "invalid base64 content",
			url:     "ss://invalid_base64_content",
			wantErr: true,
		},
		{
			name:    "invalid port",
			url:     "ss://aes-256-cfb:pass@example.com:notaport",
			wantErr: true,
		},
		{
			name:    "invalid format no colon",
			url:     "ss://YWVzLTEyOC1nY21leGFtcGxlLmNvbTo4Mzg4",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := parser.Parse(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, node)
				if tt.check != nil {
					tt.check(t, node)
				}
			}
		})
	}
}

func TestParser_ParseVMess(t *testing.T) {
	parser := proxy.NewParser()

	// {"v":"2","ps":"VMessTest","add":"vmess.example.com","port":443,"id":"uuid-1234","aid":0,"net":"ws","type":"none","host":"host.com","path":"/vmess","tls":"tls","sni":"sni.com"}
	validVMessBase64 := "eyd2JzogJzInLCAncHMnOiAnVk1lc3NUZXN0JywgJ2FkZCc6ICd2bWVzcy5leGFtcGxlLmNvbScsICdwb3J0JzogNDQzLCAnaWQnOiAndXVpZC0xMjM0JywgJ2FpZCc6IDAsICduZXQnOiAnd3MnLCAndHlwZSc6ICdub25lJywgJ2hvc3QnOiAnaG9zdC5jb20nLCAncGF0aCc6ICcvdm1lc3MnLCAndGxzJzogJ3RscycsICdzbmknOiAnc25pLmNvbSd9"

	validVMessStd := "eyJ2IjoiMiIsInBzIjoiVk1lc3NUZXN0IiwiYWRkIjoidm1lc3MuZXhhbXBsZS5jb20iLCJwb3J0IjoiNDQzIiwiaWQiOiJ1dWlkLTEyMzQiLCJhaWQiOiIwIiwibmV0Ijoid3MiLCJ0eXBlIjoibm9uZSIsImhvc3QiOiJob3N0LmNvbSIsInBhdGgiOiIvdm1lc3MiLCJ0bHMiOiJ0bHMiLCJzbmkiOiJzbmkuY29tIn0="

	tests := []struct {
		name    string
		url     string
		wantErr bool
		check   func(*testing.T, *model.ProxyNode)
	}{
		{
			name:    "valid vmess standard base64",
			url:     "vmess://" + validVMessStd,
			wantErr: false,
			check: func(t *testing.T, node *model.ProxyNode) {
				assert.Equal(t, "vmess", node.VPN)
				assert.Equal(t, "vmess.example.com", node.Server)
				assert.Equal(t, 443, node.ServerPort)
				assert.Equal(t, "uuid-1234", node.UUID)
				assert.Equal(t, 0, node.AlterID)
				assert.Equal(t, "ws", node.Transport)
				assert.Equal(t, "host.com", node.Host)
				assert.Equal(t, "/vmess", node.Path)
				assert.True(t, node.TLS)
				assert.Equal(t, "sni.com", node.SNI)
				assert.Equal(t, "VMessTest", node.Remark)
			},
		},
		{
			name:    "invalid base64",
			url:     "vmess://!!!notbase64",
			wantErr: true,
		},
		{
			name:    "invalid json in base64",
			url:     "vmess://bm90IGEganNvbg==",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := parser.Parse(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, node)
				if tt.check != nil {
					tt.check(t, node)
				}
			}
		})
	}
	_ = validVMessBase64
}

func TestParser_ParseVLESS(t *testing.T) {
	parser := proxy.NewParser()

	tests := []struct {
		name    string
		url     string
		wantErr bool
		check   func(*testing.T, *model.ProxyNode)
	}{
		{
			name:    "valid vless with all params",
			url:     "vless://uuid-5678@vless.example.com:443?type=ws&security=tls&host=example.com&path=%2Fvless&sni=vless.example.com&serviceName=grpc-srv#VLESSTest",
			wantErr: false,
			check: func(t *testing.T, node *model.ProxyNode) {
				assert.Equal(t, "vless", node.VPN)
				assert.Equal(t, "uuid-5678", node.UUID)
				assert.Equal(t, "vless.example.com", node.Server)
				assert.Equal(t, 443, node.ServerPort)
				assert.Equal(t, "ws", node.Transport)
				assert.Equal(t, "tls", node.Security)
				assert.True(t, node.TLS)
				assert.Equal(t, "example.com", node.Host)
				assert.Equal(t, "/vless", node.Path)
				assert.Equal(t, "vless.example.com", node.SNI)
				assert.Equal(t, "grpc-srv", node.ServiceName)
				assert.Equal(t, "VLESSTest", node.Remark)
			},
		},
		{
			name:    "invalid vless no at",
			url:     "vless://no-at-sign.com:443",
			wantErr: true,
		},
		{
			name:    "invalid vless bad port",
			url:     "vless://uuid@host:notaport",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := parser.Parse(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, node)
				if tt.check != nil {
					tt.check(t, node)
				}
			}
		})
	}
}

func TestParser_ParseTrojan(t *testing.T) {
	parser := proxy.NewParser()

	tests := []struct {
		name    string
		url     string
		wantErr bool
		check   func(*testing.T, *model.ProxyNode)
	}{
		{
			name:    "valid trojan with all params",
			url:     "trojan://trojan-pass@trojan.example.com:443?type=ws&security=tls&host=example.com&path=%2Ftrojan&sni=trojan.example.com#TrojanTest",
			wantErr: false,
			check: func(t *testing.T, node *model.ProxyNode) {
				assert.Equal(t, "trojan", node.VPN)
				assert.Equal(t, "trojan-pass", node.Password)
				assert.Equal(t, "trojan.example.com", node.Server)
				assert.Equal(t, 443, node.ServerPort)
				assert.Equal(t, "ws", node.Transport)
				assert.Equal(t, "tls", node.Security)
				assert.True(t, node.TLS)
				assert.Equal(t, "example.com", node.Host)
				assert.Equal(t, "/trojan", node.Path)
				assert.Equal(t, "trojan.example.com", node.SNI)
				assert.Equal(t, "TrojanTest", node.Remark)
			},
		},
		{
			name:    "invalid trojan no at",
			url:     "trojan://no-at-sign.com:443",
			wantErr: true,
		},
		{
			name:    "invalid trojan bad port",
			url:     "trojan://pass@host:notaport",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := parser.Parse(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, node)
				if tt.check != nil {
					tt.check(t, node)
				}
			}
		})
	}
}

func TestParser_UnsupportedProtocol(t *testing.T) {
	parser := proxy.NewParser()
	node, err := parser.Parse("http://example.com:8080")
	assert.Error(t, err)
	assert.Nil(t, node)
	assert.Contains(t, err.Error(), "unsupported proxy protocol")
}

func TestParser_ParseBase64EncodedURL(t *testing.T) {
	parser := proxy.NewParser()

	// Base64 encoded vless URL (from real-world example)
	b64 := "dmxlc3M6Ly8zMGE4OGM0MC01ODFlLTQ5NTItYWE0Yy04YWY5NTY4NmVkZTBAd3d3Lmdvdi51YTo4ODgwP2VuY3J5cHRpb249bm9uZSZzZWN1cml0eT1ub25lJnR5cGU9d3MmaG9zdD1yYXBpZC1sYWItOTVlZi4xNzMtNzRjLndvcmtlcnMuZGV2JnBhdGg9L3B5aXA9cHJveHlpcC5rci5jbWxpdXNzc3MubmV0I0BEZWx0YUtyb25lY2tlckdpdGh1Yg=="
	node, err := parser.Parse(b64)
	require.NoError(t, err)
	require.NotNil(t, node)
	assert.Equal(t, "vless", node.VPN)
	assert.Equal(t, "www.gov.ua", node.Server)
	assert.Equal(t, 8880, node.ServerPort)
	assert.Equal(t, "30a88c40-581e-4952-aa4c-8af95686ede0", node.UUID)
	assert.Equal(t, "@DeltaKroneckerGithub", node.Remark)
	assert.Contains(t, node.Raw, "vless://")
	assert.NotContains(t, node.Raw, "dmxlc3M6")
}
