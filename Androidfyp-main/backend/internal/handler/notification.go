package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	database "wazzafak_back/internal/database"
	"wazzafak_back/internal/middleware"
	"wazzafak_back/internal/repository"

	"github.com/go-chi/chi/v5"
)

// GetNotificationsHandler retrieves user notifications with details
func GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized: missing user ID", http.StatusUnauthorized)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	notifications, err := repository.GetUserNotificationsWithDetails(database.DB, userID, limit)
	if err != nil {
		http.Error(w, "Failed to get notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

// GetUnreadCountHandler gets unread notification count
func GetUnreadCountHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized: missing user ID", http.StatusUnauthorized)
		return
	}

	count, err := repository.GetUnreadNotificationCount(database.DB, userID)
	if err != nil {
		http.Error(w, "Failed to get unread count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"unread_count": count})
}

// MarkNotificationReadHandler marks a single notification as read
func MarkNotificationReadHandler(w http.ResponseWriter, r *http.Request) {
	notifIDStr := chi.URLParam(r, "notificationID")
	notifID, err := strconv.ParseUint(notifIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	if err := repository.MarkNotificationAsRead(database.DB, notifID); err != nil {
		http.Error(w, "Failed to mark as read", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Marked as read"})
}

// MarkAllNotificationsReadHandler marks all notifications of the user as read
func MarkAllNotificationsReadHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized: missing user ID", http.StatusUnauthorized)
		return
	}

	if err := repository.MarkAllNotificationsAsRead(database.DB, userID); err != nil {
		http.Error(w, "Failed to mark all as read", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "All marked as read"})
}
