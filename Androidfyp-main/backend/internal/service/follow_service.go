package service

import (
	"errors"

	db "wazzafak_back/internal/database"
	"wazzafak_back/internal/model"
	"wazzafak_back/internal/repository"
)

var (
	ErrFollowYourself = errors.New("you cannot follow yourself")
	ErrFollowFailed   = errors.New("failed to follow user")
	ErrUnfollowFailed = errors.New("failed to unfollow user")
)

func FollowUser(followerID, followingID uint64) error {
	if followerID == followingID {
		return ErrFollowYourself
	}
	return repository.FollowUser(db.DB, followerID, followingID)
}

func UnfollowUser(followerID, followingID uint64) error {
	return repository.UnfollowUser(db.DB, followerID, followingID)
}

func GetFollowersByUsername(username string) ([]model.User, error) {
	user, err := repository.GetUserByUsername(db.DB, username)
	if err != nil {
		return nil, err
	}

	return repository.GetFollowers(db.DB, user.ID)
}

func GetFollowingByUsername(username string) ([]model.User, error) {
	user, err := repository.GetUserByUsername(db.DB, username)
	if err != nil {
		return nil, err
	}

	return repository.GetFollowing(db.DB, user.ID)
}

// NEW: Get followers by user ID directly
func GetFollowersByUserID(userID uint64) ([]model.User, error) {
	return repository.GetFollowers(db.DB, userID)
}

// NEW: Get following by user ID directly
func GetFollowingByUserID(userID uint64) ([]model.User, error) {
	return repository.GetFollowing(db.DB, userID)
}
