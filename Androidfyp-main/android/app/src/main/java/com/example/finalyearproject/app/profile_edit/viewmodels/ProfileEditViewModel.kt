package com.example.finalyearproject.app.profile_edit.viewmodels

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.finalyearproject.app.repository.ProfileRepository
import com.example.finalyearproject.app.repository.models.UserData
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import java.io.File

/**
 * Represents the current state of the profile editing screen.
 */
data class EditState(
    val saving: Boolean = false,
    val error: String? = null,
    val photoUrl: String? = null,
    val successMessage: String? = null,
    val user: UserData? = null
)

/**
 * ViewModel responsible for handling profile edit logic:
 * - Loading user info
 * - Updating name
 * - Uploading new profile photo
 * - Updating photo URL
 * - Deleting profile photo
 */
class ProfileEditViewModel(app: Application) : AndroidViewModel(app) {

    private val repo = ProfileRepository(app.applicationContext)

    private val _state = MutableStateFlow(EditState())
    val state: StateFlow<EditState> = _state

    /**
     * 🔹 Load the currently logged-in user's profile data using the API (getMe)
     */
    fun loadUserProfile() {
        viewModelScope.launch {
            val res = repo.getMe()
            _state.value = res.fold(
                onSuccess = {
                    // Convert API response to simple UserData
                    val userData = UserData(
                        id = it.id,
                        email = it.email,
                        username = it.username,
                        name = it.name
                    )
                    EditState(user = userData)
                },
                onFailure = { EditState(error = it.message ?: "Failed to load user profile") }
            )
        }
    }

    /**
     * 🔹 Update the user's name.
     */
    fun updateName(name: String) {
        _state.value = _state.value.copy(saving = true, error = null, successMessage = null)
        viewModelScope.launch {
            val res = repo.updateName(name)
            _state.value = res.fold(
                onSuccess = { _state.value.copy(successMessage = "Name updated successfully") },
                onFailure = { _state.value.copy(error = it.message ?: "Failed to update name") }
            )
        }
    }

    /**
     * 🔹 Update the user's profile photo URL in the backend database.
     */
    fun updatePhotoUrl(photoUrl: String) {
        _state.value = _state.value.copy(saving = true, error = null, successMessage = null)
        viewModelScope.launch {
            val res = repo.updatePhotoUrl(photoUrl)
            _state.value = res.fold(
                onSuccess = {
                    _state.value.copy(
                        photoUrl = photoUrl,
                        successMessage = "Photo updated successfully"
                    )
                },
                onFailure = { _state.value.copy(error = it.message ?: "Failed to update photo URL") }
            )
        }
    }

    /**
     * 🔹 Upload a new profile photo directly to the backend server.
     */
    fun uploadPhoto(file: File) {
        _state.value = _state.value.copy(saving = true, error = null, successMessage = null)
        viewModelScope.launch {
            val res = repo.uploadPhoto(file)
            _state.value = res.fold(
                onSuccess = {
                    _state.value.copy(photoUrl = it, successMessage = "Photo uploaded successfully")
                },
                onFailure = { _state.value.copy(error = it.message ?: "Failed to upload photo") }
            )
        }
    }

    /**
     * 🔹 Delete the current profile photo.
     */
    fun deletePhoto() {
        _state.value = _state.value.copy(saving = true, error = null, successMessage = null)
        viewModelScope.launch {
            val res = repo.deletePhoto()
            _state.value = res.fold(
                onSuccess = { _state.value.copy(successMessage = "Photo removed successfully") },
                onFailure = { _state.value.copy(error = it.message ?: "Failed to delete photo") }
            )
        }
    }
}
