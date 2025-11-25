package config

const (
	StorageDir   = "storage"
	TemplatesDir = "templates"
	UsersFile    = "users.json"
	ServerPort   = "0.0.0.0:8080"
	SessionAge   = 30 * 24 * 60 * 60 // 30 days in seconds

	// Performance tuning
	MaxConcurrentUploads = 10
	ReaderBufferSize     = 32 * 1024 // 32 KB
	WriteBufferSize      = 32 * 1024 // 32 KB
)
