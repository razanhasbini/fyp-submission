package com.example.finalyearproject.app.repository.models
// SignInRequest.kt
data class SignInRequest(
    val email: String,
    val password: String
)

// SignInResponse.kt
data class SignInResponse(
    val token: String
)
