package repository

import (
	"errors"
	"wazzafak_back/internal/model"

	"gorm.io/gorm"
)

// =================== Create Post ===================
func CreatePost(db *gorm.DB, userID uint64, photoURL string, content string) error {
	post := model.NewPost_structure(userID, photoURL, content)
	result := db.Create(&post)
	return result.Error
}

// =================== Delete Post ===================
func DeletePost(db *gorm.DB, postID uint64) error {
	result := db.Delete(&model.Post{}, postID)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// =================== Get All Posts ===================
func GetAllPosts(db *gorm.DB) ([]model.Post, error) {
	var posts []model.Post
	result := db.Order("created_at DESC").Find(&posts)
	if result.Error != nil {
		return nil, result.Error
	}
	return posts, nil
}

// =================== Get Feed (following users) ===================
func GetUserFeed(db *gorm.DB, userID uint64) ([]model.Post, error) {
	var posts []model.Post
	result := db.Table("posts").
		Select("posts.*").
		Joins("INNER JOIN follows ON posts.user_id = follows.following_id").
		Where("follows.follower_id = ?", userID).
		Order("posts.created_at DESC").
		Find(&posts)
	if result.Error != nil {
		return nil, result.Error
	}
	return posts, nil
}

// =================== Get Post by ID ===================
func GetPostByID(db *gorm.DB, id uint64) (*model.Post, error) {
	var post model.Post
	result := db.Where("id = ?", id).First(&post)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("post not found")
	}
	return &post, result.Error
}

// =================== Get Posts by User ===================
func GetPostsByUserID(db *gorm.DB, userID uint64) ([]model.Post, error) {
	var posts []model.Post
	result := db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&posts)
	if result.Error != nil {
		return nil, result.Error
	}
	return posts, nil
}

// =================== Count Likes ===================
func GetLikesCount(db *gorm.DB, postID uint64) (int, error) {
	var count int64
	err := db.Table("likes").
		Where("post_id = ?", postID).
		Count(&count).Error
	return int(count), err
}

// =================== Count Comments ===================
func GetCommentsCount(db *gorm.DB, postID uint64) (int, error) {
	var count int64
	err := db.Table("comments").
		Where("post_id = ?", postID).
		Count(&count).Error
	return int(count), err
}

// =================== Check if User Liked Post ===================
func HasUserLiked(db *gorm.DB, userID uint64, postID uint64) (bool, error) {
	var exists bool
	err := db.Raw(`
		SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND post_id = ?)
	`, userID, postID).Scan(&exists).Error
	return exists, err
}

// =================== Check if User Follows Post Owner ===================
func IsFollowing(db *gorm.DB, followerID uint64, followingID uint64) (bool, error) {
	var exists bool
	err := db.Raw(`
		SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = ? AND following_id = ?)
	`, followerID, followingID).Scan(&exists).Error
	return exists, err
}
