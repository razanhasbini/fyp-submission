package com.example.finalyearproject.app.utils

import android.os.Build
import android.util.Log
import com.example.finalyearproject.app.repository.network.AiRetrofitClient

object NetworkUtils {
    private const val TAG = "NetworkUtils"
    
    // ============================================
    // CLOUD BACKEND CONFIGURATION
    // ============================================
    // Set this to your cloud backend URL after deploying to Render/Railway
    // Example: "https://your-service-name.onrender.com"
    // Leave empty to use local backend (laptop must be on)
    private const val CLOUD_BACKEND_URL = "" // Set your cloud URL here
    
    // ============================================
    
    fun configureAiServiceUrl() {
        // If cloud URL is set, use it for all devices (makes APK portable!)
        if (CLOUD_BACKEND_URL.isNotEmpty()) {
            AiRetrofitClient.setBaseUrl(CLOUD_BACKEND_URL)
            Log.d(TAG, "Using cloud backend: $CLOUD_BACKEND_URL")
            AiRetrofitClient.reset()
            return
        }
        
        // Otherwise, use local backend (requires laptop on same network)
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
            // Update this IP to match your computer's actual IP address
            // Find it with: ipconfig (Windows) or ifconfig (Mac/Linux)
            // Look for the IPv4 address of your active network adapter
            val localIP = "192.168.18.5" // Updated to match actual IP from ipconfig
            AiRetrofitClient.setBaseUrl("http://$localIP:8089")
            Log.d(TAG, "Physical device detected, using local IP: http://$localIP:8089")
        } else {
            AiRetrofitClient.setBaseUrl("http://10.0.2.2:8089")
            Log.d(TAG, "Emulator detected, using local: http://10.0.2.2:8089")
        }
        AiRetrofitClient.reset()
    }
}



