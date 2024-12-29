package model

import "time"

type UserData struct {
	UserID        int       `json:"user_id"`
	SoftAuthToken string    `json:"soft_auth_token"`
	Token         string    `json:"token"`
	Username      string    `json:"username"`
	Roles         []string  `json:"roles"`
	Accounts      []string  `json:"accounts"`
	ExpiresAt     time.Time `json:"expires_at"`
}

func (u *UserData) IsExpired(now time.Time) bool {
	return now.After(u.ExpiresAt)
}