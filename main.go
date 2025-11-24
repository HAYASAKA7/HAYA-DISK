package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/user/Test1/config"
	"github.com/user/Test1/handlers"
	"github.com/user/Test1/services"
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
	http.HandleFunc("/upload", handlers.UploadHandler)
	http.HandleFunc("/download", handlers.DownloadHandler)
	http.HandleFunc("/thumbnail", handlers.ThumbnailHandler)
	http.HandleFunc("/settings", handlers.SettingsHandler)
	http.HandleFunc("/api/update-profile", handlers.APIUpdateProfileHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(config.TemplatesDir))))

	// Start server
	fmt.Printf("Server started at http://localhost%s\n", config.ServerPort)
	if err := http.ListenAndServe(config.ServerPort, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
