package com.example.finalyearproject.app.repository.network

import android.util.Log
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.util.concurrent.TimeUnit

object AiRetrofitClient {
    private const val TAG = "AiRetrofitClient"
    private const val DEFAULT_BASE_URL = "http://10.0.2.2:8089"
    
    private var baseUrl: String = DEFAULT_BASE_URL
    private var retrofit: Retrofit? = null
    private var _aiService: AiService? = null
    
    fun setBaseUrl(url: String) {
        if (url != baseUrl) {
            Log.d(TAG, "Setting AI service base URL: $url")
            baseUrl = url
            reset()
        }
    }
    
    fun getBaseUrl(): String = baseUrl
    
    fun reset() {
        Log.d(TAG, "Resetting AI Retrofit client")
        retrofit = null
        _aiService = null
    }
    
    val retrofitInstance: Retrofit
        get() {
            if (retrofit == null) {
                val loggingInterceptor = HttpLoggingInterceptor().apply {
                    level = HttpLoggingInterceptor.Level.BODY
                }
                
                val client = OkHttpClient.Builder()
                    .addInterceptor(loggingInterceptor)
                    .connectTimeout(180, TimeUnit.SECONDS) // 3 minutes for connection
                    .readTimeout(300, TimeUnit.SECONDS) // 5 minutes for reading (CV analysis can take time)
                    .writeTimeout(300, TimeUnit.SECONDS) // 5 minutes for writing (large file uploads)
                    .retryOnConnectionFailure(true)
                    .build()
                
                retrofit = Retrofit.Builder()
                    .baseUrl(baseUrl)
                    .client(client)
                    .addConverterFactory(GsonConverterFactory.create())
                    .build()
                
                Log.d(TAG, "Created new Retrofit instance with base URL: $baseUrl")
            }
            return retrofit!!
        }
    
    val aiService: AiService
        get() {
            // Always create from current retrofit instance to ensure latest base URL
            _aiService = retrofitInstance.create(AiService::class.java)
            Log.d(TAG, "Created AiService instance with base URL: $baseUrl")
            return _aiService!!
        }
}
