package com.example.finalyearproject.app.repository.models
import com.example.finalyearproject.app.repository.models.UserData
import com.google.gson.annotations.SerializedName
data class CompleteRegistrationResponse(
    @SerializedName("success")
    val success: Boolean,
    @SerializedName("message")
    val message: String?,
    @SerializedName("statusCode")
    val statusCode: Int,
    @SerializedName("token")
    val token: String?,
    @SerializedName("userId")
    val userId: String?,
    @SerializedName("user")
    val user: UserData?
) {
    fun isSuccessful(): Boolean = success && statusCode in 200..299
}