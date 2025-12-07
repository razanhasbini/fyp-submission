package com.example.finalyearproject.app.repository.network

import android.util.Log
import okhttp3.Interceptor
import okhttp3.Response
import okhttp3.ResponseBody.Companion.toResponseBody

/**
 * Interceptor to log all HTTP requests and responses
 * This helps debug API issues
 */
class DebugLoggingInterceptor : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()

        // Log request
        Log.d("API_REQUEST", "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        Log.d("API_REQUEST", "URL: ${request.url}")
        Log.d("API_REQUEST", "Method: ${request.method}")
        Log.d("API_REQUEST", "Headers: ${request.headers}")

        val startTime = System.currentTimeMillis()

        val response = try {
            chain.proceed(request)
        } catch (e: Exception) {
            Log.e("API_REQUEST", "Request failed: ${e.message}", e)
            throw e
        }

        val duration = System.currentTimeMillis() - startTime

        // Log response
        val responseBody = response.body
        val bodyString = responseBody?.string() ?: ""

        Log.d("API_RESPONSE", "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        Log.d("API_RESPONSE", "URL: ${response.request.url}")
        Log.d("API_RESPONSE", "Status: ${response.code}")
        Log.d("API_RESPONSE", "Duration: ${duration}ms")
        Log.d("API_RESPONSE", "Body: $bodyString")
        Log.d("API_RESPONSE", "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

        // Recreate response with the body we read
        return response.newBuilder()
            .body(bodyString.toResponseBody(responseBody?.contentType()))
            .build()
    }
}

/**
 * Add this to your NetworkModule.kt when creating OkHttpClient
 *
 * Example:
 *
 * @Provides
 * @Singleton
 * fun provideOkHttpClient(
 *     authInterceptor: AuthInterceptor
 * ): OkHttpClient {
 *     return OkHttpClient.Builder()
 *         .addInterceptor(authInterceptor)
 *         .addInterceptor(DebugLoggingInterceptor()) // Add this line
 *         .connectTimeout(30, TimeUnit.SECONDS)
 *         .readTimeout(30, TimeUnit.SECONDS)
 *         .writeTimeout(30, TimeUnit.SECONDS)
 *         .build()
 * }
 */