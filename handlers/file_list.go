package handlers

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/HAYASAKA7/HAYA-DISK/config"
	"github.com/HAYASAKA7/HAYA-DISK/middleware"
	"github.com/HAYASAKA7/HAYA-DISK/models"
	"github.com/HAYASAKA7/HAYA-DISK/services"
	"github.com/HAYASAKA7/HAYA-DISK/utils"
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
	if currentFolder == "" {
		currentFolder = "/"
	}

	// Get files from database
	files, err := getFileList(username, currentFolder)
	if err != nil {
		http.Error(w, "Unable to list files", 500)
		return
	}

	// Get all folders for move functionality
	allFolders, _ := getAllFoldersFromDB(username)

	// Calculate storage statistics from database
	storageStats := calculateStorageStatsFromDB(username)

	// Get recent files from database
	recentFiles := getRecentFilesFromDB(username)

	// Determine if we're on the home page (no folder selected)
	isHomePage := currentFolder == "/" || currentFolder == ""

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

// getFileList retrieves list of files and folders from database
func getFileList(username string, currentFolder string) ([]models.FileInfo, error) {
	var fileList []models.FileInfo

	// Normalize folder path
	if currentFolder == "" {
		currentFolder = "/"
	}

	// Get files from database
	fileMetadata, err := services.GetUserFiles(username, currentFolder)
	if err != nil {
		return fileList, err
	}

	// Convert database metadata to FileInfo
	for _, meta := range fileMetadata {
		ext := strings.ToLower(filepath.Ext(meta.Filename))
		isImage := utils.IsImageFile(ext)

		// Build relative path for display
		var displayPath string
		if meta.ParentPath == "/" || meta.ParentPath == "" {
			displayPath = meta.Filename
		} else {
			displayPath = filepath.Join(meta.ParentPath, meta.Filename)
		}

		// Calculate size - for folders, calculate recursively from database
		var sizeToDisplay int64
		if meta.IsDirectory {
			sizeToDisplay = calculateFolderSize(username, displayPath)
		} else {
			sizeToDisplay = meta.FileSize
		}

		fileInfo := models.FileInfo{
			Name:     meta.Filename,
			Size:     utils.FormatFileSize(sizeToDisplay),
			Modified: meta.ModifiedAt.Format("2006-01-02 15:04"),
			IsDir:    meta.IsDirectory,
			Path:     displayPath,
			IsImage:  isImage,
			Ext:      ext,
		}

		if meta.IsDirectory {
			fileInfo.Icon = "📁"
		} else {
			fileInfo.Icon = utils.GetFileIcon(ext)
		}

		fileList = append(fileList, fileInfo)
	}

	return fileList, nil
}

// calculateFolderSize calculates the total size of all files in a folder recursively using database
func calculateFolderSize(username, folderPath string) int64 {
	size, err := services.CalculateFolderSizeDB(username, folderPath)
	if err != nil {
		return 0
	}
	return size
}

// calculateFolderSizeOld calculates the total size of all files in a folder recursively (DEPRECATED - filesystem based)
func calculateFolderSizeOld(folderPath string) int64 {
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

// getAllFolders recursively gets all folders for the move functionality (DEPRECATED - kept for compatibility)
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

// getAllFoldersFromDB gets all folders from database for the move functionality
func getAllFoldersFromDB(username string) ([]string, error) {
	folders := []string{"/"}

	folderMetadata, err := services.GetAllFoldersDB(username)
	if err != nil {
		return folders, err
	}

	for _, folder := range folderMetadata {
		// Build folder display path
		var folderPath string
		if folder.ParentPath == "/" || folder.ParentPath == "" {
			folderPath = folder.Filename
		} else {
			folderPath = filepath.Join(folder.ParentPath, folder.Filename)
		}
		folders = append(folders, folderPath)
	}

	return folders, nil
}

// calculateStorageStats analyzes storage usage by file type (DEPRECATED - kept for compatibility)
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

// calculateStorageStatsFromDB analyzes storage usage by file type using database
func calculateStorageStatsFromDB(username string) models.StorageStats {
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

	// Get all files from database (search with empty query to get all)
	allFiles, err := services.SearchUserFiles(username, "")
	if err == nil {
		for _, file := range allFiles {
			if !file.IsDirectory {
				ext := strings.ToLower(filepath.Ext(file.Filename))
				stats.TotalSize += file.FileSize

				// Categorize file using utility function
				category := utils.GetFileCategory(ext)
				if stat, exists := typeMap[category]; exists {
					stat.Size += file.FileSize
					stat.Count++
				}
			}
		}
	}

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

// getRecentFiles retrieves the 5 most recently modified files (DEPRECATED - kept for compatibility)
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

		// Extract folder path (directory containing the file)
		folderPath := filepath.Dir(f.path)
		if folderPath == "." {
			folderPath = "" // Root folder
		}
		folderPath = filepath.ToSlash(folderPath)

		recentFiles = append(recentFiles, models.RecentFile{
			Name:       f.name,
			Path:       f.path,
			FolderPath: folderPath,
			Icon:       utils.GetFileIcon(f.ext),
			Size:       utils.FormatFileSize(f.size),
			IsImage:    f.isImage,
		})
	}

	return recentFiles
}

// getRecentFilesFromDB retrieves the 5 most recently uploaded files from database
func getRecentFilesFromDB(username string) []models.RecentFile {
	var recentFiles []models.RecentFile

	// Query to get recent files, ordered by uploaded_at DESC
	query := `SELECT filename, storage_path, parent_path, file_size, uploaded_at 
			  FROM files WHERE username = ? AND is_directory = 0 
			  ORDER BY uploaded_at DESC LIMIT 5`

	rows, err := services.GetDB().Query(query, username)
	if err != nil {
		return recentFiles
	}
	defer rows.Close()

	for rows.Next() {
		var filename, storagePath, parentPath string
		var fileSize int64
		var uploadedAt string

		rows.Scan(&filename, &storagePath, &parentPath, &fileSize, &uploadedAt)

		ext := strings.ToLower(filepath.Ext(filename))
		isImage := utils.IsImageFile(ext)

		// Build display path
		var displayPath string
		if parentPath == "/" || parentPath == "" {
			displayPath = filename
		} else {
			displayPath = filepath.Join(parentPath, filename)
		}

		recentFiles = append(recentFiles, models.RecentFile{
			Name:       filename,
			Path:       displayPath,
			FolderPath: parentPath,
			Icon:       utils.GetFileIcon(ext),
			Size:       utils.FormatFileSize(fileSize),
			IsImage:    isImage,
		})
	}

	return recentFiles
}
