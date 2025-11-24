package handlers

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/Test1/config"
	"github.com/user/Test1/middleware"
	"github.com/user/Test1/models"
	"github.com/user/Test1/services"
	"github.com/user/Test1/utils"
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

	userStoragePath := services.GetUserStoragePath(username, user.UniqueCode)
	files, err := getFileList(userStoragePath)
	if err != nil {
		http.Error(w, "Unable to list files", 500)
		return
	}

	data := map[string]interface{}{
		"files":    files,
		"username": username,
	}

	tmpl, err := template.ParseFiles(filepath.Join(config.TemplatesDir, "list.html"))
	if err != nil {
		http.Error(w, "Template error", 500)
		return
	}
	tmpl.Execute(w, data)
}

// getFileList retrieves list of files in a directory
func getFileList(userStoragePath string) ([]models.FileInfo, error) {
	var fileList []models.FileInfo
	entries, err := os.ReadDir(userStoragePath)
	if err != nil {
		return fileList, err
	}

	for _, f := range entries {
		if !f.IsDir() {
			info, _ := f.Info()
			ext := strings.ToLower(filepath.Ext(f.Name()))
			isImage := utils.IsImageFile(ext)
			fileList = append(fileList, models.FileInfo{
				Name:     f.Name(),
				Size:     utils.FormatFileSize(info.Size()),
				Modified: info.ModTime().Format("2006-01-02 15:04"),
				Icon:     utils.GetFileIcon(ext),
				IsImage:  isImage,
				Ext:      ext,
			})
		}
	}
	return fileList, nil
}
