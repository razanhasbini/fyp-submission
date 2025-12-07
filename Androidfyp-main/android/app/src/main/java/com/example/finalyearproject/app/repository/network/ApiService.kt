package com.example.finalyearproject.app.repository.network

import com.example.finalyearproject.app.models.Post
import com.example.finalyearproject.app.repository.models.*
import okhttp3.MultipartBody
import retrofit2.Response
import retrofit2.http.*

interface ApiService {

    // ===== Auth =====
    @POST("send-verification-code")
    suspend fun sendVerificationCode(@Body request: SendVerificationCodeRequest): SendVerificationCodeResponse

    @POST("verify-email-code")
    suspend fun verifyEmailCode(@Body request: VerifyEmailCodeRequest): VerifyEmailCodeResponse

    @POST("complete-registration")
    suspend fun completeRegistration(@Body request: CompleteRegistrationRequest): CompleteRegistrationResponse

    @POST("login")
    suspend fun signIn(@Body request: SignInRequest): SignInResponse

    @POST("send-password-reset")
    suspend fun sendPasswordReset(@Body request: SendPasswordResetRequest): SendPasswordResetResponse

    @POST("reset-password")
    suspend fun resetPassword(@Body request: ResetPasswordRequest): SendPasswordResetResponse

    // ===== Profile =====
    @GET("users/me")
    suspend fun getMe(): UserProfileResponse

    @GET("users/me/followers")
    suspend fun getMyFollowers(): List<UserSummary>

    @GET("users/me/following")
    suspend fun getMyFollowing(): List<UserSummary>

    @PUT("users/name")
    suspend fun updateUserName(@Body body: UpdateNameRequest): BasicOkResponse

    @Multipart
    @PUT("users/photo")
    suspend fun updatePhoto(@Part photo: MultipartBody.Part): UpdatePhotoResponse

    @DELETE("users/photo")
    suspend fun deletePhoto(): BasicOkResponse

    @PUT("users/photo")
    suspend fun updatePhotoUrl(@Body request: UpdatePhotoUrlRequest): BasicOkResponse

    // ===== Posts =====
    @GET("posts/all")
    suspend fun getAllPosts(): List<PostResponse>

    @GET("users/feed")
    suspend fun getFollowingFeed(): List<PostResponse>

    @GET("posts/{postID}")
    suspend fun getPostById(@Path("postID") postID: String): PostResponse

    @POST("posts")
    suspend fun createPost(@Body request: CreatePostRequest): BasicOkResponse

    @DELETE("posts/{postID}")
    suspend fun deletePost(@Path("postID") postID: String): BasicOkResponse

    // ===== Likes =====
    @POST("posts/{postID}/like")
    suspend fun likePost(@Path("postID") postID: String): BasicOkResponse

    @POST("posts/{postID}/unlike")
    suspend fun unlikePost(@Path("postID") postID: String): BasicOkResponse

    // ===== Comments =====
    @POST("posts/{postID}/comment")
    suspend fun addComment(
        @Path("postID") postID: String,
        @Body request: AddCommentRequest
    ): BasicOkResponse

    @GET("posts/{postID}/comments")
    suspend fun getComments(@Path("postID") postID: String): List<CommentResponse>

    @DELETE("posts/{postID}/comments/{commentID}")
    suspend fun deleteComment(
        @Path("postID") postID: String,
        @Path("commentID") commentID: String
    ): BasicOkResponse

    // ===== User by username =====
    @GET("users/{username}")
    suspend fun getUserByUsername(@Path("username") username: String): UserProfileResponse

    @GET("users/{username}/posts")
    suspend fun getUserPosts(@Path("username") username: String): List<Post>

    @GET("users/{username}/followers")
    suspend fun getFollowers(@Path("username") username: String): List<UserSummary>

    @GET("users/{username}/following")
    suspend fun getFollowing(@Path("username") username: String): List<UserSummary>

    // ===== Follow/Unfollow =====
    @POST("follow/{userID}")
    suspend fun followUser(@Path("userID") userID: String): BasicOkResponse

    @DELETE("unfollow/{userID}")
    suspend fun unfollowUser(@Path("userID") userID: String): BasicOkResponse

    // ===== User Info by ID =====
    @GET("users/id/{userID}")
    suspend fun getUserById(@Path("userID") userID: String): UserProfileResponse

    // ===== Likes Count & Users =====
    @GET("posts/{postID}/likes")
    suspend fun getLikesCount(@Path("postID") postID: String): LikesCountResponse

    @GET("posts/{postID}/likes/users")
    suspend fun getPostLikes(@Path("postID") postID: String): List<LikeUserResponse>

    // ===== My Posts =====
    @GET("users/me/posts")
    suspend fun getMyPosts(): List<PostResponse>
    @GET("notifications")
    suspend fun getNotifications(): List<ApiNotification>

    @GET("notifications/unread-count")
    suspend fun getUnreadCount(): UnreadCountResponse

    @PUT("notifications/{notificationID}/read")
    suspend fun markNotificationRead(@Path("notificationID") notificationID: Long): BasicOkResponse

    @PUT("notifications/mark-all-read")
    suspend fun markAllNotificationsRead(): BasicOkResponse

}
