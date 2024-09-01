package util

import (
	"sync"
	"time"
)

var UserCache *Cache

type Cache struct {
	userDataStore  sync.Map
	userTokenStore sync.Map
}

type UserData struct {
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	Username  string    `json:"username"`
	Roles     []string  `json:"roles"`
	Accounts  []string  `json:"accounts"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (u *UserData) IsExpired(now time.Time) bool {
	return now.After(u.ExpiresAt)
}

func NewCache() *Cache {
	cache := &Cache{}
	go cache.periodicCleanup()
	return cache
}

func (c *Cache) Set(token string, value UserData) {
	if oldToken, exists := c.userTokenStore.Load(value.UserID); exists {
		c.userDataStore.Delete(oldToken)
	}

	c.userTokenStore.Store(value.UserID, token)
	c.userDataStore.Store(token, value)
}

func (c *Cache) Get(token string) (UserData, bool) {
	if userDataRaw, exists := c.userDataStore.Load(token); exists {
		userData := userDataRaw.(UserData)
		if !userData.IsExpired(time.Now()) {
			return userData, true
		}
		c.Delete(token)
	}
	return UserData{}, false
}

func (c *Cache) Delete(token string) {
	if userDataRaw, exists := c.userDataStore.LoadAndDelete(token); exists {
		userData := userDataRaw.(UserData)
		c.userTokenStore.Delete(userData.UserID)
	}
}

func (c *Cache) DeleteTokensByUserID(userID string) {
	if token, exists := c.userTokenStore.LoadAndDelete(userID); exists {
		c.userDataStore.Delete(token)
	}
}

func (c *Cache) periodicCleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupExpiredEntries()
	}
}

func (c *Cache) cleanupExpiredEntries() {
	now := time.Now()

	c.userDataStore.Range(func(token, userDataRaw interface{}) bool {
		userData := userDataRaw.(UserData)
		if userData.IsExpired(now) {
			c.userDataStore.Delete(token)
			c.userTokenStore.Delete(userData.UserID)
		}
		return true
	})
}

func (c *Cache) GetAllEntries() []UserData {
	var entries []UserData
	c.userDataStore.Range(func(_, userDataRaw interface{}) bool {
		entries = append(entries, userDataRaw.(UserData))
		return true
	})
	return entries
}
