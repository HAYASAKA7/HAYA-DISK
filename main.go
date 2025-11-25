package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/HAYASAKA7/HAYA_DISK/config"
	"github.com/HAYASAKA7/HAYA_DISK/handlers"
	"github.com/HAYASAKA7/HAYA_DISK/middleware"
	"github.com/HAYASAKA7/HAYA_DISK/services"
)

func main() {
	// Create necessary directories
	os.MkdirAll(config.StorageDir, os.ModePerm)
	os.MkdirAll(config.TemplatesDir, os.ModePerm)

	// Load existing users
	services.LoadUsers()

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

	// Start server
	fmt.Printf("Server started and accessible at:\n")
	fmt.Printf("  - Local: http://localhost:8080\n")
	fmt.Printf("  - Network: http://<your-ip>:8080\n")
	fmt.Printf("Listening on %s\n", config.ServerPort)
	if err := http.ListenAndServe(config.ServerPort, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
