package com.example.finalyearproject.app.repository.models

import com.google.gson.annotations.SerializedName

data class UserSummary(
    val id: Long,
    val name: String?,
    val username: String?,
    val email: String?,
    @SerializedName("photo_url") val photoUrl: String?,
    @SerializedName("job_position") val job_position: String?,
    @SerializedName("job_position_type") val job_position_type: String?,
    @SerializedName("is_admin") val is_admin: Boolean,
    @SerializedName("created_at") val created_at: String?,
    val role: String?,

    @SerializedName("is_following")
    var isFollowing: Boolean? = null   // must be mutable so adapter can toggle
)
