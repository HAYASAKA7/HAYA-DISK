package services

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/HAYASAKA7/HAYA-DISK/config"
)

// BackupScheduler manages scheduled backups
type BackupScheduler struct {
	settings   config.BackupSettings
	stopChan   chan struct{}
	wg         sync.WaitGroup
	isRunning  bool
	mu         sync.Mutex
	lastBackup time.Time
	lastError  error
}

// BackupResult holds the result of a backup operation
type BackupResult struct {
	Success    bool
	BackupPath string
	StartTime  time.Time
	EndTime    time.Time
	FilesCount int
	TotalSize  int64
	Error      error
}

var (
	backupScheduler *BackupScheduler
	backupMu        sync.Mutex
)

// InitBackupService initializes the backup service with default settings
func InitBackupService() *BackupScheduler {
	return InitBackupServiceWithSettings(config.DefaultBackupSettings)
}

// InitBackupServiceWithSettings initializes the backup service with custom settings
func InitBackupServiceWithSettings(settings config.BackupSettings) *BackupScheduler {
	backupMu.Lock()
	defer backupMu.Unlock()

	backupScheduler = &BackupScheduler{
		settings: settings,
		stopChan: make(chan struct{}),
	}

	// Create backup directory if it doesn't exist
	if settings.Enabled {
		if err := os.MkdirAll(settings.BackupDir, os.ModePerm); err != nil {
			log.Printf("Warning: Failed to create backup directory: %v", err)
		}
	}

	return backupScheduler
}

// GetBackupScheduler returns the global backup scheduler instance
func GetBackupScheduler() *BackupScheduler {
	backupMu.Lock()
	defer backupMu.Unlock()
	return backupScheduler
}

// Start begins the backup scheduler
func (bs *BackupScheduler) Start() {
	bs.mu.Lock()
	if bs.isRunning {
		bs.mu.Unlock()
		return
	}
	bs.isRunning = true
	bs.mu.Unlock()

	if !bs.settings.Enabled {
		log.Println("Backup service is disabled")
		return
	}

	// Check if any backup exists, if not, run immediately in background
	if !bs.hasExistingBackup() {
		log.Println("No existing backup found. Starting initial backup in background...")
		go func() {
			result := bs.RunBackup()
			if result.Success {
				log.Printf("✓ Initial backup completed: %s (Size: %s)",
					result.BackupPath, formatSize(result.TotalSize))
			} else {
				log.Printf("✗ Initial backup failed: %v", result.Error)
			}
		}()
	}

	bs.wg.Add(1)
	go bs.run()

	nextBackup := bs.settings.GetNextBackupTime()
	log.Printf("✓ Backup scheduler started. Next backup at: %s", nextBackup.Format("2006-01-02 15:04:05"))
}

// hasExistingBackup checks if any backup already exists
func (bs *BackupScheduler) hasExistingBackup() bool {
	entries, err := os.ReadDir(bs.settings.BackupDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "backup_") {
			return true
		}
	}
	return false
}

// Stop gracefully stops the backup scheduler
func (bs *BackupScheduler) Stop() {
	bs.mu.Lock()
	if !bs.isRunning {
		bs.mu.Unlock()
		return
	}
	bs.isRunning = false
	bs.mu.Unlock()

	close(bs.stopChan)
	bs.wg.Wait()
	log.Println("Backup scheduler stopped")
}

// run is the main scheduler loop
func (bs *BackupScheduler) run() {
	defer bs.wg.Done()

	for {
		duration := bs.settings.GetDurationUntilNextBackup()
		log.Printf("Next backup scheduled in: %v", duration.Round(time.Minute))

		timer := time.NewTimer(duration)

		select {
		case <-bs.stopChan:
			timer.Stop()
			return
		case <-timer.C:
			log.Println("Starting scheduled backup...")
			result := bs.RunBackup()
			if result.Success {
				log.Printf("✓ Backup completed successfully: %s (Files: %d, Size: %s)",
					result.BackupPath, result.FilesCount, formatSize(result.TotalSize))
			} else {
				log.Printf("✗ Backup failed: %v", result.Error)
			}

			// Clean old backups after successful backup
			if result.Success {
				bs.CleanOldBackups()
			}
		}
	}
}

// RunBackup executes a backup immediately
func (bs *BackupScheduler) RunBackup() BackupResult {
	result := BackupResult{
		StartTime: time.Now(),
	}

	// Create timestamp-based backup folder name
	timestamp := result.StartTime.Format("2006-01-02_150405")
	backupName := fmt.Sprintf("backup_%s", timestamp)

	var backupPath string
	var err error

	if bs.settings.CompressBackup {
		backupPath, err = bs.createCompressedBackup(backupName)
	} else {
		backupPath, err = bs.createDirectoryBackup(backupName)
	}

	result.EndTime = time.Now()

	if err != nil {
		result.Error = err
		result.Success = false
		bs.lastError = err
		return result
	}

	result.Success = true
	result.BackupPath = backupPath
	bs.lastBackup = result.StartTime
	bs.lastError = nil

	// Get backup stats
	if bs.settings.CompressBackup {
		if info, err := os.Stat(backupPath); err == nil {
			result.TotalSize = info.Size()
		}
	}

	// Log to backup history
	bs.logBackup(result)

	return result
}

// createCompressedBackup creates a zip archive of the backup
func (bs *BackupScheduler) createCompressedBackup(backupName string) (string, error) {
	zipPath := filepath.Join(bs.settings.BackupDir, backupName+".zip")

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Backup database
	if bs.settings.BackupDatabase {
		if err := bs.addFileToZip(zipWriter, "./haya-disk.db", "haya-disk.db"); err != nil {
			return "", fmt.Errorf("failed to backup database: %w", err)
		}
		log.Println("  - Database backed up")
	}

	// Backup storage
	if bs.settings.BackupStorage {
		storageDir := config.StorageDir
		if err := bs.addDirectoryToZip(zipWriter, storageDir, "storage"); err != nil {
			return "", fmt.Errorf("failed to backup storage: %w", err)
		}
		log.Println("  - Storage backed up")
	}

	return zipPath, nil
}

