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
	fileListCache   = make(map[string]*CacheEntry)
	folderSizeCache = make(map[string]int64) // Cache folder sizes permanently
	cacheMutex      sync.RWMutex
	cacheTTL        = 5 * time.Second // Cache for 5 seconds
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

	// Remove folder size cache for this user
	for key := range folderSizeCache {
		if strings.HasPrefix(key, username) {
			delete(folderSizeCache, key)
		}
	}
}

// GetFolderSize retrieves cached folder size
func GetFolderSize(folderPath string) (int64, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	size, exists := folderSizeCache[folderPath]
	return size, exists
}

// SetFolderSize caches folder size
func SetFolderSize(folderPath string, size int64) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	folderSizeCache[folderPath] = size
}

// UpdateFolderSize updates folder size by adding/subtracting file size
func UpdateFolderSize(folderPath string, sizeDelta int64) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	folderSizeCache[folderPath] += sizeDelta
}

// RecalculateFolderSize forces recalculation and caching of folder size
func RecalculateFolderSize(folderPath string, calculateFunc func(string) int64) int64 {
	size := calculateFunc(folderPath)
	SetFolderSize(folderPath, size)
	return size
}
