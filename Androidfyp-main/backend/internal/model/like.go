package model

import "time"

// ✅ Like table — connects users and posts
type Like struct {
	UserID    uint64    `gorm:"primaryKey"`
	PostID    uint64    `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
