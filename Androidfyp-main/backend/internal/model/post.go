package model

import (
	"time"
	db "wazzafak_back/internal/database"
)

type Post struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    uint64    `gorm:"not null;index" json:"user_id"`
	PhotoURL  string    `gorm:"size:500;default:''" json:"photo_url"`
	Content   string    `gorm:"type:text;default:''" json:"content"`
	CreatedAt time.Time `json:"created_at"` // ✅ Add this
	UpdatedAt time.Time `json:"updated_at"` // (optional, but useful)
}

func NewPost_structure(userID uint64, photoURL, content string) Post {
	return Post{
		ID:       db.GenerateID(),
		UserID:   userID,
		PhotoURL: photoURL,
		Content:  content,
	}
}
