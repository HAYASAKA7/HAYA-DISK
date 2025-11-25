package utils

import (
	"fmt"
	"regexp"
)

// IsValidEmail validates email format
func IsValidEmail(email string) bool {
	if email == "" {
		return true // Email is optional
	}
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}

// IsValidPhone validates phone format
func IsValidPhone(phone string) bool {
	if phone == "" {
		return true // Phone is optional
	}
	pattern := `^[0-9]{10,15}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(phone)
}

// FormatFileSize formats bytes into human-readable size
func FormatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// IsImageFile checks if file extension is an image
func IsImageFile(ext string) bool {
	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".bmp": true, ".svg": true, ".webp": true, ".ico": true,
	}
	return imageExts[ext]
}

// GetFileCategory returns the category for a file extension
func GetFileCategory(ext string) string {
	imageExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".svg": true, ".webp": true, ".ico": true}
	videoExts := map[string]bool{".mp4": true, ".avi": true, ".mkv": true, ".mov": true, ".wmv": true, ".flv": true, ".webm": true, ".m4v": true}
	audioExts := map[string]bool{".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true, ".wma": true, ".m4a": true, ".opus": true}
	docExts := map[string]bool{".pdf": true, ".doc": true, ".docx": true, ".txt": true, ".rtf": true, ".odt": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true, ".md": true}
	archiveExts := map[string]bool{".zip": true, ".rar": true, ".7z": true, ".tar": true, ".gz": true, ".bz2": true, ".xz": true}
	codeExts := map[string]bool{".go": true, ".js": true, ".py": true, ".java": true, ".cpp": true, ".c": true, ".h": true, ".html": true, ".css": true, ".json": true, ".xml": true, ".sql": true, ".sh": true, ".bat": true, ".rs": true, ".ts": true, ".php": true, ".rb": true}

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

// GetFileIcon returns an emoji icon based on file category
func GetFileIcon(ext string) string {
	category := GetFileCategory(ext)

	iconMap := map[string]string{
		"Images":    "🖼️",
		"Videos":    "🎥",
		"Audio":     "🎵",
		"Documents": "📄",
		"Archives":  "📦",
		"Code":      "💻",
		"Others":    "📁",
	}

	return iconMap[category]
}

// GetImageContentType returns the MIME type for image extensions
func GetImageContentType(ext string) string {
	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".bmp":  "image/bmp",
	}
	if ct, exists := contentTypes[ext]; exists {
		return ct
	}
	return "image/jpeg"
}
