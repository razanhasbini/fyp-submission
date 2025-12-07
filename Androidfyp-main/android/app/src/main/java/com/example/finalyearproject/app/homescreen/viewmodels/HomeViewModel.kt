package com.example.finalyearproject.app.ui.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.viewModelScope
import com.example.finalyearproject.app.models.Post
import com.example.finalyearproject.app.repository.PostRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

sealed class FeedTab {
    object ForYou : FeedTab()
    object Following : FeedTab()
}

sealed class HomeUiState {
    object Loading : HomeUiState()
    data class Success(val posts: List<Post>) : HomeUiState()
    data class Error(val message: String) : HomeUiState()
    object Empty : HomeUiState()
}

class HomeViewModel(private val repo: PostRepository) : ViewModel() {

    private val _selectedTab = MutableStateFlow<FeedTab>(FeedTab.ForYou)
    val selectedTab: StateFlow<FeedTab> = _selectedTab.asStateFlow()

    private val _uiState = MutableStateFlow<HomeUiState>(HomeUiState.Loading)
    val uiState: StateFlow<HomeUiState> = _uiState.asStateFlow()

    init { loadFeed() }

    fun onTabSelected(tab: FeedTab) {
        if (_selectedTab.value != tab) {
            _selectedTab.value = tab
            loadFeed()
        }
    }

    fun refresh() = loadFeed()

    private fun loadFeed() {
        viewModelScope.launch {
            _uiState.value = HomeUiState.Loading
            try {
                val result = when (_selectedTab.value) {
                    is FeedTab.ForYou -> repo.getAllPosts()
                    is FeedTab.Following -> repo.getFollowingFeed()
                }

                result.fold(
                    onSuccess = { posts ->
                        val safePosts = posts ?: emptyList() // ✅ null-safe
                        if (safePosts.isEmpty()) {
                            if (_selectedTab.value is FeedTab.Following) {
                                _uiState.value =
                                    HomeUiState.Error("No posts from people you follow yet.")
                            } else {
                                _uiState.value = HomeUiState.Empty
                            }
                        } else {
                            _uiState.value = HomeUiState.Success(safePosts)
                        }
                    },
                    onFailure = { err ->
                        _uiState.value = HomeUiState.Error(err.message ?: "Failed to load posts")
                    }
                )
            } catch (e: Exception) {
                _uiState.value = HomeUiState.Error(e.message ?: "Unexpected error occurred")
            }
        }
    }
}

class HomeViewModelFactory(private val repo: PostRepository) : ViewModelProvider.Factory {
    @Suppress("UNCHECKED_CAST")
    override fun <T : ViewModel> create(modelClass: Class<T>): T {
        if (modelClass.isAssignableFrom(HomeViewModel::class.java))
            return HomeViewModel(repo) as T
        throw IllegalArgumentException("Unknown ViewModel class")
    }
}
