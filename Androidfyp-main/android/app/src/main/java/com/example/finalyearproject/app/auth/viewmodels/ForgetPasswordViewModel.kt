// app/auth/viewmodels/ForgetPasswordViewModel.kt
package com.example.finalyearproject.app.auth.viewmodels

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.finalyearproject.app.repository.models.SendPasswordResetRequest
import com.example.finalyearproject.app.repository.network.*
import kotlinx.coroutines.launch

class ForgetPasswordViewModel : ViewModel() {

    private val _sendResetCodeState = MutableLiveData<Result<String>>()
    val sendResetCodeState: LiveData<Result<String>> = _sendResetCodeState

    fun sendPasswordResetCode(email: String) {
        viewModelScope.launch {
            try {
                val response = RetrofitClient.apiService.sendPasswordReset(
                    SendPasswordResetRequest(email)
                )

                if (response.error != null) {
                    _sendResetCodeState.value = Result.failure(Exception(response.error))
                } else {
                    _sendResetCodeState.value = Result.success(response.message)
                }
            } catch (e: Exception) {
                _sendResetCodeState.value = Result.failure(e)
            }
        }
    }
}