package com.example.finalyearproject.app.repository.models

import com.google.gson.annotations.SerializedName

data class UpdatePhotoUrlRequest(
    @SerializedName("photo_url")
    val photoUrl: String
)