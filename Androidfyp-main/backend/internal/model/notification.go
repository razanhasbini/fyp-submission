package model

import "time"

type Notification struct {
	ID         uint64    `json:"id" gorm:"primaryKey"`
	UserID     uint64    `json:"user_id" gorm:"not null"`      // recipient
	FromUserID uint64    `json:"from_user_id" gorm:"not null"` // actor
	PostID     *uint64   `json:"post_id,omitempty"`
	IsRead     bool      `json:"is_read" gorm:"default:false"`
	Type       string    `json:"type" gorm:"type:varchar(50);not null"`
	Message    *string   `json:"message,omitempty"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
