package com.example.finalyearproject.app.repository.models

import com.google.gson.annotations.SerializedName

data class UserProfileResponse(
    @SerializedName("id")
    val id: String,

    @SerializedName("username")
    val username: String?,

    @SerializedName("name")
    val name: String,

    @SerializedName("email")
    val email: String?,

    // This is the critical fix - map photo_url from JSON to photoUrl in Kotlin
    @SerializedName("photo_url")
    val photoUrl: String?,

    @SerializedName("is_admin")
    val isAdmin: Boolean?,

    @SerializedName("job_position")
    val jobPosition: String?,

    @SerializedName("job_position_type")
    val jobPositionType: String?,

    @SerializedName("role")
    val role: String?,

    @SerializedName("followers_count")
    val followersCount: Int = 0,

    @SerializedName("following_count")
    val followingCount: Int = 0,

    @SerializedName("created_at")
    val createdAt: String?
)