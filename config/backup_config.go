package config

import "time"

// BackupSettings defines the configuration for auto-backup
type BackupSettings struct {
	Enabled        bool   // Enable/disable auto-backup
	BackupDir      string // Backup destination directory
	ScheduleHour   int    // Hour to run backup (0-23)
	ScheduleMinute int    // Minute to run backup (0-59)
	RetentionDays  int    // Keep backups for X days
	BackupDatabase bool   // Backup the SQLite database
	BackupStorage  bool   // Backup the storage folder
	CompressBackup bool   // Compress backups to .zip
}

// DefaultBackupSettings returns the default backup configuration
// Scheduled at 3:00 AM device time, 7 days retention, with compression
var DefaultBackupSettings = BackupSettings{
	Enabled:        true,
	BackupDir:      "backups",
	ScheduleHour:   3, // 3 AM
	ScheduleMinute: 0, // :00
	RetentionDays:  7, // Keep for 7 days
	BackupDatabase: true,
	BackupStorage:  true,
	CompressBackup: true,
}

// GetNextBackupTime calculates the next scheduled backup time
func (s *BackupSettings) GetNextBackupTime() time.Time {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), s.ScheduleHour, s.ScheduleMinute, 0, 0, now.Location())

	// If the scheduled time has already passed today, schedule for tomorrow
	if now.After(next) {
		next = next.Add(24 * time.Hour)
	}

	return next
}

// GetDurationUntilNextBackup returns the duration until the next backup
func (s *BackupSettings) GetDurationUntilNextBackup() time.Duration {
	return time.Until(s.GetNextBackupTime())
}
