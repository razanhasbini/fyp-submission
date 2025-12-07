package repository

import (
	"fmt"
	"wazzafak_back/internal/model"

	"gorm.io/gorm"
)

// AddLike adds a like and creates a notification
func AddLike(db *gorm.DB, userID, postID uint64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Create like
		like := model.Like{UserID: userID, PostID: postID}
		if err := tx.Create(&like).Error; err != nil {
			return err
		}

		// Get post owner and liker's name
		var post model.Post
		if err := tx.Select("user_id").Where("id = ?", postID).First(&post).Error; err != nil {
			return err
		}

		// Don't create notification if user likes their own post
		if post.UserID == userID {
			return nil
		}

		// Get liker's name
		var liker model.User
		if err := tx.Select("name").Where("id = ?", userID).First(&liker).Error; err != nil {
			return err
		}

		// Create notification with message
		message := fmt.Sprintf("%s liked your post", liker.Name)
		notification := &model.Notification{
			UserID:     post.UserID, // recipient (post owner)
			FromUserID: userID,      // actor (the one liking)
			Type:       NotificationTypeLike,
			PostID:     &postID,
			Message:    &message,
			IsRead:     false,
		}

		return CreateNotification(tx, notification)
	})
}

// RemoveLike removes a like
func RemoveLike(db *gorm.DB, userID, postID uint64) error {
	return db.Where("user_id = ? AND post_id = ?", userID, postID).
		Delete(&model.Like{}).Error
}

// Struct for "Who Liked" response
type LikeUserInfo struct {
	UserID   uint64 `json:"user_id"`
	UserName string `json:"user_name"`
	PhotoURL string `json:"photo_url"`
}

// Get all users who liked a post
func GetUsersWhoLikedPost(db *gorm.DB, postID uint64) ([]LikeUserInfo, error) {
	var likes []LikeUserInfo
	err := db.Table("likes").
		Select("likes.user_id, users.name as user_name, users.photo_url").
		Joins("JOIN users ON likes.user_id = users.id").
		Where("likes.post_id = ?", postID).
		Find(&likes).Error
	return likes, err
}
