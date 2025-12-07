package model

import "time"

type Follow struct {
	FollowerID  uint64    `gorm:"primaryKey"`
	FollowingID uint64    `gorm:"primaryKey"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}
