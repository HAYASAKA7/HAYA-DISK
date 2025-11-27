package services

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/HAYASAKA7/HAYA-DISK/config"
	"github.com/HAYASAKA7/HAYA-DISK/models"
)

var (
	sessions = make(map[string]*models.Session)
	mu       sync.RWMutex
)

// GenerateUniqueCode generates a random unique code
func GenerateUniqueCode() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// GetUserStoragePath returns the storage path for a user
func GetUserStoragePath(username, uniqueCode string) string {
	return filepath.Join(config.StorageDir, fmt.Sprintf("%s_%s", username, uniqueCode))
}

// FindUserByCredential finds a user by email or phone (username login is disabled)
func FindUserByCredential(credential string) *models.User {
	// Try email first
	if user, _ := GetUserByEmailDB(credential); user != nil {
		return user
	}

	// Try phone
	if user, _ := GetUserByPhoneDB(credential); user != nil {
		return user
	}

	return nil
}

// FindUserByEmail finds a user by email
func FindUserByEmail(email string) *models.User {
	if email == "" {
		return nil
	}
	user, _ := GetUserByEmailDB(email)
	return user
}

// FindUserByPhoneAndRegion finds a user by phone number and region
func FindUserByPhoneAndRegion(phone, phoneRegion string) *models.User {
	if phone == "" {
		return nil
	}
	user, _ := GetUserByPhoneAndRegionDB(phone, phoneRegion)
	return user
}

// EmailExists checks if email is already registered
func EmailExists(email string) bool {
	exists, _ := EmailExistsDB(email)
	return exists
}

// PhoneExists checks if phone is already registered
func PhoneExists(phone string) bool {
	exists, _ := PhoneExistsDB(phone)
	return exists
}

// PhoneAndRegionExists checks if phone+region combination is already registered
func PhoneAndRegionExists(phone, phoneRegion string) bool {
	exists, _ := PhoneAndRegionExistsDB(phone, phoneRegion)
	return exists
}

// CreateUser creates and saves a new user
func CreateUser(username, email, phone, phoneRegion, password string) (*models.User, error) {
	uniqueCode := GenerateUniqueCode()

	// Determine login type
	loginType := ""
	if email != "" && phone != "" {
		loginType = "both"
	} else if email != "" {
		loginType = "email"
	} else {
		loginType = "phone"
	}

	createdAt := time.Now()

	// Create user in database
	err := CreateUserDB(username, email, phone, phoneRegion, password, uniqueCode, createdAt, loginType)
	if err != nil {
		return nil, err
	}

	// Create storage folder
	userStoragePath := GetUserStoragePath(username, uniqueCode)
	os.MkdirAll(userStoragePath, os.ModePerm)

	// Return the created user
	return GetUserByUsernameDB(username)
}

// UpdateUserProfile updates user's email and phone
func UpdateUserProfile(username, email, phone string) error {
	return UpdateUserProfileWithRegion(username, email, phone, "")
}

// UpdateUserProfileWithRegion updates user's email, phone, and phone region
func UpdateUserProfileWithRegion(username, email, phone, phoneRegion string) error {
	// Get user from database
	user, err := GetUserByUsernameDB(username)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Update fields
	if email != "" {
		user.Email = email
	}
	if phone != "" {
		user.Phone = phone
		user.PhoneRegion = phoneRegion
	}

	// Update login type
	if user.Email != "" && user.Phone != "" {
		user.LoginType = "both"
	} else if user.Email != "" {
		user.LoginType = "email"
	} else {
		user.LoginType = "phone"
	}

	// Update in database
	query := `UPDATE users SET email = ?, phone = ?, phone_region = ?, login_type = ? WHERE username = ?`
	_, err = GetDB().Exec(query, user.Email, user.Phone, user.PhoneRegion, user.LoginType, username)
	return err
}

// UsernameExists checks if username is already taken
func UsernameExists(username string) bool {
	user, _ := GetUserByUsernameDB(username)
	return user != nil
}

// UpdateUsername changes a user's username and renames their storage folder
func UpdateUsername(oldUsername, newUsername string) error {
	if oldUsername == newUsername {
		return nil // No change needed
	}

	// Check if old user exists
	user, err := GetUserByUsernameDB(oldUsername)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Check if new username is already taken
	if UsernameExists(newUsername) {
		return fmt.Errorf("username already exists")
	}

	// Rename storage folder
	oldPath := GetUserStoragePath(oldUsername, user.UniqueCode)
	newPath := GetUserStoragePath(newUsername, user.UniqueCode)

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename storage folder: %v", err)
	}

	// Update username in database
	query := `UPDATE users SET username = ? WHERE username = ?`
	_, err = GetDB().Exec(query, newUsername, oldUsername)
	if err != nil {
		// Rollback folder rename
		os.Rename(newPath, oldPath)
		return fmt.Errorf("failed to update username in database: %v", err)
	}

	// Also update files table
	query = `UPDATE files SET username = ? WHERE username = ?`
	GetDB().Exec(query, newUsername, oldUsername)

	return nil
}

// GetUser retrieves a user by username
func GetUser(username string) *models.User {
	user, _ := GetUserByUsernameDB(username)
	return user
}
