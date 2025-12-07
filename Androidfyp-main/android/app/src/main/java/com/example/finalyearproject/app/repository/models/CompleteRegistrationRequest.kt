package com.example.finalyearproject.app.repository.models

import com.google.gson.annotations.SerializedName

data class CompleteRegistrationRequest(
    @SerializedName("email")
    val email: String,
    @SerializedName("username")
    val username: String,
    @SerializedName("name")
    val name: String,
    @SerializedName("password")
    val password: String,
    @SerializedName("job_position")        // match backend
    val job_position: String,
    @SerializedName("job_position_type")   // match backend
    val job_position_type: String
)
