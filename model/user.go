// Package model defines the core domain entities used across LalatinaHub services.
// It has zero external dependencies and relies solely on the Go standard library.
package model

import "time"

// User represents a VPN user entity.
type User struct {
	ID         int64     `json:"id"`
	Token      string    `json:"token"`
	Password   string    `json:"password"`
	Expired    time.Time `json:"expired"`
	ServerCode string    `json:"server_code"`
	Quota      int64     `json:"quota"` // Quota in bytes
	Relay      string    `json:"relay"`
	Adblock    bool      `json:"adblock"`
	VPN        string    `json:"vpn"`
}

// IsActive returns true if the user has remaining quota, has not expired,
// and has assigned ServerCode and VPN protocol.
func (u *User) IsActive(now time.Time) bool {
	return u.Quota > 0 && !now.After(u.Expired) && u.ServerCode != "" && u.VPN != ""
}
