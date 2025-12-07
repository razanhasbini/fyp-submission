package com.example.finalyearproject.app.repository

import android.content.Context
import android.util.Log
import com.example.finalyearproject.app.repository.models.*
import com.example.finalyearproject.app.repository.network.RetrofitClient
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.MultipartBody
import okhttp3.RequestBody.Companion.asRequestBody
import java.io.File

class ProfileRepository(context: Context) {

    private val api = RetrofitClient.create(context)

    /**
     * 🔹 Get current user's profile information
     */
    suspend fun getMe(): Result<UserProfileResponse> = runCatching {
        withContext(Dispatchers.IO) { api.getMe() }
    }

    /**
     * 🔹 Get list of current user's followers
     */
    suspend fun getMyFollowers(): Result<List<UserSummary>> = runCatching {
        withContext(Dispatchers.IO) { api.getMyFollowers() }
    }

    suspend fun getMyFollowing(): Result<List<UserSummary>> = runCatching {
        withContext(Dispatchers.IO) { api.getMyFollowing() }
    }

    suspend fun getFollowerAndFollowingCounts(): Pair<Int, Int> = withContext(Dispatchers.IO) {
        try {
            val followers = api.getMyFollowers().size
            val following = api.getMyFollowing().size
            Log.d("ProfileRepository", "✅ API counts: followers=$followers, following=$following")
            followers to following
        } catch (e: Exception) {
            Log.e("ProfileRepository", "❌ Error counting followers/following: ${e.message}", e)
            0 to 0
        }
    }

    /**
     * 🔹 Update the user's display name
     */
    suspend fun updateName(name: String): Result<Unit> = runCatching {
        withContext(Dispatchers.IO) {
            api.updateUserName(UpdateNameRequest(name))
        }
        Unit
    }

    /**
     * 🔹 Update the user's photo URL in the backend database
     */
    suspend fun updatePhotoUrl(photoUrl: String): Result<Unit> = runCatching {
        withContext(Dispatchers.IO) {
            api.updatePhotoUrl(UpdatePhotoUrlRequest(photoUrl))
            Log.d("ProfileRepository", "✅ Photo URL updated in database: $photoUrl")
        }
        Unit
    }

    /**
     * 🔹 Upload a photo file directly to the backend server
     */
    suspend fun uploadPhoto(file: File): Result<String?> = runCatching {
        withContext(Dispatchers.IO) {
            val req = file.asRequestBody("image/*".toMediaType())
            val part = MultipartBody.Part.createFormData("photo", file.name, req)
            val response = api.updatePhoto(part)
            Log.d("ProfileRepository", "✅ Photo uploaded: ${response.photoUrl}")
            response.photoUrl
        }
    }

    /**
     * 🔹 Delete the user's current profile photo
     */
    suspend fun deletePhoto(): Result<Unit> = runCatching {
        withContext(Dispatchers.IO) {
            api.deletePhoto()
        }
        Unit
    }

    // ------------------------------------------------------
    // 🔥 NEW SECTION — My Posts (based on JWT)
    // ------------------------------------------------------

    /**
     * 🔹 Get current user's posts using JWT `/users/me/posts`
     */
    suspend fun getMyPosts(): Result<List<PostResponse>?> = runCatching {
        withContext(Dispatchers.IO) { api.getMyPosts() }
    }

    suspend fun deletePost(postID: String): Result<Unit> = runCatching {
        withContext(Dispatchers.IO) { api.deletePost(postID) }
        Unit
    }

}
