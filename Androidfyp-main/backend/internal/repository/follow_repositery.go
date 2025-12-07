package repository

import (
	"errors"
	"fmt"
	"wazzafak_back/internal/model"

	"gorm.io/gorm"
)

func FollowUser(db *gorm.DB, followerID, followingID uint64) error {
	if followerID == followingID {
		return errors.New("you can't follow yourself")
	}

	// Start transaction
	return db.Transaction(func(tx *gorm.DB) error {
		follow := model.Follow{
			FollowerID:  followerID,
			FollowingID: followingID,
		}

		if err := tx.Create(&follow).Error; err != nil {
			return err
		}

		// Get follower's name
		var follower model.User
		if err := tx.Select("name").Where("id = ?", followerID).First(&follower).Error; err != nil {
			return err
		}

		// Create notification with message
		message := fmt.Sprintf("%s started following you", follower.Name)
		notification := &model.Notification{
			UserID:     followingID, // recipient (the one being followed)
			FromUserID: followerID,  // actor (the one following)
			Type:       NotificationTypeFollow,
			Message:    &message,
			IsRead:     false,
		}

		return CreateNotification(tx, notification)
	})
}

func UnfollowUser(db *gorm.DB, followerID, followingID uint64) error {
	return db.Delete(&model.Follow{}, "follower_id = ? AND following_id = ?", followerID, followingID).Error
}

// GetFollowers retrieves all followers of a user
func GetFollowers(db *gorm.DB, userID uint64) ([]model.User, error) {
	var users []model.User
	result := db.Table("users").
		Joins("JOIN follows ON users.id = follows.follower_id").
		Where("follows.following_id = ?", userID).
		Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// GetFollowing retrieves all users that a user is following
func GetFollowing(db *gorm.DB, userID uint64) ([]model.User, error) {
	var users []model.User
	result := db.Table("users").
		Joins("JOIN follows ON users.id = follows.following_id").
		Where("follows.follower_id = ?", userID).
		Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}
