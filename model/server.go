package model

// Server represents a VPN edge server in the cluster.
type Server struct {
	ID         int64  `json:"id"`
	Code       string `json:"code"`
	Domain     string `json:"domain"`
	IP         string `json:"ip"`
	Country    string `json:"country"`
	UsersCount int64  `json:"users_count"`
	UsersMax   int64  `json:"users_max"`
}

// IsFull returns true if current users count reached maximum capacity.
func (s *Server) IsFull() bool {
	return s.UsersCount >= s.UsersMax
}
