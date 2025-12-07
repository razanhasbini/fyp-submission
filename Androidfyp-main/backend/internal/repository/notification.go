package repository

import (
	"wazzafak_back/internal/model"

	"gorm.io/gorm"
)

// NotificationType constants
const (
	NotificationTypeFollow  = "follow"
	NotificationTypeLike    = "like"
	NotificationTypeComment = "comment"
)

// CreateNotification creates a new notification
func CreateNotification(db *gorm.DB, notification *model.Notification) error {
	return db.Create(notification).Error
}

// GetUserNotifications retrieves all notifications for a user
func GetUserNotifications(db *gorm.DB, userID uint64, limit int) ([]model.Notification, error) {
	var notifications []model.Notification
	query := db.Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&notifications).Error
	return notifications, err
}

// GetUserNotificationsWithDetails retrieves notifications with user details
func GetUserNotificationsWithDetails(db *gorm.DB, userID uint64, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := `
		SELECT 
			n.id,
			n.user_id,
			n.from_user_id,
			n.post_id,
			n.is_read,
			n.type,
			n.message,
			n.created_at,
			u.name as from_user_name,
			u.username as from_user_username,
			u.photo_url as from_user_photo
		FROM notifications n
		JOIN users u ON u.id = n.from_user_id
		WHERE n.user_id = ?
		ORDER BY n.created_at DESC
	`

	if limit > 0 {
		query += " LIMIT ?"
		err := db.Raw(query, userID, limit).Scan(&results).Error
		return results, err
	}

	err := db.Raw(query, userID).Scan(&results).Error
	return results, err
}

// MarkNotificationAsRead marks a notification as read
func MarkNotificationAsRead(db *gorm.DB, notificationID uint64) error {
	return db.Model(&model.Notification{}).
		Where("id = ?", notificationID).
		Update("is_read", true).Error
}

// MarkAllNotificationsAsRead marks all notifications for a user as read
func MarkAllNotificationsAsRead(db *gorm.DB, userID uint64) error {
	return db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

// DeleteNotification deletes a notification
func DeleteNotification(db *gorm.DB, notificationID uint64) error {
	return db.Delete(&model.Notification{}, notificationID).Error
}

// GetUnreadNotificationCount gets count of unread notifications
func GetUnreadNotificationCount(db *gorm.DB, userID uint64) (int64, error) {
	var count int64
	err := db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// CheckNotificationExists checks if a similar notification already exists
func CheckNotificationExists(db *gorm.DB, fromUserID, userID uint64, notifType string, postID *uint64) (bool, error) {
	var count int64
	query := db.Model(&model.Notification{}).
		Where("from_user_id = ? AND user_id = ? AND type = ?", fromUserID, userID, notifType)

	if postID != nil {
		query = query.Where("post_id = ?", *postID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}
