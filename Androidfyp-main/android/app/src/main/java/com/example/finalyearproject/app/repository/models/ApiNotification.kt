package com.example.finalyearproject.app.repository.models

data class ApiNotification(
    val id: Long,
    val type: String,
    val message: String,
    val is_read: Boolean,
    val created_at: String,
    val post_id: Long?,
    val from_user_id: Long,
    val from_user_name: String,
    val from_user_photo: String?,
    val from_user_username: String,
    val user_id: Long
)
