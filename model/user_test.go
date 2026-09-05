package model_test

import (
	"testing"
	"time"

	"github.com/LalatinaHub/common/model"
	"github.com/stretchr/testify/assert"
)

func TestUser_IsActive(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		user     model.User
		expected bool
	}{
		{
			name: "active user with valid fields and quota",
			user: model.User{
				Quota:      1000,
				Expired:    now.Add(24 * time.Hour),
				ServerCode: "SG1",
				VPN:        "trojan",
			},
			expected: true,
		},
		{
			name: "inactive user with zero quota",
			user: model.User{
				Quota:      0,
				Expired:    now.Add(24 * time.Hour),
				ServerCode: "SG1",
				VPN:        "trojan",
			},
			expected: false,
		},
		{
			name: "inactive user with negative quota",
			user: model.User{
				Quota:      -50,
				Expired:    now.Add(24 * time.Hour),
				ServerCode: "SG1",
				VPN:        "trojan",
			},
			expected: false,
		},
		{
			name: "inactive user expired in past",
			user: model.User{
				Quota:      1000,
				Expired:    now.Add(-1 * time.Hour),
				ServerCode: "SG1",
				VPN:        "trojan",
			},
			expected: false,
		},
		{
			name: "inactive user missing server code",
			user: model.User{
				Quota:      1000,
				Expired:    now.Add(24 * time.Hour),
				ServerCode: "",
				VPN:        "trojan",
			},
			expected: false,
		},
		{
			name: "inactive user missing VPN protocol",
			user: model.User{
				Quota:      1000,
				Expired:    now.Add(24 * time.Hour),
				ServerCode: "SG1",
				VPN:        "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.user.IsActive(now))
		})
	}
}
