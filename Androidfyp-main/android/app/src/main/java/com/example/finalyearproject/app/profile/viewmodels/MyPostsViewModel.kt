package com.example.finalyearproject.app.profile.viewmodels

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.finalyearproject.app.repository.ProfileRepository
import com.example.finalyearproject.app.repository.models.PostResponse
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch

class MyPostsViewModel(app: Application) : AndroidViewModel(app) {

    private val repo = ProfileRepository(app)

    private val _posts = MutableStateFlow<List<PostResponse>>(emptyList())
    val posts: StateFlow<List<PostResponse>> = _posts

    private val _loading = MutableStateFlow(false)
    val loading: StateFlow<Boolean> = _loading

    fun loadMyPosts() {
        _loading.value = true

        viewModelScope.launch {
            val result = repo.getMyPosts()

            result.onSuccess { list ->
                _posts.value = list ?: emptyList()
            }

            result.onFailure {
                _posts.value = emptyList()
            }

            _loading.value = false
        }
    }

    fun deletePost(id: String) {
        viewModelScope.launch {
            val result = repo.deletePost(id)

            result.onSuccess {
                // REMOVE the post (correct fix)
                _posts.value = _posts.value.filter { it.id != id }
            }
        }
    }
}
