package middleware

import (
	"net/http"
	"time"

	"github.com/HAYASAKA7/HAYA_DISK/config"
	"github.com/HAYASAKA7/HAYA_DISK/models"
	"github.com/HAYASAKA7/HAYA_DISK/services"
)

// GetSessionCookie retrieves the session ID from cookies
func GetSessionCookie(r *http.Request) string {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// GetSessionUser retrieves the logged-in username from session
func GetSessionUser(r *http.Request) string {
	sessionID := GetSessionCookie(r)
	if sessionID == "" {
		return ""
	}

	session := services.GetSession(sessionID)
	if session == nil || time.Since(session.Timestamp) > time.Duration(config.SessionAge)*time.Second {
		return ""
	}
	return session.Username
}

// SetSessionCookie creates a new session cookie
func SetSessionCookie(w http.ResponseWriter, username string) {
	sessionID := services.GenerateSessionID()
	services.CreateSession(sessionID, &models.Session{
		Username:  username,
		Timestamp: time.Now(),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   config.SessionAge,
		HttpOnly: true,
	})
}

// ClearSessionCookie removes the session cookie
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
