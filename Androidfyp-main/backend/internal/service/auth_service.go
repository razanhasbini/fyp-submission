package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	db "wazzafak_back/internal/database"
	"wazzafak_back/internal/model"
	"wazzafak_back/utils"

	"gorm.io/gorm"
)

// Generate a 6-digit numeric verification code
var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func generateCode() string {
	return fmt.Sprintf("%06d", rnd.Intn(1000000))
}

// SendVerificationCode - Step 1: Generate code and send email (no password involved)
func SendVerificationCode(email string) (string, error) {
	// Check if email already exists in users table
	var count int64
	if err := db.DB.Table("users").Where("email = ?", email).Count(&count).Error; err != nil {
		return "", err
	}
	if count > 0 {
		return "", errors.New("email already registered")
	}

	code := generateCode()
	expiresAt := time.Now().Add(15 * time.Minute)

	// Store the verification code
	err := db.DB.Exec(`
		INSERT INTO email_verifications (email, verification_code, created_at, expires_at, verified)
		VALUES (?, ?, NOW(), ?, false)
	`, email, code, expiresAt).Error
	if err != nil {
		return "", err
	}

	// Send verification email
	err = utils.SendVerificationEmail(email, code)
	if err != nil {
		return "", err
	}

	return code, nil
}

// VerifyEmailCode - Step 2: Verify code and mark email as verified (still no user created)
func VerifyEmailCode(email, code string) error {
	var id int
	var expiresAt time.Time

	row := db.DB.Raw(`
		SELECT id, expires_at FROM email_verifications
		WHERE email = ? AND verification_code = ? AND verified = false
	`, email, code).Row()

	if err := row.Scan(&id, &expiresAt); err != nil {
		return errors.New("invalid verification code or email")
	}

	if time.Now().After(expiresAt) {
		return errors.New("verification code expired")
	}

	// Mark verification as used
	err := db.DB.Exec(`UPDATE email_verifications SET verified = true WHERE id = ?`, id).Error
	if err != nil {
		return err
	}

	return nil
}

// CompleteUserRegistration - Step 3: Create user with password (after email verified)
func CompleteUserRegistration(email, username, name, password, jobPosition, jobPositionType string) (uint64, error) {
	var userID uint64

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// Lock verification row to avoid race conditions
		var verificationID int
		row := tx.Raw(`
			SELECT id FROM email_verifications
			WHERE email = ? AND verified = true
			LIMIT 1
			FOR UPDATE
		`, email).Row()

		if err := row.Scan(&verificationID); err != nil {
			return errors.New("email not verified. Please verify your email first")
		}

		// Create new user
		user, err := model.NewUser(username, name, email, password, jobPosition, jobPositionType)
		if err != nil {
			return errors.New("failed to create user: " + err.Error())
		}

		// Save user in DB
		if err := tx.Create(user).Error; err != nil {
			return errors.New("registration failed: " + err.Error())
		}

		// Delete verification record to prevent reuse
		if err := tx.Exec(`DELETE FROM email_verifications WHERE id = ?`, verificationID).Error; err != nil {
			return errors.New("failed to delete verification record: " + err.Error())
		}

		userID = user.ID
		return nil
	})

	if err != nil {
		return 0, err
	}

	return userID, nil
}
