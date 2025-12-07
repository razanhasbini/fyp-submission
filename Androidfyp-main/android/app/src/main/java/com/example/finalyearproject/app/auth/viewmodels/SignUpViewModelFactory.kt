package com.example.finalyearproject.app.auth.viewmodels

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import com.example.finalyearproject.app.repository.AuthRepository

class SignUpViewModelFactory(private val authRepository: AuthRepository) : ViewModelProvider.Factory {
    override fun <T : ViewModel> create(modelClass: Class<T>): T {
        return if (modelClass.isAssignableFrom(SignUpViewModel::class.java)) {
            @Suppress("UNCHECKED_CAST")
            SignUpViewModel(authRepository) as T
        } else {
            throw IllegalArgumentException("Unknown ViewModel class")
        }
    }
}