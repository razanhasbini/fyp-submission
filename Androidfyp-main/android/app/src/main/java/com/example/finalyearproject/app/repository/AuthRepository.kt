package com.example.finalyearproject.app.repository

import android.content.Context
import android.util.Log
import com.example.datastore.saveToken
import com.example.finalyearproject.app.repository.models.*
import com.example.finalyearproject.app.repository.network.*

class AuthRepository(private val context: Context? = null) {

    private val api = RetrofitClient.apiService

    // Existing methods
    suspend fun sendVerificationCode(email: String): SendVerificationCodeResponse {
        val request = SendVerificationCodeRequest(email)
        return api.sendVerificationCode(request)
    }

    suspend fun verifyEmailCode(request: VerifyEmailCodeRequest): VerifyEmailCodeResponse {
        return api.verifyEmailCode(request)
    }

    suspend fun resendVerificationCode(email: String): SendVerificationCodeResponse {
        val request = SendVerificationCodeRequest(email)
        return api.sendVerificationCode(request)
    }

    suspend fun completeRegistration(request: CompleteRegistrationRequest): CompleteRegistrationResponse {
        return try {
            Log.d("API_DEBUG", "Sending complete registration request: $request")
            val response = api.completeRegistration(request)
            Log.d("API_DEBUG", "Response received: $response")
            response
        } catch (e: Exception) {
            Log.e("API_DEBUG", "Error during complete registration", e)
            throw e
        }
    }

    // ---------- NEW SIGN-IN METHOD ----------
    suspend fun signIn(request: SignInRequest): SignInResponse {
        return try {
            Log.d("API_DEBUG", "Sending sign-in request: $request")
            val response = api.signIn(request)
            Log.d("API_DEBUG", "Sign-in response: $response")

            // Save token encrypted in DataStore if context is provided
            context?.let {
                saveToken(it, response.token)
                Log.d("API_DEBUG", "Token saved encrypted in DataStore")
            }

            response
        } catch (e: Exception) {
            Log.e("API_DEBUG", "Error during sign-in", e)
            throw e
        }
    }
}
