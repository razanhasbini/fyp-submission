// internal/service/password_reset.go
package service

import (
	"errors"
	"time"

	db "wazzafak_back/internal/database"
	"wazzafak_back/internal/repository"
	"wazzafak_back/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidResetCode          = errors.New("invalid or expired reset code")
	ErrPasswordResetUserNotFound = errors.New("user not found")
)

// SendPasswordResetCode generates and sends a reset code
func SendPasswordResetCode(email string) error {
	// Check if user exists
	_, err := repository.GetUserByEmail(db.DB, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPasswordResetUserNotFound
		}
		return err
	}

	// Generate 6-digit code
	code := generateCode()
	expiresAt := time.Now().Add(15 * time.Minute)

	// Save to database
	if err := repository.SavePasswordResetCode(db.DB, email, code, expiresAt); err != nil {
		return err
	}

	// Send email using utils package
	err = utils.SendPasswordResetEmail(email, code)
	if err != nil {
		return err
	}

	return nil
}

// VerifyPasswordResetCode checks if the code is valid
func VerifyPasswordResetCode(email, code string) error {
	_, err := repository.GetPasswordResetCode(db.DB, email, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidResetCode
		}
		return err
	}
	return nil
}

// ResetPassword updates the user's password after verification
func ResetPassword(email, code, newPassword string) error {
	// Verify the code
	if err := VerifyPasswordResetCode(email, code); err != nil {
		return err
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	if err := repository.UpdateUserPassword(db.DB, email, string(hashedPassword)); err != nil {
		return err
	}

	// Delete the reset code
	return repository.DeletePasswordResetCode(db.DB, email)
}
