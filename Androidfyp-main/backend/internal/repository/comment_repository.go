package repository

import (
	"errors"
	"fmt"
	"wazzafak_back/internal/model"

	"gorm.io/gorm"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
)

// CreateComment creates a new comment in the database
// CreateComment creates a new comment and notification
// CreateComment creates a new comment and notification
func CreateComment(db *gorm.DB, comment *model.Comment) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Create comment
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		// Get post owner
		var post model.Post
		if err := tx.Select("user_id").Where("id = ?", comment.PostID).First(&post).Error; err != nil {
			return err
		}

		// Don't create notification if user comments on their own post
		if post.UserID == comment.UserID {
			return nil
		}

		// Get commenter's name
		var commenter model.User
		if err := tx.Select("name").Where("id = ?", comment.UserID).First(&commenter).Error; err != nil {
			return err
		}

		// Create notification with message including the comment content
		// Truncate comment if it's too long
		commentPreview := comment.Content
		if len(commentPreview) > 100 {
			commentPreview = commentPreview[:97] + "..."
		}

		message := fmt.Sprintf("%s commented on your post: \"%s\"", commenter.Name, commentPreview)
		notification := &model.Notification{
			UserID:     post.UserID,    // recipient (post owner)
			FromUserID: comment.UserID, // actor (the one commenting)
			Type:       NotificationTypeComment,
			PostID:     &comment.PostID,
			Message:    &message,
			IsRead:     false,
		}

		return CreateNotification(tx, notification)
	})
}

// GetCommentByID retrieves a comment by its ID
func GetCommentByID(db *gorm.DB, commentID uint64) (*model.Comment, error) {
	var comment model.Comment
	err := db.Where("id = ?", commentID).First(&comment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	return &comment, nil
}

// DeleteComment deletes a comment by its ID
func DeleteComment(db *gorm.DB, commentID uint64) error {
	result := db.Where("id = ?", commentID).Delete(&model.Comment{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCommentNotFound
	}
	return nil
}

// DeleteCommentByPost deletes a comment that belongs to a specific post
// Deprecated: Use GetCommentByID + DeleteComment with service-layer authorization instead
func DeleteCommentByPost(db *gorm.DB, postID, commentID uint64) error {
	result := db.Where("id = ? AND post_id = ?", commentID, postID).Delete(&model.Comment{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("comment not found for this post")
	}
	return nil
}

// GetCommentsByPostID retrieves all comments for a specific post
// GetCommentsByPostID retrieves all comments with user info for a specific post
func GetCommentsByPostID(db *gorm.DB, postID uint64) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := `
		SELECT 
			c.id,
			c.post_id,
			c.user_id,
			u.name AS user_name,
			u.photo_url AS user_photo_url,
			c.content,
			c.created_at
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.post_id = ?
		ORDER BY c.created_at DESC
	`

	err := db.Raw(query, postID).Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// UpdateComment updates a comment in the database
func UpdateComment(db *gorm.DB, comment *model.Comment) error {
	result := db.Model(&model.Comment{}).
		Where("id = ?", comment.ID).
		Updates(map[string]interface{}{
			"content":    comment.Content,
			"updated_at": gorm.Expr("NOW()"),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCommentNotFound
	}
	return nil
}

// PostExists checks if a post exists in the database
func PostExists(db *gorm.DB, postID uint64) (bool, error) {
	var count int64
	err := db.Model(&model.Post{}).Where("id = ?", postID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetCommentsWithUsers retrieves comments with user information for a post
func GetCommentsWithUsers(db *gorm.DB, postID uint64) ([]*model.Comment, error) {
	var comments []*model.Comment
	err := db.Where("post_id = ?", postID).
		Preload("User"). // Assuming you have a User relation in Comment model
		Order("created_at DESC").
		Find(&comments).Error

	if err != nil {
		return nil, err
	}
	return comments, nil
}

// CountCommentsByPost returns the total number of comments for a post
func CountCommentsByPost(db *gorm.DB, postID uint64) (int64, error) {
	var count int64
	err := db.Model(&model.Comment{}).Where("post_id = ?", postID).Count(&count).Error
	return count, err
}

// CountCommentsByUser returns the total number of comments by a user
func CountCommentsByUser(db *gorm.DB, userID uint64) (int64, error) {
	var count int64
	err := db.Model(&model.Comment{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
