package service

import (
	"errors"
	"time"
	db "wazzafak_back/internal/database"
	"wazzafak_back/internal/repository"

	"strconv"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your-256-bit-secret") // TODO: load from env/config

// GenerateJWT creates a JWT token for a user ID
func GenerateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(72 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// Login authenticates the user and returns a JWT token
func Login(email, password string) (string, error) {
	user, err := repository.GetUserByEmail(db.DB, email)
	if err != nil {
		return "", errors.New("invalid email")
	}

	if !checkPassword(password, user.Password) {
		return "", errors.New("invalid password")
	}

	token, err := GenerateJWT(strconv.FormatUint(user.ID, 10))
	if err != nil {
		return "", err
	}

	return token, nil
}

func checkPassword(inputPassword, storedHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(inputPassword))
	return err == nil
}
