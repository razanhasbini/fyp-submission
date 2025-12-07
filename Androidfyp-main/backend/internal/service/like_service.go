package service

import (
	"errors"
	db "wazzafak_back/internal/database"
	"wazzafak_back/internal/repository"
)

var (
	ErrPostAlreadyLiked = errors.New("post already liked")
	ErrLikeNotFound     = errors.New("like not found")
)

// ✅ Add a like
func LikePost(userID, postID uint64) error {
	return repository.AddLike(db.DB, userID, postID)
}

// ✅ Remove a like
func UnlikePost(userID, postID uint64) error {
	return repository.RemoveLike(db.DB, userID, postID)
}

// ✅ Get users who liked a post
func GetUsersWhoLikedPost(postID uint64) ([]repository.LikeUserInfo, error) {
	return repository.GetUsersWhoLikedPost(db.DB, postID)
}
