package handlers

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/HAYASAKA7/HAYA_DISK/config"
	"github.com/HAYASAKA7/HAYA_DISK/middleware"
	"github.com/HAYASAKA7/HAYA_DISK/services"
	"github.com/HAYASAKA7/HAYA_DISK/utils"
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
		}

		filePath := filepath.Join(targetPath, header.Filename)

		// LOCK before file operations
		services.LockUserFileWrite(username)
		defer services.UnlockUserFileWrite(username)

		// Check if file already exists
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
		io.Copy(f, file)

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

	// Delete file or folder
	err := os.RemoveAll(targetPath)
	if err != nil {
		http.Error(w, "Delete error", 500)
		return
	}

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

	// Create folder
	err := os.MkdirAll(targetPath, os.ModePerm)
	if err != nil {
		http.Error(w, "Failed to create folder", 500)
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

	// Move file
	err := os.Rename(sourcePath, targetPath)
	if err != nil {
		http.Error(w, "Failed to move file", 500)
		return
	}

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
