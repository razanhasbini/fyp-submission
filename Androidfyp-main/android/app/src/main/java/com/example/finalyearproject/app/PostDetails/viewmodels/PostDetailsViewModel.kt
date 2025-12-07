package com.example.finalyearproject.app.postdetails.viewmodel

import androidx.lifecycle.*
import com.example.finalyearproject.app.models.Post
import com.example.finalyearproject.app.repository.PostRepository
import com.example.finalyearproject.app.repository.models.Comment
import com.example.finalyearproject.app.repository.models.LikeUserResponse
import kotlinx.coroutines.launch
import kotlin.Result

class PostDetailsViewModel(private val repo: PostRepository) : ViewModel() {

    private var currentPostId: String = ""

    private val _post = MutableLiveData<Post?>()
    val post: LiveData<Post?> = _post

    private val _comments = MutableLiveData<List<Comment>>()
    val comments: LiveData<List<Comment>> = _comments

    private val _likesList = MutableLiveData<List<LikeUserResponse>>()
    val likesList: LiveData<List<LikeUserResponse>> = _likesList

    private val _actionResult = MutableLiveData<Result<Unit>>()
    val actionResult: LiveData<Result<Unit>> = _actionResult

    fun loadPost(postId: String) {
        currentPostId = postId
        viewModelScope.launch {
            val result = repo.getPostById(postId)
            _post.value = result.getOrNull()
            loadComments()
        }
    }

    fun loadComments() {
        viewModelScope.launch {
            val result = repo.getComments(currentPostId)
            _comments.value = result.getOrDefault(emptyList())
        }
    }

    fun addComment(text: String) {
        if (text.isBlank()) return
        viewModelScope.launch {
            val result = repo.addComment(currentPostId, text)
            _actionResult.value = result
            if (result.isSuccess) {
                loadComments()
                // Reload post to update comment count
                val postResult = repo.getPostById(currentPostId)
                _post.value = postResult.getOrNull()
            }
        }
    }

    fun deleteComment(commentId: String) {
        viewModelScope.launch {
            val result = repo.deleteComment(currentPostId, commentId)
            _actionResult.value = result
            if (result.isSuccess) {
                loadComments()
                // Reload post to update comment count
                val postResult = repo.getPostById(currentPostId)
                _post.value = postResult.getOrNull()
            }
        }
    }

    fun fetchLikesList(postId: String) {
        viewModelScope.launch {
            val result = repo.getPostLikes(postId)
            _likesList.value = result.getOrDefault(emptyList())
        }
    }
}