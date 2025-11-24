package handlers

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/HAYASAKA7/HAYA_DISK/config"
	"github.com/HAYASAKA7/HAYA_DISK/middleware"
	"github.com/HAYASAKA7/HAYA_DISK/models"
	"github.com/HAYASAKA7/HAYA_DISK/services"
	"github.com/HAYASAKA7/HAYA_DISK/utils"
)

// ListHandler displays user's files
func ListHandler(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetSessionUser(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := services.GetUser(username)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get current folder from query parameter
	currentFolder := r.URL.Query().Get("folder")

	userStoragePath := services.GetUserStoragePath(username, user.UniqueCode)

	// Build target path
	var targetPath string
	if currentFolder != "" && currentFolder != "/" {
		targetPath = filepath.Join(userStoragePath, currentFolder)
		// Security check
		absPath, _ := filepath.Abs(targetPath)
		absAllowed, _ := filepath.Abs(userStoragePath)
		if !strings.HasPrefix(absPath, absAllowed) {
			http.Error(w, "Unauthorized", 403)
			return
		}
	} else {
		targetPath = userStoragePath
		currentFolder = ""
	}

	files, err := getFileList(targetPath, currentFolder)
	if err != nil {
		http.Error(w, "Unable to list files", 500)
		return
	}

	// Get all folders for move functionality
	allFolders, _ := getAllFolders(userStoragePath, "")

	data := map[string]interface{}{
		"files":         files,
		"username":      username,
		"currentFolder": currentFolder,
		"allFolders":    allFolders,
	}

	tmpl, err := template.ParseFiles(filepath.Join(config.TemplatesDir, "list.html"))
	if err != nil {
		http.Error(w, "Template error", 500)
		return
	}
	tmpl.Execute(w, data)
}

// getFileList retrieves list of files and folders in a directory
func getFileList(targetPath string, currentFolder string) ([]models.FileInfo, error) {
	var fileList []models.FileInfo
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return fileList, err
	}

	// Add folders first
	for _, f := range entries {
		if f.IsDir() {
			info, _ := f.Info()
			var folderPath string
			if currentFolder != "" {
				folderPath = filepath.Join(currentFolder, f.Name())
			} else {
				folderPath = f.Name()
			}
			fileList = append(fileList, models.FileInfo{
				Name:     f.Name(),
				Size:     "-",
				Modified: info.ModTime().Format("2006-01-02 15:04"),
				Icon:     "📁",
				IsImage:  false,
				IsDir:    true,
				Path:     folderPath,
			})
		}
	}

	// Then add files
	for _, f := range entries {
		if !f.IsDir() {
			info, _ := f.Info()
			ext := strings.ToLower(filepath.Ext(f.Name()))
			isImage := utils.IsImageFile(ext)
			var filePath string
			if currentFolder != "" {
				filePath = filepath.Join(currentFolder, f.Name())
			} else {
				filePath = f.Name()
			}
			fileList = append(fileList, models.FileInfo{
				Name:     f.Name(),
				Size:     utils.FormatFileSize(info.Size()),
				Modified: info.ModTime().Format("2006-01-02 15:04"),
				Icon:     utils.GetFileIcon(ext),
				IsImage:  isImage,
				Ext:      ext,
				IsDir:    false,
				Path:     filePath,
			})
		}
	}
	return fileList, nil
}

// getAllFolders recursively gets all folders for the move functionality
func getAllFolders(basePath string, prefix string) ([]string, error) {
	var folders []string
	folders = append(folders, "/") // Root folder

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return folders, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			var folderPath string
			if prefix != "" {
				folderPath = filepath.Join(prefix, entry.Name())
			} else {
				folderPath = entry.Name()
			}
			folders = append(folders, folderPath)

			// Recursively get subfolders
			subFolders, _ := getAllFolders(filepath.Join(basePath, entry.Name()), folderPath)
			for _, sub := range subFolders {
				if sub != "/" {
					folders = append(folders, sub)
				}
			}
		}
	}
	return folders, nil
}
