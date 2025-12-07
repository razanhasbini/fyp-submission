package com.example.finalyearproject.app.repository.network.interceptors

import android.content.Context
import com.example.datastore.getToken
import kotlinx.coroutines.runBlocking
import okhttp3.Interceptor
import okhttp3.Response

class AuthInterceptor(private val context: Context) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val originalRequest = chain.request()

        // Fetch the token securely from DataStore
        val token = runBlocking { getToken(context) }

        // If token exists, add Authorization header
        val newRequest = if (token.isNotEmpty()) {
            originalRequest.newBuilder()
                .addHeader("Authorization", "Bearer $token")
                .build()
        } else {
            originalRequest
        }

        return chain.proceed(newRequest)
    }
}
