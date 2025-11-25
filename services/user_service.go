package services

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/HAYASAKA7/HAYA_DISK/config"
	"github.com/HAYASAKA7/HAYA_DISK/models"
)

var (
	users    = make(map[string]*models.User)
	sessions = make(map[string]*models.Session)
	mu       sync.RWMutex
)

// LoadUsers loads all users from the users.json file
func LoadUsers() {
	data, err := os.ReadFile(config.UsersFile)
	if err != nil {
		return // File doesn't exist yet
	}
	var userList []models.User
	if err := json.Unmarshal(data, &userList); err != nil {
		return
	}
	mu.Lock()
	for i := range userList {
		users[userList[i].Username] = &userList[i]
	}
	mu.Unlock()
}

// SaveUsers saves all users to the users.json file
func SaveUsers() {
	mu.RLock()
	var userList []*models.User
	for _, user := range users {
		userCopy := *user // Create a copy to avoid race conditions
		userList = append(userList, &userCopy)
	}
	mu.RUnlock()

	data, err := json.MarshalIndent(userList, "", "  ")
	if err != nil {
		return
	}

	// Write to temp file first, then atomic rename
	tempFile := config.UsersFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return
	}
	os.Rename(tempFile, config.UsersFile)
}

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

// FindUserByCredential finds a user by username, email, or phone
func FindUserByCredential(credential string) *models.User {
	mu.RLock()
	defer mu.RUnlock()

	// Try username first
	if user, exists := users[credential]; exists {
		return user
	}

	// Try email or phone
	for _, user := range users {
		if user.Email == credential || user.Phone == credential {
			return user
		}
	}
	return nil
}

// EmailExists checks if email is already registered
func EmailExists(email string) bool {
	if email == "" {
		return false
	}
	mu.RLock()
	defer mu.RUnlock()
	for _, user := range users {
		if user.Email == email {
			return true
		}
	}
	return false
}

// PhoneExists checks if phone is already registered
func PhoneExists(phone string) bool {
	if phone == "" {
		return false
	}
	mu.RLock()
	defer mu.RUnlock()
	for _, user := range users {
		if user.Phone == phone {
			return true
		}
	}
	return false
}

// CreateUser creates and saves a new user
func CreateUser(username, email, phone, password string) (*models.User, error) {
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

	newUser := &models.User{
		Username:   username,
		Email:      email,
		Phone:      phone,
		Password:   password,
		UniqueCode: uniqueCode,
		CreatedAt:  time.Now().Format("2006-01-02 15:04:05"),
		LoginType:  loginType,
	}

	userStoragePath := GetUserStoragePath(username, uniqueCode)
	os.MkdirAll(userStoragePath, os.ModePerm)

	mu.Lock()
	users[username] = newUser
	mu.Unlock()
	SaveUsers()

	return newUser, nil
}

// UpdateUserProfile updates user's email and phone
func UpdateUserProfile(username, email, phone string) error {
	mu.Lock()
	user, exists := users[username]
	if !exists {
		mu.Unlock()
		return fmt.Errorf("user not found")
	}

	// Update user
	if email != "" {
		user.Email = email
	}
	if phone != "" {
		user.Phone = phone
	}

	// Update login type
	if user.Email != "" && user.Phone != "" {
		user.LoginType = "both"
	} else if user.Email != "" {
		user.LoginType = "email"
	} else {
		user.LoginType = "phone"
	}

	mu.Unlock()
	SaveUsers()

	return nil
}

// UsernameExists checks if username is already taken
func UsernameExists(username string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, exists := users[username]
	return exists
}

// UpdateUsername changes a user's username and renames their storage folder
func UpdateUsername(oldUsername, newUsername string) error {
	if oldUsername == newUsername {
		return nil // No change needed
	}

	mu.Lock()
	defer mu.Unlock()

	// Check if old user exists
	user, exists := users[oldUsername]
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Check if new username is already taken
	if _, taken := users[newUsername]; taken {
		return fmt.Errorf("username already exists")
	}

	// Rename storage folder
	oldPath := GetUserStoragePath(oldUsername, user.UniqueCode)
	newPath := GetUserStoragePath(newUsername, user.UniqueCode)

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename storage folder: %v", err)
	}

	// Update username in user object
	user.Username = newUsername

	// Update in map
	delete(users, oldUsername)
	users[newUsername] = user

	// Save to file
	mu.Unlock()
	SaveUsers()
	mu.Lock()

	return nil
}

// GetUser retrieves a user by username
func GetUser(username string) *models.User {
	mu.RLock()
	defer mu.RUnlock()
	return users[username]
}
