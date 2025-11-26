package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HAYASAKA7/HAYA-DISK/config"
	"github.com/HAYASAKA7/HAYA-DISK/models"
	"github.com/HAYASAKA7/HAYA-DISK/services"
)

// MigrateFromJSON migrates users from users.json and files from storage to database
func MigrateFromJSON() error {
	log.Println("Starting migration from JSON to SQLite database...")

	// Step 1: Migrate users
	if err := migrateUsers(); err != nil {
		return fmt.Errorf("failed to migrate users: %w", err)
	}

	// Step 2: Migrate files
	if err := migrateFiles(); err != nil {
		return fmt.Errorf("failed to migrate files: %w", err)
	}

	log.Println("Migration completed successfully!")
	return nil
}

// migrateUsers migrates users from users.json to database
func migrateUsers() error {
	log.Println("Migrating users from users.json...")

	// Check if users.json exists
	if _, err := os.Stat(config.UsersFile); os.IsNotExist(err) {
		log.Println("users.json not found, skipping user migration")
		return nil
	}

	// Read users.json
	data, err := os.ReadFile(config.UsersFile)
	if err != nil {
		return fmt.Errorf("failed to read users.json: %w", err)
	}

	var users []models.User
	if err := json.Unmarshal(data, &users); err != nil {
		return fmt.Errorf("failed to parse users.json: %w", err)
	}

	// Insert each user into database
	for _, user := range users {
		// Check if user already exists
		existingUser, _ := services.GetUserByUsernameDB(user.Username)
		if existingUser != nil {
			log.Printf("User %s already exists in database, skipping", user.Username)
			continue
		}

		// Parse created_at time
		createdAt, err := time.Parse("2006-01-02 15:04:05", user.CreatedAt)
		if err != nil {
			// If parsing fails, use current time
			createdAt = time.Now()
		}

		// Insert user
		err = services.CreateUserDB(
			user.Username,
			user.Email,
			user.Phone,
			user.Password,
			user.UniqueCode,
			createdAt,
			user.LoginType,
		)
		if err != nil {
			log.Printf("Failed to insert user %s: %v", user.Username, err)
			continue
		}

		log.Printf("Migrated user: %s", user.Username)
	}

	log.Printf("Migrated %d users", len(users))
	return nil
}

// migrateFiles scans storage directory and adds all files to database
func migrateFiles() error {
	log.Println("Migrating files from storage directory...")

	// Check if storage directory exists
	if _, err := os.Stat(config.StorageDir); os.IsNotExist(err) {
		log.Println("storage directory not found, skipping file migration")
		return nil
	}

	// Get all users from database
	users, err := services.GetAllUsersDB()
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}

	totalFiles := 0
	for _, user := range users {
		userStoragePath := services.GetUserStoragePath(user.Username, user.UniqueCode)

		// Check if user storage exists
		if _, err := os.Stat(userStoragePath); os.IsNotExist(err) {
			log.Printf("Storage path not found for user %s: %s", user.Username, userStoragePath)
			continue
		}

		// Walk through user's storage directory
		fileCount := 0
		err := filepath.Walk(userStoragePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			// Skip the root directory
			if path == userStoragePath {
				return nil
			}

			// Get relative path from user's storage root
			relativePath, err := filepath.Rel(userStoragePath, path)
			if err != nil {
				return nil
			}

			// Normalize path separators for database storage
			relativePath = filepath.ToSlash(relativePath)

			// Check if already exists in database
			exists, _ := services.FileExistsInDB(user.Username, relativePath)
			if exists {
				return nil
			}

			// Determine parent path
			parentPath := filepath.Dir(relativePath)
			if parentPath == "." {
				parentPath = "/"
			} else {
				parentPath = filepath.ToSlash(parentPath)
			}

			// Get filename
			filename := filepath.Base(relativePath)

			// Calculate file hash (only for files, not directories)
			var fileHash string
			var mimeType string
			if !info.IsDir() {
				fileHash = CalculateFileSHA256(path)
				ext := strings.ToLower(filepath.Ext(filename))
				mimeType = getMimeType(ext)
			}

			// Add to database
			err = services.AddFileMetadata(
				user.Username,
				filename,
				relativePath,
				parentPath,
				mimeType,
				fileHash,
				info.Size(),
				info.IsDir(),
			)
			if err != nil {
				log.Printf("Failed to add file %s for user %s: %v", relativePath, user.Username, err)
				return nil
			}

			fileCount++
			if fileCount%100 == 0 {
				log.Printf("Migrated %d files for user %s...", fileCount, user.Username)
			}

			return nil
		})

		if err != nil {
			log.Printf("Error walking storage for user %s: %v", user.Username, err)
			continue
		}

		log.Printf("Migrated %d files for user: %s", fileCount, user.Username)
		totalFiles += fileCount
	}

	log.Printf("Total files migrated: %d", totalFiles)
	return nil
}

// getMimeType returns a basic MIME type based on file extension
func getMimeType(ext string) string {
	mimeTypes := map[string]string{
		// Images
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".bmp":  "image/bmp",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
		// Videos
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mkv":  "video/x-matroska",
		".mov":  "video/quicktime",
		".wmv":  "video/x-ms-wmv",
		".flv":  "video/x-flv",
		".webm": "video/webm",
		".m4v":  "video/x-m4v",
		// Audio
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".flac": "audio/flac",
		".aac":  "audio/aac",
		".ogg":  "audio/ogg",
		".wma":  "audio/x-ms-wma",
		".m4a":  "audio/mp4",
		".opus": "audio/opus",
		// Documents
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".txt":  "text/plain",
		".rtf":  "application/rtf",
		".odt":  "application/vnd.oasis.opendocument.text",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".md":   "text/markdown",
		// Archives
		".zip": "application/zip",
		".rar": "application/x-rar-compressed",
		".7z":  "application/x-7z-compressed",
		".tar": "application/x-tar",
		".gz":  "application/gzip",
		// Code
		".html": "text/html",
		".css":  "text/css",
		".js":   "text/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".py":   "text/x-python",
		".go":   "text/x-go",
		".java": "text/x-java",
		".c":    "text/x-c",
		".cpp":  "text/x-c++",
		".php":  "text/x-php",
	}

	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}
	return "application/octet-stream"
}
