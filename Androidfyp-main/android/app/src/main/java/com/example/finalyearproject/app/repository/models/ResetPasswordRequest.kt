// app/repository/models/ResetPasswordRequest.kt
package com.example.finalyearproject.app.repository.models

data class ResetPasswordRequest(
    val email: String,
    val code: String,
    val new_password: String
)

data class ResetPasswordResponse(
    val message: String,
    val error: String? = null
)