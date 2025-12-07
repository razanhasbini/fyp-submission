package com.example.finalyearproject.app.repository.models

data class Comment(
    val id: String,
    val postId: String,
    val userId: String,
    val userName: String,
    val userPhotoUrl: String?,
    val content: String,
    val createdAt: String,
    val isOwner: Boolean // ✅ add this

)
