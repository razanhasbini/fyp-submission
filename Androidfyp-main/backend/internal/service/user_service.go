package service

import (
	"errors"

	database "wazzafak_back/internal/database"
	"wazzafak_back/internal/model"
	"wazzafak_back/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrUsernameExists = errors.New("username already exists")
	ErrEmailExists    = errors.New("email already exists")
)

func CreateUser(username, name, email, password, jobPosition, jobPositionType string) (*model.User, error) {
	var existingUser model.User

	// Check username
	if err := database.DB.Where("username = ?", username).First(&existingUser).Error; err == nil {
		return nil, ErrUsernameExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Check email
	if err := database.DB.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, ErrEmailExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	user, err := model.NewUser(username, name, email, password, jobPosition, jobPositionType)
	if err != nil {
		return nil, err
	}

	if err := repository.CreateUser_indatabase(database.DB, user); err != nil {
		return nil, err
	}

	return user, nil
}

func UpdateUserPhoto(userID uint64, photoURL string) error {
	return repository.UpdateUserPhoto(database.DB, userID, photoURL)
}

func DeleteUserPhoto(userID uint64, defaultPhotoURL string) error {
	return repository.UpdateUserPhoto(database.DB, userID, defaultPhotoURL)
}

func UpdateUserName(userID uint64, name string) error {
	return repository.UpdateUserName(database.DB, userID, name)
}

func GetUserByID(userID uint64) (*model.User, error) {
	return repository.GetUserByID(database.DB, userID)
}

func GetUserByUsername(username string) (*model.User, error) {
	return repository.GetUserByUsername(database.DB, username)
}

func GetUserPhotoByID(userID uint64) (string, error) {
	user, err := repository.GetUserByID(database.DB, userID)
	if err != nil {
		return "", err
	}
	return user.PhotoURL, nil
}
