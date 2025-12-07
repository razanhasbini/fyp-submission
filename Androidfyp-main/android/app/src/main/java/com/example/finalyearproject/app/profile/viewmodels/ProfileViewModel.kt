package com.example.finalyearproject.app.profile.viewmodels

import android.app.Application
import android.util.Log
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.finalyearproject.app.repository.ProfileRepository
import com.example.finalyearproject.app.repository.models.UserProfileResponse
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

data class ProfileUiState(
    val loading: Boolean = false,
    val error: String? = null,
    val profile: UserProfileResponse? = null
)

class ProfileViewModel(app: Application) : AndroidViewModel(app) {
    private val repo = ProfileRepository(app)
    private val _state = MutableStateFlow(ProfileUiState(loading = true))
    val state: StateFlow<ProfileUiState> = _state

    companion object {
        @JvmStatic var triggerRefresh: Boolean = false
    }

    fun loadProfile() {
        Log.d("ProfileViewModel", "🔄 loadProfile() called")
        _state.value = _state.value.copy(loading = true, error = null)
        viewModelScope.launch {
            val result = repo.getMe()
            _state.value = result.fold(
                onSuccess = {
                    Log.d("ProfileViewModel", "✅ Profile loaded: ${it.name}")
                    ProfileUiState(loading = false, profile = it)
                },
                onFailure = {
                    Log.e("ProfileViewModel", "❌ Profile load error: ${it.message}")
                    ProfileUiState(loading = false, error = it.message)
                }
            )
        }
    }

    suspend fun getFollowerAndFollowingCounts(): Pair<Int, Int> {
        Log.d("ProfileViewModel", "📊 Fetching follower/following counts")
        return repo.getFollowerAndFollowingCounts()
    }
}
