package com.example.finalyearproject.app.auth.viewmodels

import android.content.Context
import android.util.Patterns
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.finalyearproject.app.repository.AuthRepository
import com.example.finalyearproject.app.repository.models.CompleteRegistrationRequest
import com.example.finalyearproject.app.repository.models.VerifyEmailCodeRequest
import com.example.datastore.saveToken
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import android.util.Log

class SignUpViewModel(private val authRepository: AuthRepository) : ViewModel() {

    companion object {
        private const val TAG = "SignUpViewModel"
    }

    sealed class SignUpResult {
        object Success : SignUpResult()
        data class Error(val errorMessage: String) : SignUpResult()
    }

    private var selectedJobPosition: String = ""
    private var selectedSpecialization: String = ""
    private var selectedSubSpecialization: String = ""
    private var selectedFinalSpecialization: String = ""

    fun onJobPositionSelected(position: String) {
        selectedJobPosition = position
    }

    fun onSpecializationSelected(specialization: String) {
        selectedSpecialization = specialization
    }

    fun onSubSpecializationSelected(subSpecialization: String) {
        selectedSubSpecialization = subSpecialization
    }

    fun onFinalSpecializationSelected(finalSpecialization: String) {
        selectedFinalSpecialization = finalSpecialization
    }

    // Step 1: Send verification code
    fun sendVerificationCode(email: String, callback: (SignUpResult) -> Unit) {
        val validationError = validateEmail(email)
        if (validationError != null) {
            callback(SignUpResult.Error(validationError))
            return
        }

        viewModelScope.launch {
            try {
                val response = authRepository.sendVerificationCode(email)
                Log.d(TAG, "sendVerificationCode API response: $response")

                val result = if (response.isSuccessful() || response.message?.contains("sent", ignoreCase = true) == true) {
                    SignUpResult.Success
                } else {
                    SignUpResult.Error(response.message ?: "Failed to send verification code")
                }

                withContext(Dispatchers.Main) {
                    callback(result)
                }
            } catch (e: Exception) {
                Log.e(TAG, "sendVerificationCode Exception: ${e.message}")
                withContext(Dispatchers.Main) {
                    callback(SignUpResult.Error(e.message ?: "Failed to send verification code"))
                }
            }
        }
    }

    // Step 2: Verify email code
    fun verifyEmailCode(email: String, code: String, callback: (SignUpResult) -> Unit) {
        if (code.isBlank()) {
            callback(SignUpResult.Error("Please enter the verification code"))
            return
        }

        viewModelScope.launch {
            try {
                val request = VerifyEmailCodeRequest(email, code)
                val response = authRepository.verifyEmailCode(request)
                Log.d(TAG, "verifyEmailCode API response: $response")

                // Fixed: allow success even if success=false, based on message
                val result = if (response.success || response.message?.contains("proceed to complete registration", true) == true) {
                    SignUpResult.Success
                } else {
                    SignUpResult.Error(response.message ?: "Invalid verification code")
                }

                withContext(Dispatchers.Main) {
                    callback(result)
                }
            } catch (e: Exception) {
                Log.e(TAG, "verifyEmailCode Exception: ${e.message}")
                withContext(Dispatchers.Main) {
                    callback(SignUpResult.Error(e.message ?: "Verification failed"))
                }
            }
        }
    }

    // Resend code
    fun resendVerificationCode(email: String, callback: (SignUpResult) -> Unit) {
        viewModelScope.launch {
            try {
                val response = authRepository.sendVerificationCode(email)
                Log.d(TAG, "resendVerificationCode API response: $response")

                val result = if (response.isSuccessful() || response.message?.contains("sent", ignoreCase = true) == true) {
                    SignUpResult.Success
                } else {
                    SignUpResult.Error(response.message ?: "Failed to resend code")
                }

                withContext(Dispatchers.Main) {
                    callback(result)
                }
            } catch (e: Exception) {
                Log.e(TAG, "resendVerificationCode Exception: ${e.message}")
                withContext(Dispatchers.Main) {
                    callback(SignUpResult.Error(e.message ?: "Resend failed"))
                }
            }
        }
    }

    // Step 3: Complete registration
    fun completeRegistration(
        fullName: String,
        email: String,
        password: String,
        jobPosition: String,
        specialization: String,
        subSpecialization: String,
        finalSpecialization: String,
        context: Context,
        callback: (SignUpResult) -> Unit
    ) {
        viewModelScope.launch {
            try {
                val username = email.substringBefore("@").lowercase()
                val request = CompleteRegistrationRequest(
                    email = email,
                    username = username,
                    name = fullName,
                    password = password,
                    job_position = jobPosition,
                    job_position_type = finalSpecialization.ifEmpty { specialization }
                )

                val response = authRepository.completeRegistration(request)
                Log.d(TAG, "completeRegistration API response: $response")

                // FIXED: Check both success flag AND message content
                val isSuccessful = response.isSuccessful() ||
                        response.message?.contains("successfully", ignoreCase = true) == true ||
                        response.message?.contains("registered", ignoreCase = true) == true

                if (isSuccessful) {
                    val token = response.token
                    if (!token.isNullOrBlank()) {
                        saveToken(context, token)
                        withContext(Dispatchers.Main) {
                            callback(SignUpResult.Success)
                        }
                    } else {
                        // Even if no token, registration succeeded, navigate to sign in
                        Log.d(TAG, "Registration successful but no token received - user will sign in manually")
                        withContext(Dispatchers.Main) {
                            callback(SignUpResult.Success)
                        }
                    }
                } else {
                    withContext(Dispatchers.Main) {
                        callback(SignUpResult.Error(response.message ?: "Registration failed"))
                    }
                }
            } catch (e: Exception) {
                Log.e(TAG, "completeRegistration Exception: ${e.message}")
                withContext(Dispatchers.Main) {
                    callback(SignUpResult.Error(e.message ?: "Registration error"))
                }
            }
        }
    }

    private fun validateEmail(email: String): String? {
        return when {
            email.isBlank() -> "Email cannot be empty"
            !Patterns.EMAIL_ADDRESS.matcher(email).matches() -> "Invalid email format"
            else -> null
        }
    }
}