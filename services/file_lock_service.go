package services

import (
	"sync"
)

var (
	// Per-user file operation locks
	userFileLocks = make(map[string]*sync.RWMutex)
	lockMapMutex  sync.RWMutex
)

// GetUserFileLock gets or creates a lock for a specific user
func GetUserFileLock(username string) *sync.RWMutex {
	lockMapMutex.RLock()
	lock, exists := userFileLocks[username]
	lockMapMutex.RUnlock()

	if exists {
		return lock
	}

	// Create new lock if doesn't exist
	lockMapMutex.Lock()
	defer lockMapMutex.Unlock()

	// Double-check after acquiring write lock
	if lock, exists := userFileLocks[username]; exists {
		return lock
	}

	lock = &sync.RWMutex{}
	userFileLocks[username] = lock
	return lock
}

// LockUserFileWrite locks for write operations (upload, delete, move)
func LockUserFileWrite(username string) {
	GetUserFileLock(username).Lock()
}

// UnlockUserFileWrite unlocks after write operations
func UnlockUserFileWrite(username string) {
	GetUserFileLock(username).Unlock()
}

// LockUserFileRead locks for read operations (download, list)
func LockUserFileRead(username string) {
	GetUserFileLock(username).RLock()
}

// UnlockUserFileRead unlocks after read operations
func UnlockUserFileRead(username string) {
	GetUserFileLock(username).RUnlock()
}
