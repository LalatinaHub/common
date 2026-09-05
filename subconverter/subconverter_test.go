package subconverter_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/LalatinaHub/common/model"
	"github.com/LalatinaHub/common/subconverter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func sampleNodes() []model.ProxyNode {
	return []model.ProxyNode{
		{
			VPN:        "trojan",
			Server:     "tr.example.com",
			ServerPort: 443,
			Password:   "trojanpass",
			Transport:  "ws",
			TLS:        true,
			Host:       "tr.example.com",
			Path:       "/trojan-ws",
			SNI:        "tr.example.com",
			Remark:     "Trojan-SG",
		},
		{
			VPN:         "vless",
			Server:      "vl.example.com",
			ServerPort:  443,
			UUID:        "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			Transport:   "grpc",
			TLS:         true,
			Host:        "vl.example.com",
			SNI:         "vl.example.com",
			ServiceName: "vless-grpc",
			Remark:      "VLESS-ID",
		},
		{
			VPN:        "vmess",
			Server:     "vm.example.com",
			ServerPort: 443,
			UUID:       "11111111-2222-3333-4444-555555555555",
			Transport:  "ws",
			TLS:        true,
			Host:       "vm.example.com",
			Path:       "/vmess-ws",
			Remark:     "VMess-US",
		},
		{
			VPN:        "shadowsocks",
			Server:     "ss.example.com",
			ServerPort: 8388,
			Method:     "aes-128-gcm",
			Password:   "sspassword",
			Remark:     "SS-JP",
		},
	}
}

func TestSubconverter_NewFromRaw(t *testing.T) {
	rawInput := `trojan://trojanpass@tr.example.com:443?type=ws&security=tls&host=tr.example.com&path=%2Ftrojan-ws&sni=tr.example.com#Trojan-SG
vless://aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee@vl.example.com:443?type=grpc&security=tls&serviceName=vless-grpc#VLESS-ID
ss://YWVzLTEyOC1nY206c3NwYXNzd29yZA==@ss.example.com:8388#SS-JP`

	sub, err := subconverter.NewFromRaw(rawInput)
	require.NoError(t, err)
	assert.Equal(t, 3, sub.NodeCount())

	nodes := sub.Nodes()
	assert.Equal(t, "trojan", nodes[0].VPN)
	assert.Equal(t, "Trojan-SG", nodes[0].Remark)
	assert.Equal(t, "vless", nodes[1].VPN)
	assert.Equal(t, "VLESS-ID", nodes[1].Remark)
	assert.Equal(t, "shadowsocks", nodes[2].VPN)
	assert.Equal(t, "SS-JP", nodes[2].Remark)
}

func TestSubconverter_NewFromRaw_CommaSeparated(t *testing.T) {
	rawInput := `trojan://trojanpass@tr.example.com:443#Trojan-1,trojan://trojanpass@tr2.example.com:443#Trojan-2`
	sub, err := subconverter.NewFromRaw(rawInput)
	require.NoError(t, err)
	assert.Equal(t, 2, sub.NodeCount())
}

func TestSubconverter_NewFromRaw_Empty(t *testing.T) {
	sub, err := subconverter.NewFromRaw("   \n\n  ")
	assert.Error(t, err)
	assert.Nil(t, sub)
}

func TestSubconverter_ToRawAndBase64(t *testing.T) {
	sub := subconverter.New(sampleNodes())

	rawList := sub.ToRaw()
	assert.Len(t, rawList, 4)
	assert.Contains(t, rawList[0], "trojan://")
	assert.Contains(t, rawList[1], "vless://")
	assert.Contains(t, rawList[2], "vmess://")
	assert.Contains(t, rawList[3], "ss://")

	rawStr := sub.ToRawString()
	assert.Contains(t, rawStr, "\n")

	b64 := sub.ToBase64()
	decoded, err := base64.StdEncoding.DecodeString(b64)
	require.NoError(t, err)
	assert.Equal(t, rawStr, string(decoded))
}

func TestSubconverter_ToClash(t *testing.T) {
	sub := subconverter.New(sampleNodes())

	yamlStr, err := sub.ToClash("")
	require.NoError(t, err)
	assert.NotEmpty(t, yamlStr)

	var clashConfig map[string]any
	err = yaml.Unmarshal([]byte(yamlStr), &clashConfig)
	require.NoError(t, err)

	proxies, ok := clashConfig["proxies"].([]any)
	require.True(t, ok)
	assert.Len(t, proxies, 4)

	proxyGroups, ok := clashConfig["proxy-groups"].([]any)
	require.True(t, ok)
	assert.Len(t, proxyGroups, 3)

	rules, ok := clashConfig["rules"].([]any)
	require.True(t, ok)
	assert.Contains(t, rules, "MATCH,PROXIES")
}

func TestSubconverter_ToClash_TemplateCF(t *testing.T) {
	sub := subconverter.New(sampleNodes())

	yamlStr, err := sub.ToClash("cf")
	require.NoError(t, err)

	var clashConfig map[string]any
	err = yaml.Unmarshal([]byte(yamlStr), &clashConfig)
	require.NoError(t, err)

	proxies, ok := clashConfig["proxies"].([]any)
	require.True(t, ok)
	assert.Len(t, proxies, 5) // 4 original + 1 Trojan UDP

	lastProxy := proxies[4].(map[string]any)
	assert.Equal(t, "Trojan UDP", lastProxy["name"])
	assert.Equal(t, "trojan", lastProxy["type"])
	assert.Equal(t, "ss.example.com", lastProxy["server"]) // Last server used

	rules, ok := clashConfig["rules"].([]any)
	require.True(t, ok)
	assert.Equal(t, "NETWORK,UDP,Trojan UDP", rules[0])
}

func TestSubconverter_ToSingbox(t *testing.T) {
	sub := subconverter.New(sampleNodes())

	jsonStr, err := sub.ToSingbox("standard", "")
	require.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	var sbConfig map[string]any
	err = json.Unmarshal([]byte(jsonStr), &sbConfig)
	require.NoError(t, err)

	outbounds, ok := sbConfig["outbounds"].([]any)
	require.True(t, ok)
	// 2 groups (select, auto) + 4 nodes + 3 system (direct, block, dns-out) = 9
	assert.Len(t, outbounds, 9)

	firstOB := outbounds[0].(map[string]any)
	assert.Equal(t, "selector", firstOB["type"])
	assert.Equal(t, "select", firstOB["tag"])
}

func TestSubconverter_ToSFA_And_BFR(t *testing.T) {
	sub := subconverter.New(sampleNodes())

	sfaStr, err := sub.ToSFA("cf")
	require.NoError(t, err)

	var sfaConfig map[string]any
	err = json.Unmarshal([]byte(sfaStr), &sfaConfig)
	require.NoError(t, err)

	outbounds, ok := sfaConfig["outbounds"].([]any)
	require.True(t, ok)

	tags := []string{}
	for _, ob := range outbounds {
		m := ob.(map[string]any)
		tags = append(tags, m["tag"].(string))
	}

	assert.Contains(t, tags, "Internet")
	assert.Contains(t, tags, "Lock Region ID")
	assert.Contains(t, tags, "Best Latency")
	assert.Contains(t, tags, "Trojan UDP")

	// Verify BFR
	bfrStr, err := sub.ToBFR("")
	require.NoError(t, err)
	assert.Contains(t, bfrStr, "Internet")
}
