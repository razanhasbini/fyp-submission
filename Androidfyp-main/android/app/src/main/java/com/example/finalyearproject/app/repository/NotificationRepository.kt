package com.example.finalyearproject.app.notifications.repository

import com.example.finalyearproject.app.repository.models.Notification
import com.example.finalyearproject.app.repository.network.ApiService

class NotificationRepository(private val api: ApiService) {

    suspend fun getNotifications(): List<Notification> {
        val rawNotifications = api.getNotifications()

        return rawNotifications.map { n ->
            val relatedId = when (n.type.uppercase()) {
                "LIKE", "COMMENT" -> n.post_id?.toString()   // map post_id correctly
                else -> null
            }

            Notification(
                id = n.id,
                title = n.from_user_name,
                message = n.message,
                isRead = n.is_read,
                timestamp = n.created_at,
                type = n.type,
                relatedId = relatedId
            )
        }
    }

    suspend fun getUnreadCount() = api.getUnreadCount()

    suspend fun markNotificationRead(notificationID: Long) =
        api.markNotificationRead(notificationID)

    suspend fun markAllNotificationsRead() = api.markAllNotificationsRead()
}
