package services

import (
	"crypto/rand"
	"fmt"
	"sync"

	"github.com/HAYASAKA7/HAYA_DISK/models"
)

var (
	sessionStore = make(map[string]*models.Session)
	sessionMu    sync.RWMutex
)

// CreateSession creates a new session
func CreateSession(sessionID string, session *models.Session) {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	sessionStore[sessionID] = session
}

// GetSession retrieves a session by ID
func GetSession(sessionID string) *models.Session {
	sessionMu.RLock()
	defer sessionMu.RUnlock()
	return sessionStore[sessionID]
}

// DeleteSession removes a session
func DeleteSession(sessionID string) {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	delete(sessionStore, sessionID)
}

// GenerateSessionID generates a random session ID
func GenerateSessionID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
