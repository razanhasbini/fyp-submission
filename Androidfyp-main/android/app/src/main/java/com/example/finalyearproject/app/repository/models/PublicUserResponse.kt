package com.example.finalyearproject.app.repository.models

import com.google.gson.annotations.SerializedName

data class PublicUserResponse(
    @SerializedName("user_id") val id: String, // ✅ renamed for consistency
    @SerializedName("email") val email: String?,
    @SerializedName("name") val name: String?,
    @SerializedName("username") val username: String?,
    @SerializedName("photo_url") val photo_url: String?,
    @SerializedName("created_at") val created_at: String?
)
