package com.example.finalyearproject.app.models

import com.google.gson.annotations.SerializedName

data class Post(
    val id: String,
    @SerializedName("user_id") val userId: String,
    @SerializedName("user_name") val userName: String? = null,
    @SerializedName("user_photo_url") val userPhotoUrl: String? = null,
    val content: String,
    @SerializedName("photo_url") val photoUrl: String?,
    @SerializedName("created_at") val createdAt: String, // ✅ This is now working
    var likesCount: Int = 0,
    var commentsCount: Int = 0,
    var isLiked: Boolean = false,
    var isFollowing: Boolean = false,
    @SerializedName("is_owner") val isOwner: Boolean = false

)