// createDirectoryBackup creates an uncompressed directory backup
func (bs *BackupScheduler) createDirectoryBackup(backupName string) (string, error) {
	backupPath := filepath.Join(bs.settings.BackupDir, backupName)

	if err := os.MkdirAll(backupPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Backup database
	if bs.settings.BackupDatabase {
		srcDB := "./haya-disk.db"
		dstDB := filepath.Join(backupPath, "haya-disk.db")
		if err := copyFile(srcDB, dstDB); err != nil {
			return "", fmt.Errorf("failed to backup database: %w", err)
		}
		log.Println("  - Database backed up")
	}

	// Backup storage
	if bs.settings.BackupStorage {
		srcStorage := config.StorageDir
		dstStorage := filepath.Join(backupPath, "storage")
		if err := copyDirectory(srcStorage, dstStorage); err != nil {
			return "", fmt.Errorf("failed to backup storage: %w", err)
		}
		log.Println("  - Storage backed up")
	}

	return backupPath, nil
}

// addFileToZip adds a single file to the zip archive
func (bs *BackupScheduler) addFileToZip(zipWriter *zip.Writer, srcPath, zipPath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = zipPath
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

// addDirectoryToZip recursively adds a directory to the zip archive
func (bs *BackupScheduler) addDirectoryToZip(zipWriter *zip.Writer, srcDir, zipDir string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create relative path for zip
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		zipPath := filepath.Join(zipDir, relPath)
		// Normalize path separators for zip
		zipPath = strings.ReplaceAll(zipPath, "\\", "/")

		if info.IsDir() {
			// Add directory entry
			if relPath != "." {
				_, err := zipWriter.Create(zipPath + "/")
				return err
			}
			return nil
		}

		// Add file
		return bs.addFileToZip(zipWriter, path, zipPath)
	})
}

// CleanOldBackups removes backups older than the retention period
func (bs *BackupScheduler) CleanOldBackups() {
	if bs.settings.RetentionDays <= 0 {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -bs.settings.RetentionDays)
	entries, err := os.ReadDir(bs.settings.BackupDir)
	if err != nil {
		log.Printf("Warning: Failed to read backup directory: %v", err)
		return
	}

	var deletedCount int
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "backup_") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			path := filepath.Join(bs.settings.BackupDir, entry.Name())
			if entry.IsDir() {
				err = os.RemoveAll(path)
			} else {
				err = os.Remove(path)
			}
			if err != nil {
				log.Printf("Warning: Failed to delete old backup %s: %v", entry.Name(), err)
			} else {
				deletedCount++
				log.Printf("  - Deleted old backup: %s", entry.Name())
			}
		}
	}

	if deletedCount > 0 {
		log.Printf("✓ Cleaned up %d old backup(s)", deletedCount)
	}
}

// logBackup writes backup information to the log file
func (bs *BackupScheduler) logBackup(result BackupResult) {
	logPath := filepath.Join(bs.settings.BackupDir, "backup_log.txt")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Warning: Failed to write backup log: %v", err)
		return
	}
	defer f.Close()

	status := "SUCCESS"
	if !result.Success {
		status = "FAILED"
	}

	logEntry := fmt.Sprintf("[%s] %s - Path: %s, Duration: %v, Size: %s\n",
		result.StartTime.Format("2006-01-02 15:04:05"),
		status,
		result.BackupPath,
		result.EndTime.Sub(result.StartTime).Round(time.Second),
		formatSize(result.TotalSize),
	)

	f.WriteString(logEntry)
}

// GetStatus returns the current backup status
func (bs *BackupScheduler) GetStatus() map[string]interface{} {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	status := map[string]interface{}{
		"enabled":       bs.settings.Enabled,
		"isRunning":     bs.isRunning,
		"scheduleTime":  fmt.Sprintf("%02d:%02d", bs.settings.ScheduleHour, bs.settings.ScheduleMinute),
		"retentionDays": bs.settings.RetentionDays,
		"backupDir":     bs.settings.BackupDir,
		"compressed":    bs.settings.CompressBackup,
	}

	if !bs.lastBackup.IsZero() {
		status["lastBackup"] = bs.lastBackup.Format("2006-01-02 15:04:05")
	}

	if bs.lastError != nil {
		status["lastError"] = bs.lastError.Error()
	}

	if bs.settings.Enabled && bs.isRunning {
		status["nextBackup"] = bs.settings.GetNextBackupTime().Format("2006-01-02 15:04:05")
	}

	return status
}

// ListBackups returns a list of available backups
func (bs *BackupScheduler) ListBackups() ([]map[string]interface{}, error) {
	entries, err := os.ReadDir(bs.settings.BackupDir)
	if err != nil {
		return nil, err
	}

	var backups []map[string]interface{}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "backup_") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backup := map[string]interface{}{
			"name":      entry.Name(),
			"size":      info.Size(),
			"sizeHuman": formatSize(info.Size()),
			"createdAt": info.ModTime().Format("2006-01-02 15:04:05"),
			"isDir":     entry.IsDir(),
		}
		backups = append(backups, backup)
	}

	// Sort by name (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i]["name"].(string) > backups[j]["name"].(string)
	})

	return backups, nil
}

// ==================== Helper Functions ====================

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return dstFile.Sync()
}

// copyDirectory recursively copies a directory
func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// formatSize formats bytes to human readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
