// app/repository/models/SendPasswordResetRequest.kt
package com.example.finalyearproject.app.repository.models

data class SendPasswordResetRequest(
    val email: String
)

data class SendPasswordResetResponse(
    val message: String,
    val error: String? = null
)