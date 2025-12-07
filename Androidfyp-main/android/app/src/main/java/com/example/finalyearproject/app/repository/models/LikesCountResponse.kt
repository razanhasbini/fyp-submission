package com.example.finalyearproject.app.repository.models

import com.google.gson.annotations.SerializedName

data class LikesCountResponse(
    @SerializedName("count") val count: Int
)