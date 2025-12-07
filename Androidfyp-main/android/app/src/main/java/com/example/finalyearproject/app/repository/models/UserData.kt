package com.example.finalyearproject.app.repository.models

import com.google.gson.annotations.SerializedName

data class UserData(
    @SerializedName("id")
    val id: String?,
    @SerializedName("email")
    val email: String?,
    @SerializedName("username")
    val username: String?,
    @SerializedName("name")
    val name: String?
)