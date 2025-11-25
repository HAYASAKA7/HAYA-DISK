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
			http.Error(w, "Unauthorized", http.StatusForbidden)
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

	// Calculate storage statistics
	storageStats := calculateStorageStats(userStoragePath)

	// Get recent files
	recentFiles := getRecentFiles(userStoragePath)

	// Determine if we're on the home page (no folder selected)
	isHomePage := currentFolder == ""

	data := map[string]interface{}{
		"files":         files,
		"username":      username,
		"currentFolder": currentFolder,
		"allFolders":    allFolders,
		"storageStats":  storageStats,
		"recentFiles":   recentFiles,
		"isHomePage":    isHomePage,
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

			// Get folder size from cache or calculate if not cached
			fullFolderPath := filepath.Join(targetPath, f.Name())
			folderSize, cached := services.GetFolderSize(fullFolderPath)
			if !cached {
				// Calculate and cache the folder size
				folderSize = services.RecalculateFolderSize(fullFolderPath, calculateFolderSize)
			}

			fileList = append(fileList, models.FileInfo{
				Name:     f.Name(),
				Size:     utils.FormatFileSize(folderSize),
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

// calculateFolderSize calculates the total size of all files in a folder recursively
func calculateFolderSize(folderPath string) int64 {
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

// calculateStorageStats analyzes storage usage by file type
func calculateStorageStats(basePath string) models.StorageStats {
	stats := models.StorageStats{
		FileTypes: []models.FileTypeStats{},
	}

	typeMap := make(map[string]*models.FileTypeStats)
	colorMap := map[string]string{
		"Images":    "#4CAF50",
		"Videos":    "#2196F3",
		"Audio":     "#FF9800",
		"Documents": "#9C27B0",
		"Archives":  "#F44336",
		"Code":      "#00BCD4",
		"Others":    "#9E9E9E",
	}

	// Initialize categories
	categories := []string{"Images", "Videos", "Audio", "Documents", "Archives", "Code", "Others"}
	for _, cat := range categories {
		typeMap[cat] = &models.FileTypeStats{
			Type:  cat,
			Color: colorMap[cat],
		}
	}

	// Walk through all files recursively
	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		size := info.Size()
		stats.TotalSize += size

		// Categorize file using utility function
		category := utils.GetFileCategory(ext)
		if stat, exists := typeMap[category]; exists {
			stat.Size += size
			stat.Count++
		}

		return nil
	})

	// Calculate percentages and format sizes
	stats.TotalSizeStr = utils.FormatFileSize(stats.TotalSize)

	for _, cat := range categories {
		stat := typeMap[cat]
		if stat.Count > 0 {
			stat.SizeStr = utils.FormatFileSize(stat.Size)
			if stats.TotalSize > 0 {
				stat.Percentage = float64(stat.Size) / float64(stats.TotalSize) * 100
			}
			stats.FileTypes = append(stats.FileTypes, *stat)
		}
	}

	return stats
}

// getRecentFiles retrieves the 5 most recently modified files
func getRecentFiles(basePath string) []models.RecentFile {
	type fileWithTime struct {
		name    string
		path    string
		modTime int64
		size    int64
		isImage bool
		ext     string
	}

	var allFiles []fileWithTime

	// Walk through all files recursively
	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Get relative path from base
		relPath, _ := filepath.Rel(basePath, path)
		ext := strings.ToLower(filepath.Ext(info.Name()))

		allFiles = append(allFiles, fileWithTime{
			name:    info.Name(),
			path:    filepath.ToSlash(relPath),
			modTime: info.ModTime().Unix(),
			size:    info.Size(),
			isImage: utils.IsImageFile(ext),
			ext:     ext,
		})

		return nil
	})

	// Sort by modification time (newest first)
	for i := 0; i < len(allFiles)-1; i++ {
		for j := i + 1; j < len(allFiles); j++ {
			if allFiles[j].modTime > allFiles[i].modTime {
				allFiles[i], allFiles[j] = allFiles[j], allFiles[i]
			}
		}
	}

	// Take top 5
	var recentFiles []models.RecentFile
	count := 5
	if len(allFiles) < count {
		count = len(allFiles)
	}

	for i := 0; i < count; i++ {
		f := allFiles[i]
		recentFiles = append(recentFiles, models.RecentFile{
			Name:    f.name,
			Path:    f.path,
			Icon:    utils.GetFileIcon(f.ext),
			Size:    utils.FormatFileSize(f.size),
			IsImage: f.isImage,
		})
	}

	return recentFiles
}
