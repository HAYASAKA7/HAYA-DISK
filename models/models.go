package models

import "time"

// User represents a user account
type User struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Password   string `json:"password"`
	UniqueCode string `json:"unique_code"`
	CreatedAt  string `json:"created_at"`
	LoginType  string `json:"login_type"` // "email", "phone", or "both"
}

// Session represents an active user session
type Session struct {
	Username  string
	Timestamp time.Time
}

// FileInfo contains metadata about a file
type FileInfo struct {
	Name     string
	Size     string
	Modified string
	Icon     string
	IsImage  bool
	Ext      string
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// UpdateProfileResponse represents a profile update response
type UpdateProfileResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
