package com.example.finalyearproject.app.repository.network

import okhttp3.MultipartBody
import okhttp3.RequestBody
import retrofit2.Response
import retrofit2.http.*

interface AiService {
    
    @GET("/v1/health")
    suspend fun healthCheck(): Response<HealthCheckResponse>
    
    @Multipart
    @POST("/v1/cv/upload")
    suspend fun uploadCV(
        @Part("user_id") userId: RequestBody,
        @Part file: MultipartBody.Part
    ): Response<CvUploadResponse>
    
    @POST("/v1/cv/ingest")
    suspend fun ingestCV(@Body request: CvIngestRequest): Response<CvIngestResponse>
    
    @POST("/v1/session/start")
    suspend fun startSession(@Body request: StartSessionRequest): Response<StartSessionResponse>
    
    @POST("/v1/session/end")
    suspend fun endSession(@Body request: EndSessionRequest): Response<EndSessionResponse>
    
    @POST("/v1/next-question")
    suspend fun getNextQuestion(@Body request: NextQuestionRequest): Response<NextQuestionResponse>
    
    @POST("/v1/evaluate")
    suspend fun evaluateAnswer(@Body request: EvaluateRequest): Response<EvaluateResponse>
    
    @POST("/v1/feedback")
    suspend fun submitFeedback(@Body request: FeedbackRequest): Response<FeedbackResponse>
    
    @POST("/v1/kb/seed")
    suspend fun seedKnowledgeBase(@Body request: SeedKBRequest): Response<SeedKBResponse>
    
    @GET("/v1/user/feedback/cv")
    suspend fun getCvFeedback(@Query("user_id") userId: Long): Response<CvFeedbackResponse>
    
    @GET("/v1/user/scores")
    suspend fun getUserScores(@Query("user_id") userId: Long): Response<UserScoresResponse>
    
    @GET("/v1/user/feedback/technical")
    suspend fun getTechnicalFeedback(@Query("user_id") userId: Long): Response<TechnicalFeedbackResponse>
    
    @GET("/v1/user/feedback/behavioral")
    suspend fun getBehavioralFeedback(@Query("user_id") userId: Long): Response<BehavioralFeedbackResponse>
}

data class HealthCheckResponse(val ok: Boolean)

data class CvUploadResponse(
    val message: String,
    val cv_id: Long? = null,
    val analysis_result: String? = null,
    val ai_suggestion: String? = null
)

data class UserScoresResponse(
    val user_id: Long,
    val technical_score: Double,
    val behavioral_score: Double,
    val cv_analysis_score: Double
)

data class CvIngestRequest(
    val user_id: Long,
    val text: String
)

data class CvIngestResponse(
    val message: String,
    val cv_id: Long
)

data class StartSessionRequest(
    val user_id: Long,
    val major: String? = null
)

data class StartSessionResponse(
    val session_id: String,
    val message: String
)

data class EndSessionRequest(
    val session_id: String,
    val user_id: Long
)

data class EndSessionResponse(
    val grade: Int,
    val message: String
)

data class NextQuestionRequest(
    val user_id: Long,
    val domain: String,
    val difficulty: String = "medium"
)

data class NextQuestionResponse(
    val question: String
)

data class EvaluateRequest(
    val session_id: String,
    val user_id: Long,
    val domain: String,
    val question: String,
    val answer: String
)

data class EvaluateResponse(
    val score: Double,
    val feedback: String
)

data class FeedbackRequest(
    val session_id: String,
    val user_id: Long,
    val overall_score: Double,
    val technical_score: Double,
    val communication_score: Double,
    val confidence_score: Double,
    val text_feedback: String? = null
)

data class FeedbackResponse(
    val message: String
)

data class SeedKBRequest(
    val domain: String,
    val items: List<String>
)

data class SeedKBResponse(
    val message: String,
    val count: Int
)

data class CvFeedbackResponse(
    val grade: Int?,
    val ai_response: String?,
    val ai_suggestion: String?
)

data class TechnicalFeedbackResponse(
    val user_id: Long,
    val technical_score: Double,
    val overall_score: Double,
    val feedback: String?,
    val created_at: String?,
    val message: String?
)

data class BehavioralFeedbackResponse(
    val user_id: Long,
    val communication_score: Double,
    val confidence_score: Double,
    val overall_score: Double,
    val feedback: String?,
    val created_at: String?,
    val message: String?
)
