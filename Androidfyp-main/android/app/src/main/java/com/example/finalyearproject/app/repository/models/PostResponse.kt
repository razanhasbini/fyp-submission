package com.example.finalyearproject.app.repository.models

import com.google.gson.annotations.SerializedName

data class PostResponse(
    @SerializedName("id") val id: String,
    @SerializedName("user_id") val user_id: String,
    @SerializedName("photo_url") val photo_url: String?,
    @SerializedName("content") val content: String,
    @SerializedName("created_at") val created_at: String?,

    // ✅ Optional fields — safely ignored if backend doesn't send them
    @SerializedName("likes_count") val likes_count: Int? = 0,
    @SerializedName("comments_count") val comments_count: Int? = 0,
    @SerializedName("is_liked") val is_liked: Boolean? = false,
    @SerializedName("is_following") val is_following: Boolean? = false,
    @SerializedName("is_owner") val is_owner: Boolean? = false  // ✅ Added is_owner field
)