package service

import (
	"errors"
	"fmt"
	"time"

	db "wazzafak_back/internal/database"
	"wazzafak_back/internal/model"
	"wazzafak_back/internal/repository"
)

var (
	ErrEmptyComment    = errors.New("comment content cannot be empty")
	ErrCommentNotFound = errors.New("comment not found")
)

// AddComment creates a new comment on a post
func AddComment(userID, postID uint64, content string) (*model.Comment, error) {
	if content == "" {
		return nil, ErrEmptyComment
	}

	// Check if post exists
	postExists, err := repository.PostExists(db.DB, postID)
	if err != nil {
		return nil, err
	}
	if !postExists {
		return nil, ErrPostNotFound
	}

	comment := &model.Comment{
		PostID:  postID,
		UserID:  userID,
		Content: content,
	}

	err = repository.CreateComment(db.DB, comment)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// DeleteCommentFromPost deletes a comment only if the current user owns it
func DeleteCommentFromPost(postID, commentID, userID uint64) error {
	comment, err := repository.GetCommentByID(db.DB, commentID)
	if err != nil {
		if err == repository.ErrCommentNotFound {
			return ErrCommentNotFound
		}
		return err
	}

	if comment.PostID != postID {
		return ErrCommentNotFound
	}

	if comment.UserID != userID {
		return ErrUnauthorized
	}

	return repository.DeleteComment(db.DB, commentID)
}

// ✅ Struct returned to frontend
type CommentResponse struct {
	ID           uint64 `json:"id"`
	PostID       uint64 `json:"post_id"`
	UserID       uint64 `json:"user_id"`
	UserName     string `json:"user_name"`
	UserPhotoURL string `json:"user_photo_url"`
	Content      string `json:"content"`
	CreatedAt    string `json:"created_at"`
	IsOwner      bool   `json:"is_owner"`
}

// ✅ Fixed GetCommentsByPost — handles time.Time safely
func GetCommentsByPost(postID, currentUserID uint64) ([]CommentResponse, error) {
	rawComments, err := repository.GetCommentsByPostID(db.DB, postID)
	if err != nil {
		return nil, err
	}

	var comments []CommentResponse
	for _, c := range rawComments {
		// Convert IDs safely
		var id, postId, userId uint64
		switch v := c["id"].(type) {
		case int64:
			id = uint64(v)
		case uint64:
			id = v
		}

		switch v := c["post_id"].(type) {
		case int64:
			postId = uint64(v)
		case uint64:
			postId = v
		}

		switch v := c["user_id"].(type) {
		case int64:
			userId = uint64(v)
		case uint64:
			userId = v
		}

		// Strings / text
		userName := fmt.Sprint(c["user_name"])
		userPhoto := fmt.Sprint(c["user_photo_url"])
		content := fmt.Sprint(c["content"])

		// Handle created_at safely
		var createdAt string
		switch v := c["created_at"].(type) {
		case time.Time:
			createdAt = v.UTC().Format(time.RFC3339)
		case string:
			createdAt = v
		case []byte:
			createdAt = string(v)
		default:
			createdAt = ""
		}

		comments = append(comments, CommentResponse{
			ID:           id,
			PostID:       postId,
			UserID:       userId,
			UserName:     userName,
			UserPhotoURL: userPhoto,
			Content:      content,
			CreatedAt:    createdAt,
			IsOwner:      userId == currentUserID,
		})
	}

	return comments, nil
}

// GetCommentByID retrieves a comment by its ID
func GetCommentByID(commentID uint64) (*model.Comment, error) {
	comment, err := repository.GetCommentByID(db.DB, commentID)
	if err != nil {
		if err == repository.ErrCommentNotFound {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}
	return comment, nil
}

// UpdateComment updates a comment content with ownership check
func UpdateComment(commentID, userID uint64, newContent string) (*model.Comment, error) {
	if newContent == "" {
		return nil, ErrEmptyComment
	}

	comment, err := repository.GetCommentByID(db.DB, commentID)
	if err != nil {
		if err == repository.ErrCommentNotFound {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}

	if comment.UserID != userID {
		return nil, ErrUnauthorized
	}

	comment.Content = newContent
	err = repository.UpdateComment(db.DB, comment)
	if err != nil {
		return nil, err
	}

	return comment, nil
}
