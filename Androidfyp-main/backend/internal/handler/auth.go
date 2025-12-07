package handler

import (
	"encoding/json"
	"net/http"
	"net/mail"

	"wazzafak_back/internal/service"
)

// Step 1: User submits email only
type SendVerificationCodeRequest struct {
	Email string `json:"email"`
}

// Step 2: User submits email + code
type VerifyEmailCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// Step 3: User submits complete registration info (after email is verified)
type CompleteRegistrationRequest struct {
	Email           string `json:"email"`
	Username        string `json:"username"`
	Name            string `json:"name"`
	Password        string `json:"password"`
	JobPosition     string `json:"job_position"`
	JobPositionType string `json:"job_position_type"`
}

// SendVerificationCodeHandler - Step 1: Send code to email (NO password needed)
func SendVerificationCodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req SendVerificationCodeRequest
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

	// ✅ All validations passed, now send the code
	_, err := service.SendVerificationCode(req.Email)
	if err != nil {
		switch err.Error() {
		case "email already registered":
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "Email already exists"})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Verification code sent to email",
	})
}

// VerifyEmailCodeHandler - Step 2: Verify code (NO password, just confirm email)
func VerifyEmailCodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req VerifyEmailCodeRequest
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

	err := service.VerifyEmailCode(req.Email, req.Code)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Email verified successfully. Proceed to complete registration.",
	})
}

// CompleteRegistrationHandler - Step 3: Create user with password (after email verified)
func CompleteRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req CompleteRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid input"})
		return
	}

	// Validate required fields
	if req.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Username cannot be empty"})
		return
	}

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Name cannot be empty"})
		return
	}

	if req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email cannot be empty"})
		return
	}

	if req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Password cannot be empty"})
		return
	}

	if req.JobPosition == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Job position cannot be empty"})
		return
	}

	if req.JobPositionType == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Job position type cannot be empty"})
		return
	}

	userID, err := service.CompleteUserRegistration(
		req.Email,
		req.Username,
		req.Name,
		req.Password,
		req.JobPosition,
		req.JobPositionType,
	)
	if err != nil {
		switch err.Error() {
		case "email not verified. Please verify your email first":
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		case "email already registered":
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "Email already exists"})
		default:
			if err.Error() == "registration failed: Error 1062: Duplicate entry" ||
				err.Error() == "registration failed: UNIQUE constraint failed" {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]string{"error": "Username already exists"})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create user"})
			}
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user_id": userID,
	})
}
