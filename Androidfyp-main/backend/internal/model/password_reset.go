// internal/model/password_reset.go
package model

import "time"

type PasswordReset struct {
	Email     string    `gorm:"primaryKey"`
	Code      string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (PasswordReset) TableName() string {
	return "password_resets"
}
