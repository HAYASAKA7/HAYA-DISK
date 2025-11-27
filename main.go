package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/HAYASAKA7/HAYA-DISK/config"
	"github.com/HAYASAKA7/HAYA-DISK/handlers"
	"github.com/HAYASAKA7/HAYA-DISK/middleware"
	"github.com/HAYASAKA7/HAYA-DISK/services"
	"github.com/HAYASAKA7/HAYA-DISK/utils"
)

// autoMigrate runs migration automatically if needed
func autoMigrate() {
	// Check if users.json exists
	if _, err := os.Stat(config.UsersFile); os.IsNotExist(err) {
		return // No migration needed
	}

	// Check if database is empty (no users)
	users, err := services.GetAllUsersDB()
	if err != nil || len(users) > 0 {
		return // Database already has data, skip migration
	}

	log.Println("Detected users.json - running automatic migration...")
	if err := utils.MigrateFromJSON(); err != nil {
		log.Printf("Warning: Auto-migration failed: %v", err)
		log.Println("You can run the migration tool manually: ./migrate.exe")
		return
	}
	log.Println("✓ Auto-migration completed successfully!")
}

func main() {
	// Create necessary directories
	os.MkdirAll(config.StorageDir, os.ModePerm)
	os.MkdirAll(config.TemplatesDir, os.ModePerm)

	// Initialize database (replaces LoadUsers)
	if err := services.InitDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer services.CloseDatabase()

	// Auto-migrate if users.json exists and database is empty
	autoMigrate()

	// Initialize and start backup scheduler
	backupScheduler := services.InitBackupService()
	backupScheduler.Start()
	defer backupScheduler.Stop()

	// Register HTTP handlers
	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/list", handlers.ListHandler)
	http.HandleFunc("/upload", middleware.RateLimitMiddleware(handlers.UploadHandler))
	http.HandleFunc("/download", handlers.DownloadHandler)
	http.HandleFunc("/delete", handlers.DeleteHandler)
	http.HandleFunc("/create-folder", handlers.CreateFolderHandler)
	http.HandleFunc("/move-file", handlers.MoveFileHandler)
	http.HandleFunc("/thumbnail", handlers.ThumbnailHandler)
	http.HandleFunc("/settings", handlers.SettingsHandler)
	http.HandleFunc("/api/get-user-info", handlers.APIGetUserInfoHandler)
	http.HandleFunc("/api/update-profile", handlers.APIUpdateProfileHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(config.TemplatesDir))))
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		fmt.Printf("Server started and accessible at:\n")
		fmt.Printf("  - Local: http://localhost:8080\n")
		fmt.Printf("  - Network: http://<your-ip>:8080\n")
		fmt.Printf("Listening on %s\n", config.ServerPort)
		if err := http.ListenAndServe(config.ServerPort, nil); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-stop
	log.Println("Shutting down server...")
}
