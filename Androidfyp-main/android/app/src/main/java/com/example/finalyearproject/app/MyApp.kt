package com.example.finalyearproject

import android.app.Application
import androidx.appcompat.app.AppCompatDelegate

class MyApp : Application() {
    override fun onCreate() {
        super.onCreate()
        // Force light mode across the whole app
        AppCompatDelegate.setDefaultNightMode(AppCompatDelegate.MODE_NIGHT_NO)
    }
}
