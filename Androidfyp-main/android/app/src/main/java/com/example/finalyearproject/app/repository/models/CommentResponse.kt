package com.example.finalyearproject.app.repository.models

import com.google.gson.annotations.SerializedName

data class CommentResponse(
    @SerializedName("id") val id: String,
    @SerializedName("post_id") val post_id: String,
    @SerializedName("user_id") val user_id: String,
    @SerializedName("content") val content: String,
    @SerializedName("created_at") val created_at: String?,
    @SerializedName("is_owner") val is_owner: Boolean? = false  // ✅ Added is_owner field
)