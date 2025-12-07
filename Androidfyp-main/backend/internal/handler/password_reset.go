// internal/handler/password_reset.go
package handler

import (
	"encoding/json"
	"net/http"
	"net/mail"

	"wazzafak_back/internal/service"
)

// Step 1: User submits email to request password reset
type SendPasswordResetRequest struct {
	Email string `json:"email"`
}

// Step 2: User submits email + code + new password to complete reset
type ResetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

// SendPasswordResetCodeHandler - Step 1: Send reset code to email
func SendPasswordResetCodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req SendPasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid input"})
		return
	}

	// Validate email is not empty
	if req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email cannot be empty"})
		return
	}

	// Validate email format
	if _, err := mail.ParseAddress(req.Email); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid email format"})
		return
	}

	// ✅ All validations passed, now send the reset code
	err := service.SendPasswordResetCode(req.Email)
	if err != nil {
		switch err {
		case service.ErrPasswordResetUserNotFound:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to send reset code"})
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password reset code sent to email",
	})
}

// ResetPasswordHandler - Step 2: Reset password and delete code from database
func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid input"})
		return
	}

	// Validate required fields
	if req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email cannot be empty"})
		return
	}

	if req.Code == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Code cannot be empty"})
		return
	}

	if req.NewPassword == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "New password cannot be empty"})
		return
	}

	if len(req.NewPassword) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Password must be at least 8 characters"})
		return
	}

	err := service.ResetPassword(req.Email, req.Code, req.NewPassword)
	if err != nil {
		switch err {
		case service.ErrInvalidResetCode:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or expired code"})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to reset password"})
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password reset successfully",
	})
}
