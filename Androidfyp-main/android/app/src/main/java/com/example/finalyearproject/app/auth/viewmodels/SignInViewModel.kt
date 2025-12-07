package com.example.finalyearproject.app.auth.viewmodels

import android.content.Context
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.finalyearproject.app.repository.AuthRepository
import com.example.finalyearproject.app.repository.models.SignInRequest
import com.example.finalyearproject.app.repository.models.SignInResponse
import kotlinx.coroutines.launch

sealed class SignInResult {
    data class Success(val response: SignInResponse) : SignInResult()
    data class Error(val errorMessage: String) : SignInResult()
}

class SignInViewModel(private val context: Context) : ViewModel() {

    private val repository = AuthRepository(context)

    fun signIn(email: String, password: String, callback: (SignInResult) -> Unit) {
        viewModelScope.launch {
            try {
                val response = repository.signIn(SignInRequest(email, password))
                callback(SignInResult.Success(response))
            } catch (e: Exception) {
                // If using Retrofit with HttpException
                val message = if (e is retrofit2.HttpException && (e.code() == 400 || e.code() == 401)) {
                    "Email or password is incorrect"
                } else {
                    e.localizedMessage ?: "Unknown error"
                }

                callback(SignInResult.Error(message))
            }
        }

}
}
