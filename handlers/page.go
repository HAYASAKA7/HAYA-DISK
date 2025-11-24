package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/user/Test1/config"
	"github.com/user/Test1/middleware"
	"github.com/user/Test1/models"
	"github.com/user/Test1/services"
	"github.com/user/Test1/utils"
)

// IndexHandler redirects to login or list based on session
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	username := middleware.GetSessionUser(r)
	if username != "" {
		http.Redirect(w, r, "/list", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// SettingsHandler displays user settings
func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	username := middleware.GetSessionUser(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user := services.GetUser(username)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"username":  user.Username,
		"email":     user.Email,
		"phone":     user.Phone,
		"loginType": user.LoginType,
	}

	tmpl, err := template.ParseFiles(filepath.Join(config.TemplatesDir, "list.html"))
	if err != nil {
		http.Error(w, "Template error", 500)
		return
	}
	tmpl.Execute(w, data)
}

// APIUpdateProfileHandler handles profile updates via API
func APIUpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := middleware.GetSessionUser(r)
	if username == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: false, Message: "Unauthorized"})
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: false, Message: "Invalid request"})
		return
	}

	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Phone)

	// Validation
	if email == "" && phone == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: false, Message: "Please provide either email or phone number"})
		return
	}

	if email != "" && !utils.IsValidEmail(email) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: false, Message: "Invalid email format"})
		return
	}

	if phone != "" && !utils.IsValidPhone(phone) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: false, Message: "Phone must be 10-15 digits"})
		return
	}

	user := services.GetUser(username)
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: false, Message: "User not found"})
		return
	}

	// Check if email already exists (and it's not the user's current email)
	if email != "" && email != user.Email && services.EmailExists(email) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: false, Message: "Email already registered"})
		return
	}

	// Check if phone already exists (and it's not the user's current phone)
	if phone != "" && phone != user.Phone && services.PhoneExists(phone) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: false, Message: "Phone number already registered"})
		return
	}

	// Update user profile
	if err := services.UpdateUserProfile(username, email, phone); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: false, Message: "Update failed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: true, Message: "Profile updated successfully"})
}
