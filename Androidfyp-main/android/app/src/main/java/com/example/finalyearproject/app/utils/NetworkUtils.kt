package com.example.finalyearproject.app.utils

import android.os.Build
import android.util.Log
import com.example.finalyearproject.app.repository.network.AiRetrofitClient

object NetworkUtils {
    private const val TAG = "NetworkUtils"
    
    fun configureAiServiceUrl() {
        val isPhysicalDevice = Build.FINGERPRINT.contains("generic").not() &&
                Build.FINGERPRINT.contains("unknown").not() &&
                Build.MODEL.contains("google_sdk").not() &&
                Build.MODEL.contains("Emulator").not() &&
                Build.MODEL.contains("Android SDK built for").not() &&
                Build.MANUFACTURER.contains("Genymotion").not() &&
                Build.BRAND.startsWith("generic").not() &&
                Build.DEVICE.startsWith("generic").not() &&
                Build.PRODUCT.contains("google_sdk").not()
        
        if (isPhysicalDevice) {
            AiRetrofitClient.setBaseUrl("http://192.168.18.5:8089")
            Log.d(TAG, "Physical device detected, using IP: http://192.168.18.5:8089")
        } else {
            AiRetrofitClient.setBaseUrl("http://10.0.2.2:8089")
            Log.d(TAG, "Emulator detected, using: http://10.0.2.2:8089")
        }
        AiRetrofitClient.reset()
    }
}



