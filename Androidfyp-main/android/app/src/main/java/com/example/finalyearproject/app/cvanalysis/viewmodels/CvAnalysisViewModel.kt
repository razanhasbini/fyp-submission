package com.example.finalyearproject.app.cvanalysis.viewmodels

import android.app.Application
import android.net.Uri
import android.util.Log
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.finalyearproject.app.repository.AiRepository
import com.example.finalyearproject.app.repository.ProfileRepository
import kotlinx.coroutines.delay
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
                        Log.d(TAG, "CV upload successful: ${response.message}")
                        Log.d(TAG, "Analysis result: ${response.analysis_result?.take(100)}...")
                        Log.d(TAG, "AI suggestions: ${response.ai_suggestion?.take(100)}...")
                        
                        // If analysis is in response, use it immediately
                        val analysisText = response.analysis_result?.takeIf { it.isNotBlank() }
                            ?: response.message.takeIf { it.isNotBlank() }
                            ?: "CV uploaded successfully. Analysis is being processed..."
                        
                        val suggestionsText = response.ai_suggestion?.takeIf { it.isNotBlank() }
                            ?: "Loading suggestions..."
                        
                        _state.value = _state.value.copy(
                            loading = false,
                            uploadSuccess = true,
                            analysisResult = analysisText,
                            suggestions = suggestionsText
                        )
                        
                        // If analysis wasn't in response, wait a bit and load from database
                        if (response.analysis_result.isNullOrBlank()) {
                            Log.d(TAG, "Analysis not in response, loading from database after delay...")
                            delay(3000) // Wait for backend to process
                            loadCvFeedback()
                        }
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
    
    fun loadCvFeedback() {
        _state.value = _state.value.copy(loading = true, error = null)
        
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
                
                Log.d(TAG, "Loading CV feedback for user: $userId")
                
                val result = aiRepository.getCvFeedback(userId)
                
                result.fold(
                    onSuccess = { feedback ->
                        Log.d(TAG, "CV feedback loaded: Grade=${feedback.grade}")
                        Log.d(TAG, "Analysis length: ${feedback.ai_response?.length ?: 0}")
                        Log.d(TAG, "Suggestions length: ${feedback.ai_suggestion?.length ?: 0}")
                        
                        val analysisText = feedback.ai_response?.takeIf { it.isNotBlank() }
                            ?: "No analysis available. Please upload a CV to get feedback."
                        
                        val suggestionsText = feedback.ai_suggestion?.takeIf { it.isNotBlank() }
                            ?: "No suggestions available."
                        
                        _state.value = _state.value.copy(
                            loading = false,
                            uploadSuccess = true,
                            analysisResult = analysisText,
                            suggestions = suggestionsText
                        )
                    },
                    onFailure = { error ->
                        Log.d(TAG, "No CV feedback found or error: ${error.message}")
                        // Only show error if it's not a "not found" error
                        val errorMessage = if (error.message?.contains("No CV analysis found", ignoreCase = true) == true) {
                            null // Don't show error for "not found" - user just hasn't uploaded yet
                        } else {
                            error.message
                        }
                        _state.value = _state.value.copy(
                            loading = false,
                            error = errorMessage,
                            // Keep existing analysis if available, don't clear it
                            analysisResult = _state.value.analysisResult,
                            suggestions = _state.value.suggestions
                        )
                    }
                )
            } catch (e: Exception) {
                Log.e(TAG, "Exception loading CV feedback", e)
                _state.value = _state.value.copy(
                    loading = false,
                    error = null, // Don't show error on exception
                    // Keep existing analysis if available
                    analysisResult = _state.value.analysisResult,
                    suggestions = _state.value.suggestions
                )
            }
        }
    }
}
