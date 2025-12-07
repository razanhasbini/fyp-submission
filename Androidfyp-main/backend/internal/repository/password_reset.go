// internal/repository/password_reset.go
package repository

import (
	"time"
	"wazzafak_back/internal/model"

	"gorm.io/gorm"
)

// SavePasswordResetCode stores or updates a password reset code
func SavePasswordResetCode(db *gorm.DB, email, code string, expiresAt time.Time) error {
	reset := &model.PasswordReset{
		Email:     email,
		Code:      code,
		ExpiresAt: expiresAt,
	}

	// Use Save to insert or update
	return db.Save(reset).Error
}

// GetPasswordResetCode retrieves a password reset entry
func GetPasswordResetCode(db *gorm.DB, email, code string) (*model.PasswordReset, error) {
	var reset model.PasswordReset
	err := db.Where("email = ? AND code = ? AND expires_at > ?", email, code, time.Now()).
		First(&reset).Error
	if err != nil {
		return nil, err
	}
	return &reset, nil
}

// DeletePasswordResetCode removes the reset code after use
func DeletePasswordResetCode(db *gorm.DB, email string) error {
	return db.Where("email = ?", email).Delete(&model.PasswordReset{}).Error
}

// UpdateUserPassword updates a user's password
func UpdateUserPassword(db *gorm.DB, email, hashedPassword string) error {
	result := db.Model(&model.User{}).
		Where("email = ?", email).
		Update("password", hashedPassword)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
