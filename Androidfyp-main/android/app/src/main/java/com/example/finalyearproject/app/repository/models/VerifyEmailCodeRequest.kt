package com.example.finalyearproject.app.repository.models
import com.google.gson.annotations.SerializedName
data class VerifyEmailCodeRequest(
    @SerializedName("email")
    val email: String,
    @SerializedName("code")
    val code: String
)
