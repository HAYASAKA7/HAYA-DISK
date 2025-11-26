package handlers

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/HAYASAKA7/HAYA-DISK/config"
	"github.com/HAYASAKA7/HAYA-DISK/middleware"
	"github.com/HAYASAKA7/HAYA-DISK/services"
	"github.com/HAYASAKA7/HAYA-DISK/utils"
)

// UploadHandler handles file uploads
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetSessionUser(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		// Show upload page
		user := services.GetUser(username)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		userStoragePath := services.GetUserStoragePath(username, user.UniqueCode)
		folders, _ := getFolderList(userStoragePath)

		data := map[string]interface{}{
			"username": username,
			"folders":  folders,
		}

		tmpl, _ := template.ParseFiles(filepath.Join(config.TemplatesDir, "upload.html"))
		tmpl.Execute(w, data)
		return
	}

	if r.Method == http.MethodPost {
		user := services.GetUser(username)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "File error", 400)
			return
		}
		defer file.Close()

		// Get target folder from form
		folder := r.FormValue("folder")
		userStoragePath := services.GetUserStoragePath(username, user.UniqueCode)

		// Normalize folder path
		if folder == "/" {
			folder = ""
		}

		// Build target path
		var targetPath string
		if folder != "" && folder != "/" {
			targetPath = filepath.Join(userStoragePath, folder)
			// Security check
			if !isPathSafe(targetPath, userStoragePath) {
				http.Error(w, "Invalid folder", http.StatusForbidden)
				return
			}
		} else {
			targetPath = userStoragePath
			folder = "/"
		}

		filePath := filepath.Join(targetPath, header.Filename)

		// LOCK before file operations
		services.LockUserFileWrite(username)
		defer services.UnlockUserFileWrite(username)

		// Check if file already exists in database
		relativePath, _ := filepath.Rel(userStoragePath, filePath)
		exists, _ := services.FileExistsInDB(username, relativePath)
		if exists {
			http.Error(w, "File already exists", http.StatusConflict)
			return
		}

		// Check if file already exists on disk (safety check)
		if _, err := os.Stat(filePath); err == nil {
			http.Error(w, "File already exists", http.StatusConflict)
			return
		}

		f, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Save error", 500)
			return
		}
		defer f.Close()

		// Copy file content
		written, err := io.Copy(f, file)
		if err != nil {
			os.Remove(filePath) // Cleanup on error
			http.Error(w, "Save error", 500)
			return
		}

		// Calculate file hash
		fileHash := utils.CalculateFileSHA256(filePath)

		// Get MIME type
		mimeType := header.Header.Get("Content-Type")
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		// Save file metadata to database
		err = services.AddFileMetadata(
			username,
			header.Filename,
			relativePath,
			folder,
			mimeType,
			fileHash,
			written,
			false, // not a directory
		)
		if err != nil {
			// If database insert fails, remove the file
			os.Remove(filePath)
			http.Error(w, "Failed to save file metadata", 500)
			return
		}

		// Update folder size cache
		fileInfo, _ := f.Stat()
		services.UpdateFolderSize(targetPath, fileInfo.Size())

		// Invalidate cache after upload
		services.InvalidateUserCache(username)

		// Redirect back to the folder where the file was uploaded
		if folder != "" && folder != "/" {
			http.Redirect(w, r, "/list?folder="+folder, http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/list", http.StatusSeeOther)
		}
	}
}

// DownloadHandler handles file downloads
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetSessionUser(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Lock for read operation
	services.LockUserFileRead(username)
	defer services.UnlockUserFileRead(username)

	name := r.URL.Query().Get("name")
	folder := r.URL.Query().Get("folder")
	if name == "" {
		http.Error(w, "Missing file name", 400)
		return
	}

	user := services.GetUser(username)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userStoragePath := services.GetUserStoragePath(username, user.UniqueCode)

	// Build file path
	var filePath string
	if folder != "" && folder != "/" {
		filePath = filepath.Join(userStoragePath, folder, name)
	} else {
		filePath = filepath.Join(userStoragePath, name)
	}

	// Security check: ensure path is within user's storage
	if !isPathSafe(filePath, userStoragePath) {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", 404)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Disposition", "attachment; filename="+name)
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, f)
}

