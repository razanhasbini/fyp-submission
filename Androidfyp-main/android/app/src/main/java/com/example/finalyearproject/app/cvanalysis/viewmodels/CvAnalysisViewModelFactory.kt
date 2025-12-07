package com.example.finalyearproject.app.cvanalysis.viewmodels

import android.app.Application
import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider

class CvAnalysisViewModelFactory(private val application: Application) : ViewModelProvider.Factory {
    @Suppress("UNCHECKED_CAST")
    override fun <T : ViewModel> create(modelClass: Class<T>): T {
        if (modelClass.isAssignableFrom(CvAnalysisViewModel::class.java)) {
            return CvAnalysisViewModel(application) as T
        }
        throw IllegalArgumentException("Unknown ViewModel class")
    }
}



