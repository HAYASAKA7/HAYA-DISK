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
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".bmp":  true,
	}
	return imageExts[ext]
}

// GetFileIcon returns an emoji icon for file type
func GetFileIcon(ext string) string {
	iconMap := map[string]string{
		".pdf":  "📄",
		".doc":  "📝",
		".docx": "📝",
		".xls":  "📊",
		".xlsx": "📊",
		".ppt":  "🎬",
		".pptx": "🎬",
		".zip":  "📦",
		".rar":  "📦",
		".7z":   "📦",
		".mp4":  "🎥",
		".avi":  "🎥",
		".mkv":  "🎥",
		".mov":  "🎥",
		".mp3":  "🎵",
		".wav":  "🎵",
		".flac": "🎵",
		".txt":  "📋",
		".md":   "📋",
		".jpg":  "🖼️",
		".jpeg": "🖼️",
		".png":  "🖼️",
		".gif":  "🖼️",
		".webp": "🖼️",
	}
	if icon, exists := iconMap[ext]; exists {
		return icon
	}
	return "📁"
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
