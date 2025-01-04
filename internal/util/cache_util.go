package util

import (
	"sync"
	"time"

	"github.com/wise-hub/gateway/internal/model"
)

var UserCache *Cache

type Cache struct {
	userDataStore  sync.Map
	userTokenStore sync.Map
}

func NewCache() *Cache {
	cache := &Cache{}
	go cache.periodicCleanup()
	return cache
}

func (c *Cache) Set(token string, value model.UserData) {
	if oldToken, exists := c.userTokenStore.Load(value.UserID); exists {
		c.userDataStore.Delete(oldToken)
	}

	c.userTokenStore.Store(value.UserID, token)
	c.userDataStore.Store(token, value)
}

func (c *Cache) Get(token string) (model.UserData, bool) {
	if userDataRaw, exists := c.userDataStore.Load(token); exists {
		userData := userDataRaw.(model.UserData)
		now := time.Now()

		if userData.IsExpired(now) {
			c.Delete(token)
			return model.UserData{}, false
		}

		if userData.ExpiresAt.Sub(now) < 10 * time.Minute {
			userData.ExpiresAt = now.Add(1 * time.Hour)
			c.userDataStore.Store(token, userData)
		}

		return userData, true
	}
	return model.UserData{}, false
}


func (c *Cache) Delete(token string) {
	if userDataRaw, exists := c.userDataStore.LoadAndDelete(token); exists {
		userData := userDataRaw.(model.UserData)
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
		userData := userDataRaw.(model.UserData)
		if userData.IsExpired(now) {
			c.userDataStore.Delete(token)
			c.userTokenStore.Delete(userData.UserID)
		}
		return true
	})
}

func (c *Cache) GetAllEntries() []model.UserData {
	var entries []model.UserData
	c.userDataStore.Range(func(_, userDataRaw interface{}) bool {
		entries = append(entries, userDataRaw.(model.UserData))
		return true
	})
	return entries
}