// ThumbnailHandler handles image thumbnail display
func ThumbnailHandler(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetSessionUser(r)
	if username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Lock for read operation
	services.LockUserFileRead(username)
	defer services.UnlockUserFileRead(username)

	name := r.URL.Query().Get("name")
	folder := r.URL.Query().Get("folder")
	if name == "" {
		http.Error(w, "Missing file name", 400)
		return
	}

	user := services.GetUser(username)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ext := strings.ToLower(filepath.Ext(name))
	if !utils.IsImageFile(ext) {
		http.Error(w, "Not an image file", 400)
		return
	}

	userStoragePath := services.GetUserStoragePath(username, user.UniqueCode)

	// Build file path
	var filePath string
	if folder != "" && folder != "/" {
		filePath = filepath.Join(userStoragePath, folder, name)
	} else {
		filePath = filepath.Join(userStoragePath, name)
	}

	// Security check
	if !isPathSafe(filePath, userStoragePath) {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", 404)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", utils.GetImageContentType(ext))
	io.Copy(w, f)
}

// isPathSafe checks if a file path is within allowed directory
func isPathSafe(filePath, allowedDir string) bool {
	absPath, _ := filepath.Abs(filePath)
	absAllowed, _ := filepath.Abs(allowedDir)
	return strings.HasPrefix(absPath, absAllowed)
}

