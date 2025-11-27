package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/HAYASAKA7/HAYA-DISK/config"
	"github.com/HAYASAKA7/HAYA-DISK/middleware"
	"github.com/HAYASAKA7/HAYA-DISK/models"
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
		loginType := r.FormValue("login_type") // "email" or "phone"
		password := r.FormValue("password")

		var user *models.User

		if loginType == "phone" {
			phone := strings.TrimSpace(r.FormValue("phone"))
			phoneRegion := strings.TrimSpace(r.FormValue("phone_region"))
			// Clean phone number
			phone = utils.CleanPhoneNumber(phone)
			user = services.FindUserByPhoneAndRegion(phone, phoneRegion)
		} else {
			// Default to email login
			email := strings.TrimSpace(r.FormValue("email"))
			user = services.FindUserByEmail(email)
		}

		if user == nil || user.Password != password {
			tmpl, _ := template.ParseFiles(filepath.Join(config.TemplatesDir, "login.html"))
			tmpl.Execute(w, map[string]string{"error": "Invalid email/phone or password"})
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

	// Prepare template data with phone regions
	templateData := map[string]interface{}{
		"PhoneRegions": utils.GetAllPhoneRegions(),
	}

	if r.Method == http.MethodPost {
		username := strings.TrimSpace(r.FormValue("username"))
		email := strings.TrimSpace(r.FormValue("email"))
		phone := strings.TrimSpace(r.FormValue("phone"))
		phoneRegion := strings.TrimSpace(r.FormValue("phone_region"))
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		// Clean phone number
		if phone != "" {
			phone = utils.CleanPhoneNumber(phone)
		}

		// Validation
		var errorMsg string
		if username == "" || password == "" {
			errorMsg = "Username and password are required"
		} else if email == "" && phone == "" {
			errorMsg = "Please provide either email or phone number"
		} else if email != "" && !utils.ValidateEmail(email) {
			errorMsg = "Invalid email format. Please enter a valid email address."
		} else if phone != "" && !utils.ValidatePhone(phone, phoneRegion) {
			errorMsg = utils.GetPhoneValidationError(phoneRegion)
		} else if password != confirmPassword {
			errorMsg = "Passwords do not match"
		} else if email != "" && services.EmailExists(email) {
			errorMsg = "Email already registered"
		} else if phone != "" && services.PhoneExists(phone) {
			errorMsg = "Phone number already registered"
		}

		if errorMsg != "" {
			templateData["error"] = errorMsg
			templateData["username"] = username
			templateData["email"] = email
			templateData["phone"] = phone
			templateData["phone_region"] = phoneRegion
			tmpl, _ := template.ParseFiles(filepath.Join(config.TemplatesDir, "register.html"))
			tmpl.Execute(w, templateData)
			return
		}

		// Create user with phone region
		_, err := services.CreateUser(username, email, phone, phoneRegion, password)
		if err != nil {
			templateData["error"] = "Registration failed"
			tmpl, _ := template.ParseFiles(filepath.Join(config.TemplatesDir, "register.html"))
			tmpl.Execute(w, templateData)
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
	tmpl.Execute(w, templateData)
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	middleware.ClearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
