package services

import (
	"strings"
	"sync"
	"time"
)

type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

var (
	fileListCache = make(map[string]*CacheEntry)
	cacheMutex    sync.RWMutex
	cacheTTL      = 5 * time.Second // Cache for 5 seconds
)

// GetCachedFileList retrieves cached file list
func GetCachedFileList(key string) (interface{}, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	entry, exists := fileListCache[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Data, true
}

// SetCachedFileList stores file list in cache
func SetCachedFileList(key string, data interface{}) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	fileListCache[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(cacheTTL),
	}
}

// InvalidateUserCache clears cache for a specific user
func InvalidateUserCache(username string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Remove all entries starting with username
	for key := range fileListCache {
		if strings.HasPrefix(key, username) {
			delete(fileListCache, key)
		}
	}
}
