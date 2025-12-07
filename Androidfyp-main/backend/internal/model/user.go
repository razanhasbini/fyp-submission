package model

import (
	"time"

	db "wazzafak_back/internal/database"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID              uint64    `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"not null" json:"name"`
	Username        string    `gorm:"unique;not null" json:"username"`
	Email           string    `gorm:"unique;not null" json:"email"`
	Password        string    `gorm:"not null" json:"-"` // Hide password in JSON
	PhotoURL        string    `gorm:"not null;default:'https://upload.wikimedia.org/wikipedia/commons/9/99/Sample_User_Icon.png'" json:"photo_url"`
	IsAdmin         bool      `gorm:"default:false" json:"is_admin"`
	JobPosition     string    `gorm:"not null" json:"job_position"`
	JobPositionType string    `gorm:"not null" json:"job_position_type"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// NewUser creates a new User with hashed password and generated ID
func NewUser(username, name, email, password, jobPosition, jobPositionType string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:              db.GenerateID(),
		Username:        username,
		Name:            name,
		Email:           email,
		Password:        string(hashedPassword),
		JobPosition:     jobPosition,
		JobPositionType: jobPositionType,
		PhotoURL:        "https://upload.wikimedia.org/wikipedia/commons/9/99/Sample_User_Icon.png",
		IsAdmin:         false,
	}, nil
}

// CheckPassword checks if the provided password matches the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
