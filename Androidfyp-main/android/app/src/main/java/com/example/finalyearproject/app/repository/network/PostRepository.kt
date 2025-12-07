package com.example.finalyearproject.app.repository

import com.example.finalyearproject.app.models.Post
import com.example.finalyearproject.app.repository.models.*
import com.example.finalyearproject.app.repository.network.ApiService
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.withContext

class PostRepository(
    private val apiService: ApiService
) {

    /**
     * Fetch all posts (For You feed) with complete user data
     */
    suspend fun getAllPosts(): Result<List<Post>> = withContext(Dispatchers.IO) {
        try {
            val response = apiService.getAllPosts()
            val postsResponse = response ?: emptyList()

            val postsWithUserData = postsResponse.map { postResponse ->
                async { enrichPostWithUserData(postResponse) }
            }.awaitAll()

            val validPosts = postsWithUserData.filterNotNull()
            Result.success(validPosts)
        } catch (e: Exception) {
            if (e.message?.contains("was null") == true) {
                Result.success(emptyList())
            } else {
                Result.failure(e)
            }
        }
    }

    /**
     * Fetch following feed with complete user data
     */
    suspend fun getFollowingFeed(): Result<List<Post>> = withContext(Dispatchers.IO) {
        try {
            val response = apiService.getFollowingFeed()
            val postsResponse = response ?: emptyList()

            val postsWithUserData = postsResponse.map { postResponse ->
                async { enrichPostWithUserData(postResponse) }
            }.awaitAll()

            val validPosts = postsWithUserData.filterNotNull()
            Result.success(validPosts)
        } catch (e: Exception) {
            if (e.message?.contains("was null") == true) {
                Result.success(emptyList())
            } else {
                Result.failure(e)
            }
        }
    }

    /**
     * Fetch a single post by ID with user data
     */
    suspend fun getPostById(postId: String): Result<Post> = withContext(Dispatchers.IO) {
        try {
            val postResponse = apiService.getPostById(postId)
            val enrichedPost = enrichPostWithUserData(postResponse)

            if (enrichedPost != null) Result.success(enrichedPost)
            else Result.failure(Exception("Failed to load post data"))
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Enrich a post response with user info and backend metadata
     */
    private suspend fun enrichPostWithUserData(postResponse: PostResponse): Post? {
        return try {
            val userResponse = apiService.getUserById(postResponse.user_id)

            Post(
                id = postResponse.id,
                userId = postResponse.user_id,
                userName = userResponse.name ?: userResponse.username ?: "Unknown User",
                userPhotoUrl = userResponse.photoUrl,
                content = postResponse.content,
                photoUrl = postResponse.photo_url,
                createdAt = postResponse.created_at ?: "",

                // ✅ Preserve backend state
                likesCount = postResponse.likes_count ?: 0,
                commentsCount = postResponse.comments_count ?: 0,
                isLiked = postResponse.is_liked ?: false,
                isFollowing = postResponse.is_following ?: false,
                isOwner = postResponse.is_owner ?: false  // ✅ Get isOwner from backend
            )
        } catch (e: Exception) {
            e.printStackTrace()
            null
        }
    }

    suspend fun getPostLikes(postId: String): Result<List<LikeUserResponse>> {
        return try {
            val response = apiService.getPostLikes(postId)
            Result.success(response)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Create a new post
     */
    suspend fun createPost(content: String, photoUrl: String?): Result<Unit> =
        withContext(Dispatchers.IO) {
            try {
                val request = CreatePostRequest(content = content, photo_url = photoUrl)
                apiService.createPost(request)
                Result.success(Unit)
            } catch (e: Exception) {
                Result.failure(e)
            }
        }

    /**
     * Delete a post
     */
    suspend fun deletePost(postId: String): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            apiService.deletePost(postId)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Like a post
     */
    suspend fun likePost(postId: String): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            apiService.likePost(postId)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Unlike a post
     */
    suspend fun unlikePost(postId: String): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            apiService.unlikePost(postId)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Get total likes for a post
     */
    suspend fun getLikesCount(postId: String): Result<Int> = withContext(Dispatchers.IO) {
        try {
            val response = apiService.getLikesCount(postId)
            Result.success(response.count)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Add a comment
     */
    suspend fun addComment(postId: String, content: String): Result<Unit> =
        withContext(Dispatchers.IO) {
            try {
                val request = AddCommentRequest(content = content)
                apiService.addComment(postId, request)
                Result.success(Unit)
            } catch (e: Exception) {
                Result.failure(e)
            }
        }

    /**
     * Get comments for a post (with user info)
     */
    suspend fun getComments(postId: String): Result<List<Comment>> =
        withContext(Dispatchers.IO) {
            try {
                val commentsResponse = apiService.getComments(postId) ?: emptyList()

                // If backend sends [] — still fine.
                if (commentsResponse.isEmpty()) {
                    return@withContext Result.success(emptyList())
                }

                val commentsWithUserData = commentsResponse.map { commentResponse ->
                    async {
                        try {
                            val userResponse = apiService.getUserById(commentResponse.user_id)
                            Comment(
                                id = commentResponse.id,
                                postId = postId,
                                userId = commentResponse.user_id,
                                userName = userResponse.name ?: "Unknown User",
                                userPhotoUrl = userResponse.photoUrl,
                                content = commentResponse.content,
                                createdAt = commentResponse.created_at ?: "",
                                isOwner = commentResponse.is_owner ?: false
                            )
                        } catch (e: Exception) {
                            null
                        }
                    }
                }.awaitAll().filterNotNull()

                Result.success(commentsWithUserData)
            } catch (e: Exception) {
                // Return empty list on harmless server or network issues, not failure
                if (e.message?.contains("EOF") == true || e.message?.contains("empty") == true)
                    Result.success(emptyList())
                else
                    Result.failure(e)
            }
        }

    /**
     * Follow a user
     */
    suspend fun followUser(userId: String): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            apiService.followUser(userId)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Unfollow a user
     */
    suspend fun unfollowUser(userId: String): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            apiService.unfollowUser(userId)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Delete a comment
     */
    suspend fun deleteComment(postId: String, commentId: String): Result<Unit> =
        withContext(Dispatchers.IO) {
            try {
                apiService.deleteComment(postId, commentId)
                Result.success(Unit)
            } catch (e: Exception) {
                Result.failure(e)
            }
        }
}