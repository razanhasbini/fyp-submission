package model

import "time"

type Comment struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	PostID    uint64    `gorm:"not null" json:"post_id"`
	UserID    uint64    `gorm:"not null" json:"user_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Comment) TableName() string {
	return "comments"
}
