package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/HAYASAKA7/HAYA_DISK/models"
	_ "modernc.org/sqlite"
)

var db *sql.DB

// InitDatabase initializes the SQLite database and creates tables
func InitDatabase() error {
	var err error
	db, err = sql.Open("sqlite", "./haya-disk.db?_pragma=foreign_keys(1)")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT,
		phone TEXT,
		password TEXT NOT NULL,
		unique_code TEXT NOT NULL UNIQUE,
		created_at DATETIME NOT NULL,
		login_type TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_user_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_user_phone ON users(phone);
	CREATE INDEX IF NOT EXISTS idx_user_username ON users(username);

	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		filename TEXT NOT NULL,
		storage_path TEXT NOT NULL UNIQUE,
		parent_path TEXT NOT NULL DEFAULT '/',
		file_size INTEGER NOT NULL DEFAULT 0,
		mime_type TEXT,
		file_hash TEXT,
		is_directory BOOLEAN NOT NULL DEFAULT 0,
		uploaded_at DATETIME NOT NULL,
		modified_at DATETIME NOT NULL,
		FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_file_user ON files(username);
	CREATE INDEX IF NOT EXISTS idx_file_parent ON files(username, parent_path);
	CREATE INDEX IF NOT EXISTS idx_file_path ON files(storage_path);
	CREATE INDEX IF NOT EXISTS idx_file_hash ON files(file_hash);
	`

	_, err = db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return db
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// ==================== USER DATABASE OPERATIONS ====================

// CreateUserDB creates a new user in the database
func CreateUserDB(username, email, phone, password, uniqueCode string, createdAt time.Time, loginType string) error {
	query := `INSERT INTO users (username, email, phone, password, unique_code, created_at, login_type) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := db.Exec(query, username, email, phone, password, uniqueCode, createdAt, loginType)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUserByUsernameDB retrieves a user by username
func GetUserByUsernameDB(username string) (*models.User, error) {
	query := `SELECT username, email, phone, password, unique_code, created_at, login_type 
			  FROM users WHERE username = ?`

	var user models.User
	err := db.QueryRow(query, username).Scan(
		&user.Username, &user.Email, &user.Phone, &user.Password,
		&user.UniqueCode, &user.CreatedAt, &user.LoginType,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByEmailDB retrieves a user by email
func GetUserByEmailDB(email string) (*models.User, error) {
	if email == "" {
		return nil, nil
	}

	query := `SELECT username, email, phone, password, unique_code, created_at, login_type 
			  FROM users WHERE email = ?`

	var user models.User
	err := db.QueryRow(query, email).Scan(
		&user.Username, &user.Email, &user.Phone, &user.Password,
		&user.UniqueCode, &user.CreatedAt, &user.LoginType,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetUserByPhoneDB retrieves a user by phone
func GetUserByPhoneDB(phone string) (*models.User, error) {
	if phone == "" {
		return nil, nil
	}

	query := `SELECT username, email, phone, password, unique_code, created_at, login_type 
			  FROM users WHERE phone = ?`

	var user models.User
	err := db.QueryRow(query, phone).Scan(
		&user.Username, &user.Email, &user.Phone, &user.Password,
		&user.UniqueCode, &user.CreatedAt, &user.LoginType,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by phone: %w", err)
	}

	return &user, nil
}

// EmailExistsDB checks if an email already exists in the database
func EmailExistsDB(email string) (bool, error) {
	if email == "" {
		return false, nil
	}

	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = ?`
	err := db.QueryRow(query, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

// PhoneExistsDB checks if a phone number already exists in the database
func PhoneExistsDB(phone string) (bool, error) {
	if phone == "" {
		return false, nil
	}

	var count int
	query := `SELECT COUNT(*) FROM users WHERE phone = ?`
	err := db.QueryRow(query, phone).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check phone existence: %w", err)
	}

	return count > 0, nil
}

// GetAllUsersDB retrieves all users from the database
func GetAllUsersDB() ([]*models.User, error) {
	query := `SELECT username, email, phone, password, unique_code, created_at, login_type FROM users`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.Username, &user.Email, &user.Phone, &user.Password,
			&user.UniqueCode, &user.CreatedAt, &user.LoginType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// ==================== FILE DATABASE OPERATIONS ====================

// AddFileMetadata adds a file metadata record to the database
func AddFileMetadata(username, filename, storagePath, parentPath, mimeType, fileHash string, fileSize int64, isDir bool) error {
	query := `INSERT INTO files (username, filename, storage_path, parent_path, file_size, mime_type, file_hash, is_directory, uploaded_at, modified_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := db.Exec(query, username, filename, storagePath, parentPath, fileSize, mimeType, fileHash, isDir, now, now)
	if err != nil {
		return fmt.Errorf("failed to add file metadata: %w", err)
	}
	return nil
}

// GetUserFiles retrieves all files for a user in a specific parent folder
func GetUserFiles(username, parentPath string) ([]models.FileMetadata, error) {
	// Normalize parent path
	if parentPath == "" {
		parentPath = "/"
	}

	query := `SELECT id, username, filename, storage_path, parent_path, file_size, mime_type, file_hash, is_directory, uploaded_at, modified_at 
			  FROM files WHERE username = ? AND parent_path = ? ORDER BY is_directory DESC, filename ASC`

	rows, err := db.Query(query, username, parentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get user files: %w", err)
	}
	defer rows.Close()

	var files []models.FileMetadata
	for rows.Next() {
		var file models.FileMetadata
		var mimeType, fileHash sql.NullString

		err := rows.Scan(
			&file.ID, &file.Username, &file.Filename, &file.StoragePath,
			&file.ParentPath, &file.FileSize, &mimeType, &fileHash,
			&file.IsDirectory, &file.UploadedAt, &file.ModifiedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}

		file.MimeType = mimeType.String
		file.FileHash = fileHash.String
		files = append(files, file)
	}

	return files, nil
}

// DeleteFileMetadata deletes a file metadata record from the database
func DeleteFileMetadata(username, storagePath string) error {
	query := `DELETE FROM files WHERE username = ? AND storage_path = ?`
	_, err := db.Exec(query, username, storagePath)
	if err != nil {
		return fmt.Errorf("failed to delete file metadata: %w", err)
	}
	return nil
}

// DeleteFileMetadataRecursive deletes a file/folder and all its children
func DeleteFileMetadataRecursive(username, storagePath string) error {
	// Delete the file/folder itself
	query1 := `DELETE FROM files WHERE username = ? AND storage_path = ?`
	_, err := db.Exec(query1, username, storagePath)
	if err != nil {
		return fmt.Errorf("failed to delete file metadata: %w", err)
	}

	// Delete all children (for folders)
	query2 := `DELETE FROM files WHERE username = ? AND storage_path LIKE ?`
	_, err = db.Exec(query2, username, storagePath+"/%")
	if err != nil {
		return fmt.Errorf("failed to delete child files: %w", err)
	}

	return nil
}

// RenameFileMetadata updates the filename and storage path when a file is renamed
func RenameFileMetadata(username, oldPath, newPath, newFilename string) error {
	query := `UPDATE files SET filename = ?, storage_path = ?, modified_at = ? 
			  WHERE username = ? AND storage_path = ?`

	_, err := db.Exec(query, newFilename, newPath, time.Now(), username, oldPath)
	if err != nil {
		return fmt.Errorf("failed to rename file metadata: %w", err)
	}
	return nil
}

// MoveFileMetadata updates the parent path when a file is moved
func MoveFileMetadata(username, storagePath, newParentPath string) error {
	query := `UPDATE files SET parent_path = ?, modified_at = ? 
			  WHERE username = ? AND storage_path = ?`

	_, err := db.Exec(query, newParentPath, time.Now(), username, storagePath)
	if err != nil {
		return fmt.Errorf("failed to move file metadata: %w", err)
	}
	return nil
}

// FileExistsInDB checks if a file exists in the database
func FileExistsInDB(username, storagePath string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM files WHERE username = ? AND storage_path = ?`
	err := db.QueryRow(query, username, storagePath).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return count > 0, nil
}

// GetFileByPath retrieves a file metadata by username and storage path
func GetFileByPath(username, storagePath string) (*models.FileMetadata, error) {
	query := `SELECT id, username, filename, storage_path, parent_path, file_size, mime_type, file_hash, is_directory, uploaded_at, modified_at 
			  FROM files WHERE username = ? AND storage_path = ?`

	var file models.FileMetadata
	var mimeType, fileHash sql.NullString

	err := db.QueryRow(query, username, storagePath).Scan(
		&file.ID, &file.Username, &file.Filename, &file.StoragePath,
		&file.ParentPath, &file.FileSize, &mimeType, &fileHash,
		&file.IsDirectory, &file.UploadedAt, &file.ModifiedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file by path: %w", err)
	}

	file.MimeType = mimeType.String
	file.FileHash = fileHash.String

	return &file, nil
}

// GetUserStorageStats calculates total storage size and file count for a user
func GetUserStorageStats(username string) (totalSize int64, fileCount int, err error) {
	query := `SELECT COALESCE(SUM(file_size), 0), COUNT(*) FROM files WHERE username = ? AND is_directory = 0`

	err = db.QueryRow(query, username).Scan(&totalSize, &fileCount)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get storage stats: %w", err)
	}

	return totalSize, fileCount, nil
}

// SearchUserFiles searches for files by filename pattern
func SearchUserFiles(username, searchQuery string) ([]models.FileMetadata, error) {
	query := `SELECT id, username, filename, storage_path, parent_path, file_size, mime_type, file_hash, is_directory, uploaded_at, modified_at 
			  FROM files WHERE username = ? AND filename LIKE ? ORDER BY is_directory DESC, filename ASC`

	rows, err := db.Query(query, username, "%"+searchQuery+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}
	defer rows.Close()

	var files []models.FileMetadata
	for rows.Next() {
		var file models.FileMetadata
		var mimeType, fileHash sql.NullString

		err := rows.Scan(
			&file.ID, &file.Username, &file.Filename, &file.StoragePath,
			&file.ParentPath, &file.FileSize, &mimeType, &fileHash,
			&file.IsDirectory, &file.UploadedAt, &file.ModifiedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}

		file.MimeType = mimeType.String
		file.FileHash = fileHash.String
		files = append(files, file)
	}

	return files, nil
}

// GetAllFoldersDB retrieves all folders for a user (for move/copy operations)
func GetAllFoldersDB(username string) ([]models.FileMetadata, error) {
	query := `SELECT id, username, filename, storage_path, parent_path, file_size, mime_type, file_hash, is_directory, uploaded_at, modified_at 
			  FROM files WHERE username = ? AND is_directory = 1 ORDER BY parent_path, filename`

	rows, err := db.Query(query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get folders: %w", err)
	}
	defer rows.Close()

	var folders []models.FileMetadata
	for rows.Next() {
		var folder models.FileMetadata
		var mimeType, fileHash sql.NullString

		err := rows.Scan(
			&folder.ID, &folder.Username, &folder.Filename, &folder.StoragePath,
			&folder.ParentPath, &folder.FileSize, &mimeType, &fileHash,
			&folder.IsDirectory, &folder.UploadedAt, &folder.ModifiedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan folder: %w", err)
		}

		folder.MimeType = mimeType.String
		folder.FileHash = fileHash.String
		folders = append(folders, folder)
	}

	return folders, nil
}
