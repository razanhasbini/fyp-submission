// app/auth/viewmodels/ChangePassViewModel.kt
package com.example.finalyearproject.app.auth.viewmodels

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.finalyearproject.app.repository.models.ResetPasswordRequest
import com.example.finalyearproject.app.repository.network.*
import kotlinx.coroutines.launch

class ChangePassViewModel : ViewModel() {

    private val _resetPasswordState = MutableLiveData<Result<String>>()
    val resetPasswordState: LiveData<Result<String>> = _resetPasswordState

    fun resetPassword(email: String, code: String, newPassword: String) {
        viewModelScope.launch {
            try {
                val response = RetrofitClient.apiService.resetPassword(
                    ResetPasswordRequest(email, code, newPassword)
                )

                if (response.error != null) {
                    _resetPasswordState.value = Result.failure(Exception(response.error))
                } else {
                    _resetPasswordState.value = Result.success(response.message)
                }
            } catch (e: Exception) {
                _resetPasswordState.value = Result.failure(e)
            }
        }
    }
}