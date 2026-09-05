package udprelay

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

const defaultBufferSize = 2048

var bufferPool = sync.Pool{
	New: func() any {
		b := make([]byte, defaultBufferSize)
		return &b
	},
}

// Relay sends raw UDP data to the target host and port, awaiting a response packet with timeout.
func Relay(ctx context.Context, host string, port int, data []byte, timeout time.Duration) ([]byte, error) {
	if host == "" {
		return nil, errors.New("target host cannot be empty")
	}
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("invalid target port: %d", port)
	}
	if len(data) == 0 {
		return nil, errors.New("data payload cannot be empty")
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))

	dialer := net.Dialer{
		Timeout: timeout,
	}

	conn, err := dialer.DialContext(ctx, "udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial UDP target %s: %w", addr, err)
	}
	defer conn.Close()

	deadline := time.Now().Add(timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}

	if err := conn.SetDeadline(deadline); err != nil {
		return nil, fmt.Errorf("failed to set UDP deadline: %w", err)
	}

	if _, err := conn.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write UDP packet: %w", err)
	}

	bufPtr := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(bufPtr)
	buffer := *bufPtr

	n, err := conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read UDP response: %w", err)
	}

	result := make([]byte, n)
	copy(result, buffer[:n])
	return result, nil
}

// RelayBase64 decodes a Base64-encoded UDP payload, relays it to target,
// and returns the Base64-encoded response payload.
func RelayBase64(ctx context.Context, host string, port int, b64Data string, timeout time.Duration) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		// Fallback to URL-safe encoding
		decoded, err = base64.URLEncoding.DecodeString(b64Data)
		if err != nil {
			return "", fmt.Errorf("invalid base64 data: %w", err)
		}
	}

	resp, err := Relay(ctx, host, port, decoded, timeout)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(resp), nil
}
