package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/Test1/config"
	"github.com/user/Test1/middleware"
	"github.com/user/Test1/services"
	"github.com/user/Test1/utils"
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
		http.ServeFile(w, r, filepath.Join(config.TemplatesDir, "upload.html"))
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

		userStoragePath := services.GetUserStoragePath(username, user.UniqueCode)
		filePath := filepath.Join(userStoragePath, header.Filename)
		f, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Save error", 500)
			return
		}
		defer f.Close()
		io.Copy(f, file)
		http.Redirect(w, r, "/list", http.StatusSeeOther)
	}
}

// DownloadHandler handles file downloads
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetSessionUser(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	name := r.URL.Query().Get("name")
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
	filePath := filepath.Join(userStoragePath, name)

	// Security check: ensure path is within user's storage
	if !isPathSafe(filePath, userStoragePath) {
		http.Error(w, "Unauthorized", 403)
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
		http.Error(w, "Unauthorized", 401)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing file name", 400)
		return
	}

	user := services.GetUser(username)
	if user == nil {
		http.Error(w, "Unauthorized", 401)
		return
	}

	ext := strings.ToLower(filepath.Ext(name))
	if !utils.IsImageFile(ext) {
		http.Error(w, "Not an image file", 400)
		return
	}

	userStoragePath := services.GetUserStoragePath(username, user.UniqueCode)
	filePath := filepath.Join(userStoragePath, name)

	// Security check
	if !isPathSafe(filePath, userStoragePath) {
		http.Error(w, "Unauthorized", 403)
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
