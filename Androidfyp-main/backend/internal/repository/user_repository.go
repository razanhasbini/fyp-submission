package repository

import (
	"errors"
	"wazzafak_back/internal/model"

	"gorm.io/gorm"
)

// CreateUser inserts a new user into the database
func CreateUser_indatabase(db *gorm.DB, user *model.User) error {
	result := db.Create(&user)
	return result.Error
}

// GetUsers retrieves all users from the database
func GetUsers(db *gorm.DB) ([]model.User, error) {
	var users []model.User
	result := db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// UpdateUserPhoto updates a user's photo URL
func UpdateUserPhoto(db *gorm.DB, id uint64, photoURL string) error {
	result := db.Model(&model.User{}).Where("id = ?", id).Update("photo_url", photoURL)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// UpdateUserName updates a user's name
func UpdateUserName(db *gorm.DB, id uint64, name string) error {
	result := db.Model(&model.User{}).Where("id = ?", id).Update("name", name)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// GetUserByEmail finds a user by their email
func GetUserByEmail(db *gorm.DB, email string) (*model.User, error) {
	var user model.User
	result := db.Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	return &user, result.Error
}

// GetUserByID finds a user by their ID
func GetUserByID(db *gorm.DB, id uint64) (*model.User, error) {
	var user model.User
	result := db.Where("id = ?", id).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	return &user, result.Error
}

// GetUserByUsername finds a user by their username
func GetUserByUsername(db *gorm.DB, username string) (*model.User, error) {
	var user model.User
	result := db.Where("username = ?", username).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	return &user, result.Error
}
