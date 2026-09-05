package udprelay_test

import (
	"context"
	"encoding/base64"
	"net"
	"testing"
	"time"

	"github.com/LalatinaHub/common/udprelay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelay_Success(t *testing.T) {
	// Start local mock UDP echo server
	serverConn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	})
	require.NoError(t, err)
	defer serverConn.Close()

	serverPort := serverConn.LocalAddr().(*net.UDPAddr).Port

	// Echo goroutine
	go func() {
		buf := make([]byte, 1024)
		for {
			n, clientAddr, err := serverConn.ReadFrom(buf)
			if err != nil {
				return
			}
			// Echo back payload prefixed with "echo:"
			resp := append([]byte("echo:"), buf[:n]...)
			serverConn.WriteTo(resp, clientAddr)
		}
	}()

	ctx := context.Background()
	payload := []byte("hello-udp")

	resp, err := udprelay.Relay(ctx, "127.0.0.1", serverPort, payload, 2*time.Second)
	require.NoError(t, err)
	assert.Equal(t, "echo:hello-udp", string(resp))

	// Test RelayBase64
	b64Payload := base64.StdEncoding.EncodeToString(payload)
	b64Resp, err := udprelay.RelayBase64(ctx, "127.0.0.1", serverPort, b64Payload, 2*time.Second)
	require.NoError(t, err)

	decodedResp, err := base64.StdEncoding.DecodeString(b64Resp)
	require.NoError(t, err)
	assert.Equal(t, "echo:hello-udp", string(decodedResp))
}

func TestRelay_ValidationErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("empty host", func(t *testing.T) {
		_, err := udprelay.Relay(ctx, "", 8080, []byte("data"), time.Second)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "host cannot be empty")
	})

	t.Run("invalid port", func(t *testing.T) {
		_, err := udprelay.Relay(ctx, "127.0.0.1", 0, []byte("data"), time.Second)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid target port")
	})

	t.Run("empty payload", func(t *testing.T) {
		_, err := udprelay.Relay(ctx, "127.0.0.1", 8080, nil, time.Second)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "payload cannot be empty")
	})

	t.Run("invalid base64", func(t *testing.T) {
		_, err := udprelay.RelayBase64(ctx, "127.0.0.1", 8080, "!!not-base64", time.Second)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid base64")
	})
}

func TestRelay_Timeout(t *testing.T) {
	// UDP server that never responds
	serverConn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	})
	require.NoError(t, err)
	defer serverConn.Close()

	serverPort := serverConn.LocalAddr().(*net.UDPAddr).Port

	ctx := context.Background()
	_, err = udprelay.Relay(ctx, "127.0.0.1", serverPort, []byte("ping"), 100*time.Millisecond)
	assert.Error(t, err)
}
