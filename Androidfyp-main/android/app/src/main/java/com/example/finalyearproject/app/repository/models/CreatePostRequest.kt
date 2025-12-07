package com.example.finalyearproject.app.repository.models

import com.google.gson.annotations.SerializedName

data class CreatePostRequest(
    @SerializedName("content") val content: String,
    @SerializedName("photo_url") val photo_url: String?
)