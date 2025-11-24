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

	// Calculate storage statistics
	storageStats := calculateStorageStats(userStoragePath)

	data := map[string]interface{}{
		"files":         files,
		"username":      username,
		"currentFolder": currentFolder,
		"allFolders":    allFolders,
		"storageStats":  storageStats,
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

		// Categorize file
		category := categorizeFile(ext)
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

// categorizeFile returns the category for a file extension
func categorizeFile(ext string) string {
	imageExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".svg": true, ".webp": true, ".ico": true}
	videoExts := map[string]bool{".mp4": true, ".avi": true, ".mkv": true, ".mov": true, ".wmv": true, ".flv": true, ".webm": true, ".m4v": true}
	audioExts := map[string]bool{".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true, ".wma": true, ".m4a": true}
	docExts := map[string]bool{".pdf": true, ".doc": true, ".docx": true, ".txt": true, ".rtf": true, ".odt": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true}
	archiveExts := map[string]bool{".zip": true, ".rar": true, ".7z": true, ".tar": true, ".gz": true, ".bz2": true, ".xz": true}
	codeExts := map[string]bool{".go": true, ".js": true, ".py": true, ".java": true, ".cpp": true, ".c": true, ".h": true, ".html": true, ".css": true, ".json": true, ".xml": true, ".sql": true, ".sh": true, ".bat": true}

	if imageExts[ext] {
		return "Images"
	}
	if videoExts[ext] {
		return "Videos"
	}
	if audioExts[ext] {
		return "Audio"
	}
	if docExts[ext] {
		return "Documents"
	}
	if archiveExts[ext] {
		return "Archives"
	}
	if codeExts[ext] {
		return "Code"
	}
	return "Others"
}
