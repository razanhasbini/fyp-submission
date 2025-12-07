package com.example.finalyearproject.app.repository.models

data class Notification(
    val id: Long,
    val title: String,
    val message: String,
    val isRead: Boolean,
    val timestamp: String,
    val type: String, // "FOLLOW", "LIKE", "COMMENT", etc.
    val relatedId: String? = null // userId or postId
)
