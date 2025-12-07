package com.example.finalyearproject.app.interview

import android.util.Log

class InterviewWebSocketClient {
    companion object {
        private const val TAG = "InterviewWebSocket"
    }
    
    fun connect() {
        Log.d(TAG, "Connecting to WebSocket")
    }
    
    fun disconnect() {
        Log.d(TAG, "Disconnecting from WebSocket")
    }
}
