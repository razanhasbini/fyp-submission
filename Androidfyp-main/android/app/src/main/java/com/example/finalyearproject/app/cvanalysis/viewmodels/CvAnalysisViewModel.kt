package com.example.finalyearproject.app.cvanalysis.viewmodels

import android.app.Application
import android.net.Uri
import android.util.Log
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.finalyearproject.app.repository.AiRepository
import com.example.finalyearproject.app.repository.ProfileRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

data class CvAnalysisUiState(
    val loading: Boolean = false,
    val error: String? = null,
    val analysisResult: String? = null,
    val suggestions: String? = null,
    val uploadSuccess: Boolean = false
)

class CvAnalysisViewModel(app: Application) : AndroidViewModel(app) {
    private val aiRepository = AiRepository(app)
    private val profileRepository = ProfileRepository(app)
    
    private val _state = MutableStateFlow(CvAnalysisUiState())
    val state: StateFlow<CvAnalysisUiState> = _state
    
    companion object {
        private const val TAG = "CvAnalysisViewModel"
    }
    
    fun uploadAndAnalyzeCV(uri: Uri) {
        _state.value = _state.value.copy(loading = true, error = null, uploadSuccess = false)
        
        viewModelScope.launch {
            try {
                val profileResult = profileRepository.getMe()
                val userId = profileResult.getOrNull()?.id?.toLongOrNull() ?: run {
                    _state.value = _state.value.copy(
                        loading = false,
                        error = "User not logged in"
                    )
                    return@launch
                }
                
                Log.d(TAG, "Starting CV upload for user: $userId")
                
                val result = aiRepository.uploadCV(uri, userId)
                
                result.fold(
                    onSuccess = { response ->
                        Log.d(TAG, "CV upload successful: ${response.message}, Analysis: ${response.analysis_result}, Suggestions: ${response.ai_suggestion}")
                        _state.value = _state.value.copy(
                            loading = false,
                            uploadSuccess = true,
                            analysisResult = response.analysis_result ?: response.message,
                            suggestions = response.ai_suggestion
                        )
                    },
                    onFailure = { error ->
                        Log.e(TAG, "CV upload failed", error)
                        _state.value = _state.value.copy(
                            loading = false,
                            error = error.message ?: "Failed to upload CV"
                        )
                    }
                )
            } catch (e: Exception) {
                Log.e(TAG, "Exception during CV upload", e)
                _state.value = _state.value.copy(
                    loading = false,
                    error = e.message ?: "Unknown error occurred"
                )
            }
        }
    }
    
    fun resetState() {
        _state.value = CvAnalysisUiState()
    }
}
