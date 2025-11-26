package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/HAYASAKA7/HAYA-DISK/config"
	"github.com/HAYASAKA7/HAYA-DISK/middleware"
	"github.com/HAYASAKA7/HAYA-DISK/services"
	"github.com/HAYASAKA7/HAYA-DISK/utils"
)

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetSessionUser(r)
	if username != "" {
		http.Redirect(w, r, "/list", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		credential := r.FormValue("credential") // Can be username, email, or phone
		password := r.FormValue("password")

		user := services.FindUserByCredential(credential)

		if user == nil || user.Password != password {
			tmpl, _ := template.ParseFiles(filepath.Join(config.TemplatesDir, "login.html"))
			tmpl.Execute(w, map[string]string{"error": "Invalid credentials or password"})
			return
		}

		middleware.SetSessionCookie(w, user.Username)
		http.Redirect(w, r, "/list", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles(filepath.Join(config.TemplatesDir, "login.html"))
	if err != nil {
		http.Error(w, "Template error", 500)
		return
	}
	tmpl.Execute(w, nil)
}

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetSessionUser(r)
	if username != "" {
		http.Redirect(w, r, "/list", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		username := strings.TrimSpace(r.FormValue("username"))
		email := strings.TrimSpace(r.FormValue("email"))
		phone := strings.TrimSpace(r.FormValue("phone"))
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		// Validation
		var errorMsg string
		if username == "" || password == "" {
			errorMsg = "Username and password are required"
		} else if email == "" && phone == "" {
			errorMsg = "Please provide either email or phone number"
		} else if email != "" && !utils.IsValidEmail(email) {
			errorMsg = "Invalid email format"
		} else if phone != "" && !utils.IsValidPhone(phone) {
			errorMsg = "Phone must be 10-15 digits"
		} else if password != confirmPassword {
			errorMsg = "Passwords do not match"
		} else if email != "" && services.EmailExists(email) {
			errorMsg = "Email already registered"
		} else if phone != "" && services.PhoneExists(phone) {
			errorMsg = "Phone number already registered"
		}

		if errorMsg != "" {
			tmpl, _ := template.ParseFiles(filepath.Join(config.TemplatesDir, "register.html"))
			tmpl.Execute(w, map[string]string{"error": errorMsg})
			return
		}

		// Create user
		_, err := services.CreateUser(username, email, phone, password)
		if err != nil {
			tmpl, _ := template.ParseFiles(filepath.Join(config.TemplatesDir, "register.html"))
			tmpl.Execute(w, map[string]string{"error": "Registration failed"})
			return
		}

		middleware.SetSessionCookie(w, username)
		http.Redirect(w, r, "/list", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles(filepath.Join(config.TemplatesDir, "register.html"))
	if err != nil {
		http.Error(w, "Template error", 500)
		return
	}
	tmpl.Execute(w, nil)
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	middleware.ClearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
