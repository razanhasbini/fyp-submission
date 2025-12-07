package com.example.finalyearproject.app.repository.models
import com.google.gson.annotations.SerializedName
data class VerifyEmailCodeResponse(
    @SerializedName("success")
    val success: Boolean,
    @SerializedName("message")
    val message: String?,
    @SerializedName("statusCode")
    val statusCode: Int
) {
    fun isSuccessful(): Boolean = success && statusCode in 200..299
}