// DeleteHandler handles file and folder deletion
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetSessionUser(r)
	if username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	name := r.URL.Query().Get("name")
	folder := r.URL.Query().Get("folder")
	if name == "" {
		http.Error(w, "Missing file/folder name", 400)
		return
	}

	user := services.GetUser(username)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userStoragePath := services.GetUserStoragePath(username, user.UniqueCode)

	// Build full path
	var targetPath string
	if folder != "" && folder != "/" {
		targetPath = filepath.Join(userStoragePath, folder, name)
	} else {
		targetPath = filepath.Join(userStoragePath, name)
	}

	// Security check
	if !isPathSafe(targetPath, userStoragePath) {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Lock for write operation
	services.LockUserFileWrite(username)
	defer services.UnlockUserFileWrite(username)

	// Get relative path for database operations
	relativePath, _ := filepath.Rel(userStoragePath, targetPath)

	// Get file/folder size before deletion from database
	var deletedSize int64
	// Check if it's a directory from database
	fileMetadata, err := services.GetFileByPath(username, relativePath)
	if err == nil && fileMetadata != nil {
		if fileMetadata.IsDirectory {
			// Calculate folder size from database
			deletedSize, _ = services.CalculateFolderSizeDB(username, relativePath)
		} else {
			deletedSize = fileMetadata.FileSize
		}
	}

	// Delete from database first (including children if folder)
	err = services.DeleteFileMetadataRecursive(username, relativePath)
	if err != nil {
		http.Error(w, "Database delete error", 500)
		return
	}

	// Delete file or folder from disk
	err = os.RemoveAll(targetPath)
	if err != nil {
		http.Error(w, "Delete error", 500)
		return
	}

	// Update parent folder size cache
	var parentFolder string
	if folder != "" && folder != "/" {
		parentFolder = filepath.Join(userStoragePath, folder)
	} else {
		parentFolder = userStoragePath
	}
	services.UpdateFolderSize(parentFolder, -deletedSize)

	// Invalidate cache after deletion
	services.InvalidateUserCache(username)

	// Redirect back to folder or root
	if folder != "" && folder != "/" {
		http.Redirect(w, r, "/list?folder="+folder, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/list", http.StatusSeeOther)
	}
}

// CreateFolderHandler handles folder creation
func CreateFolderHandler(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetSessionUser(r)
	if username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	folderName := r.FormValue("folder_name")
	currentFolder := r.FormValue("current_folder")

	if folderName == "" {
		http.Error(w, "Missing folder name", 400)
		return
	}

	// Sanitize folder name
	folderName = strings.TrimSpace(folderName)
	if strings.Contains(folderName, "..") || strings.ContainsAny(folderName, "/\\") {
		http.Error(w, "Invalid folder name", 400)
		return
	}

	user := services.GetUser(username)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userStoragePath := services.GetUserStoragePath(username, user.UniqueCode)

	// Normalize current folder
	if currentFolder == "" || currentFolder == "/" {
		currentFolder = "/"
	}

	// Build target path
	var targetPath string
	if currentFolder != "" && currentFolder != "/" {
		targetPath = filepath.Join(userStoragePath, currentFolder, folderName)
	} else {
		targetPath = filepath.Join(userStoragePath, folderName)
	}

	// Security check
	if !isPathSafe(targetPath, userStoragePath) {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Lock for write operation
	services.LockUserFileWrite(username)
	defer services.UnlockUserFileWrite(username)

	// Get relative path for database
	relativePath, _ := filepath.Rel(userStoragePath, targetPath)

	// Check if folder already exists in database
	exists, _ := services.FileExistsInDB(username, relativePath)
	if exists {
		http.Error(w, "Folder already exists", http.StatusConflict)
		return
	}

	// Create folder on disk
	err := os.MkdirAll(targetPath, os.ModePerm)
	if err != nil {
		http.Error(w, "Failed to create folder", 500)
		return
	}

	// Add folder metadata to database
	err = services.AddFileMetadata(
		username,
		folderName,
		relativePath,
		currentFolder,
		"",   // no mime type for folders
		"",   // no hash for folders
		0,    // folders have 0 size
		true, // is directory
	)
	if err != nil {
		// Cleanup on error
		os.Remove(targetPath)
		http.Error(w, "Failed to save folder metadata", 500)
		return
	}

	// Invalidate cache after folder creation
	services.InvalidateUserCache(username)

	// Redirect back
	if currentFolder != "" && currentFolder != "/" {
		http.Redirect(w, r, "/list?folder="+currentFolder, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/list", http.StatusSeeOther)
	}
}

// MoveFileHandler handles moving files to folders
func MoveFileHandler(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetSessionUser(r)
	if username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileName := r.FormValue("file_name")
	sourceFolder := r.FormValue("source_folder")
	targetFolder := r.FormValue("target_folder")

	if fileName == "" {
		http.Error(w, "Missing file name", 400)
		return
	}

	user := services.GetUser(username)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userStoragePath := services.GetUserStoragePath(username, user.UniqueCode)

	// Normalize folder paths
	if sourceFolder == "" || sourceFolder == "/" {
		sourceFolder = "/"
	}
	if targetFolder == "" || targetFolder == "/" {
		targetFolder = "/"
	}

	// Build source path
	var sourcePath string
	if sourceFolder != "" && sourceFolder != "/" {
		sourcePath = filepath.Join(userStoragePath, sourceFolder, fileName)
	} else {
		sourcePath = filepath.Join(userStoragePath, fileName)
	}

	// Build target path
	var targetPath string
	if targetFolder != "" && targetFolder != "/" {
		targetPath = filepath.Join(userStoragePath, targetFolder, fileName)
	} else {
		targetPath = filepath.Join(userStoragePath, fileName)
	}

	// Security checks
	if !isPathSafe(sourcePath, userStoragePath) || !isPathSafe(targetPath, userStoragePath) {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Lock for write operation
	services.LockUserFileWrite(username)
	defer services.UnlockUserFileWrite(username)

	// Get relative paths for database
	sourceRelPath, _ := filepath.Rel(userStoragePath, sourcePath)
	targetRelPath, _ := filepath.Rel(userStoragePath, targetPath)

	// Get file size before moving
	var fileSize int64
	if info, err := os.Stat(sourcePath); err == nil {
		fileSize = info.Size()
	}

	// Move file on disk
	err := os.Rename(sourcePath, targetPath)
	if err != nil {
		http.Error(w, "Failed to move file", 500)
		return
	}

	// Update database: change storage_path and parent_path
	query := `UPDATE files SET storage_path = ?, parent_path = ?, modified_at = datetime('now') WHERE username = ? AND storage_path = ?`
	_, err = services.GetDB().Exec(query, targetRelPath, targetFolder, username, sourceRelPath)
	if err != nil {
		// Rollback file move on database error
		os.Rename(targetPath, sourcePath)
		http.Error(w, "Failed to update file metadata", 500)
		return
	}

	// Update folder size caches
	var sourceFolderPath, targetFolderPath string
	if sourceFolder != "" && sourceFolder != "/" {
		sourceFolderPath = filepath.Join(userStoragePath, sourceFolder)
	} else {
		sourceFolderPath = userStoragePath
	}
	if targetFolder != "" && targetFolder != "/" {
		targetFolderPath = filepath.Join(userStoragePath, targetFolder)
	} else {
		targetFolderPath = userStoragePath
	}

	// Subtract from source, add to target
	services.UpdateFolderSize(sourceFolderPath, -fileSize)
	services.UpdateFolderSize(targetFolderPath, fileSize)

	// Invalidate cache after moving file
	services.InvalidateUserCache(username)

	// Redirect back
	if sourceFolder != "" && sourceFolder != "/" {
		http.Redirect(w, r, "/list?folder="+sourceFolder, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/list", http.StatusSeeOther)
	}
}

// getFolderList returns list of folders in a directory
func getFolderList(basePath string) ([]string, error) {
	var folders []string
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return folders, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			folders = append(folders, entry.Name())
		}
	}
	return folders, nil
}

// calculateFolderSizeInOps calculates folder size for file operations (DEPRECATED - now uses database)
// Kept for backward compatibility only
func calculateFolderSizeInOps(folderPath string) int64 {
	var totalSize int64
	filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	return totalSize
}
