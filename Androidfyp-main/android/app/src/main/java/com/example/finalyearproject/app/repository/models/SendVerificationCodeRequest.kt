package com.example.finalyearproject.app.repository.models
import com.google.gson.annotations.SerializedName
data class SendVerificationCodeRequest(
    @SerializedName("email")
    val email: String
)