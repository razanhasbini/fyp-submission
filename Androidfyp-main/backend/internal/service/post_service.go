package service

import (
	"errors"

	db "wazzafak_back/internal/database"
	"wazzafak_back/internal/model"
	"wazzafak_back/internal/repository"

	"gorm.io/gorm"
)

// --- Error definitions ---
var (
	ErrPostNotFound     = errors.New("post not found")
	ErrInvalidPostInput = errors.New("either content or photo_url must be provided")
	ErrUnauthorized     = errors.New("unauthorized")
)

// =================== Create a new post ===================
func CreatePost(userID uint64, photoURL, content string) error {
	if content == "" && photoURL == "" {
		return ErrInvalidPostInput
	}
	return repository.CreatePost(db.DB, userID, photoURL, content)
}

// =================== Delete a post (only owner) ===================
func DeletePost(postID, userID uint64) error {
	post, err := repository.GetPostByID(db.DB, postID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrPostNotFound
	}
	if err != nil {
		return err
	}

	if post.UserID != userID {
		return ErrUnauthorized
	}

	return repository.DeletePost(db.DB, postID)
}

// =================== Get all posts ===================
func GetAllPosts() ([]model.Post, error) {
	return repository.GetAllPosts(db.DB)
}

// =================== Get feed (following users' posts) ===================
func GetUserFeed(userID uint64) ([]model.Post, error) {
	return repository.GetUserFeed(db.DB, userID)
}

// =================== Get a post by ID ===================
func GetPostByID(postID uint64) (*model.Post, error) {
	post, err := repository.GetPostByID(db.DB, postID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPostNotFound
	}
	return post, err
}

// =================== Get posts by username ===================
func GetPostsByUsername(username string) ([]model.Post, error) {
	user, err := repository.GetUserByUsername(db.DB, username)
	if err != nil {
		return nil, err
	}
	return repository.GetPostsByUserID(db.DB, user.ID)
}

// =================== Get likes count for a post ===================
func GetLikesCount(postID uint64) (int, error) {
	return repository.GetLikesCount(db.DB, postID)
}

// =================== Get comments count for a post ===================
func GetCommentsCount(postID uint64) (int, error) {
	return repository.GetCommentsCount(db.DB, postID)
}

// =================== Check if user liked a post ===================
func HasUserLiked(userID uint64, postID uint64) (bool, error) {
	return repository.HasUserLiked(db.DB, userID, postID)
}

// =================== Check if user follows post owner ===================
func IsFollowing(followerID uint64, followingID uint64) (bool, error) {
	return repository.IsFollowing(db.DB, followerID, followingID)
}

// =================== Get posts by userID (from JWT) ===================
func GetMyPosts(userID uint64) ([]model.Post, error) {
	return repository.GetPostsByUserID(db.DB, userID)
}
