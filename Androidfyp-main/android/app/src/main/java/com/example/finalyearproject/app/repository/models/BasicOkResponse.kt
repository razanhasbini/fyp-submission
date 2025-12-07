package com.example.finalyearproject.app.repository.models

import com.google.gson.annotations.SerializedName

data class BasicOkResponse(
    @SerializedName("message") val message: String
)