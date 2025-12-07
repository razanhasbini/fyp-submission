package com.example.finalyearproject.app.repository.models

import com.google.gson.annotations.SerializedName

data class AddCommentRequest(
    @SerializedName("content") val content: String
)