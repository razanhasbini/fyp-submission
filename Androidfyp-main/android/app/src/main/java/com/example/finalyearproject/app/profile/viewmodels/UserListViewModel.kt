package com.example.finalyearproject.app.profile.viewmodels

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.finalyearproject.app.repository.ProfileRepository
import com.example.finalyearproject.app.repository.models.UserSummary
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

enum class ListType { FOLLOWERS, FOLLOWING }

data class UserListState(
    val loading: Boolean = false,
    val error: String? = null,
    val users: List<UserSummary> = emptyList()
)

class UserListViewModel(app: Application) : AndroidViewModel(app) {
    private val repo = ProfileRepository(app)
    private val _state = MutableStateFlow(UserListState())
    val state: StateFlow<UserListState> = _state

    fun load(type: ListType) {
        _state.value = UserListState(loading = true)

        viewModelScope.launch {
            try {
                when (type) {
                    ListType.FOLLOWERS -> {
                        val followersResult = repo.getMyFollowers()
                        val followingResult = repo.getMyFollowing()

                        if (followersResult.isSuccess && followingResult.isSuccess) {
                            val followers = followersResult.getOrThrow()
                            val following = followingResult.getOrThrow()

                            val followingIds = following.map { it.id }.toSet()

                            // ✅ Mark mutual followers as followed
                            val fixedFollowers = followers.onEach { user ->
                                user.isFollowing = followingIds.contains(user.id)
                            }

                            _state.value = UserListState(users = fixedFollowers)
                        } else {
                            val errorMsg = followersResult.exceptionOrNull()?.message
                                ?: followingResult.exceptionOrNull()?.message
                                ?: "Failed to fetch users"
                            _state.value = UserListState(error = errorMsg)
                        }
                    }

                    ListType.FOLLOWING -> {
                        val res = repo.getMyFollowing()
                        _state.value = res.fold(
                            onSuccess = { list ->
                                list.onEach { it.isFollowing = true }
                                UserListState(users = list)
                            },
                            onFailure = { UserListState(error = it.message) }
                        )
                    }
                }
            } catch (e: Exception) {
                _state.value = UserListState(error = e.message ?: "Unknown error")
            }
        }
    }
}
