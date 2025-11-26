package handlers

import (
	"encoding/json"
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

	newUsername := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)
	phone := strings.TrimSpace(req.Phone)

	// Validation - at least one field must be provided
	if newUsername == "" && email == "" && phone == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: false, Message: "Please provide at least one field to update"})
		return
	}

	// Validate new username if provided
	if newUsername != "" && newUsername != username {
		if len(newUsername) < 3 || len(newUsername) > 20 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: false, Message: "Username must be 3-20 characters"})
			return
		}
		if services.UsernameExists(newUsername) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: false, Message: "Username already taken"})
			return
		}
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

	// Update username first if changed
	updatedUsername := username
	if newUsername != "" && newUsername != username {
		if err := services.UpdateUsername(username, newUsername); err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: false, Message: "Failed to update username: " + err.Error()})
			return
		}
		updatedUsername = newUsername

		// Update session with new username
		middleware.UpdateSession(r, newUsername)
	}

	// Update user profile (email and phone)
	if email != "" || phone != "" {
		if err := services.UpdateUserProfile(updatedUsername, email, phone); err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: false, Message: "Update failed"})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.UpdateProfileResponse{Success: true, Message: "Profile updated successfully"})
}

// APIGetUserInfoHandler returns the current user's information
func APIGetUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := middleware.GetSessionUser(r)
	if username == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Unauthorized"})
		return
	}

	user := services.GetUser(username)
	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "User not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.UserInfoResponse{
		Username: user.Username,
		Email:    user.Email,
		Phone:    user.Phone,
	})
}
