package com.example.finalyearproject.app.repository

import android.content.Context
import android.net.Uri
import android.provider.OpenableColumns
import android.util.Log
import com.example.finalyearproject.app.repository.network.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.coroutines.withTimeout
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.MultipartBody
import okhttp3.RequestBody.Companion.asRequestBody
import okhttp3.RequestBody.Companion.toRequestBody
import java.io.File
import java.io.FileOutputStream
import java.io.InputStream

/**
 * Repository for AI Interview Service interactions
 * 
 * This repository handles all communication with the wazzafak-ai backend
 * including CV uploads, interview questions, evaluations, and feedback.
 */
class AiRepository(private val context: Context) {
    
    // Get service instance dynamically to ensure it uses the latest URL
    private val aiService: AiService
        get() = AiRetrofitClient.aiService
    
    companion object {
        private const val TAG = "AiRepository"
    }
    
    // ==================== Health Check ====================
    
    /**
     * Health check for AI service
     */
    suspend fun healthCheck(): Result<Boolean> {
        return try {
            val response = aiService.healthCheck()
            if (response.isSuccessful && response.body() != null) {
                Result.success(response.body()!!.ok)
            } else {
                Result.success(false)
            }
        } catch (e: Exception) {
            Log.e(TAG, "Health check failed", e)
            Result.failure(e)
        }
    }
    
    // ==================== CV Management ====================
    
    /**
     * Upload CV file to AI service for analysis
     * 
     * @param uri File URI from file picker
     * @param userId User ID for the CV
     * @return CvUploadResponse with analysis results
     */
    suspend fun uploadCV(uri: Uri, userId: Long): Result<CvUploadResponse> {
        return try {
            Log.d(TAG, "Starting CV upload for user: $userId")
            
            // Convert URI to File for multipart upload with proper file type detection
            val (file, mimeType) = try {
                uriToFile(uri)
            } catch (e: Exception) {
                Log.e(TAG, "File conversion failed", e)
                return Result.failure(Exception("File error: ${e.message}"))
            }
            
            Log.d(TAG, "File prepared: ${file.name}, size: ${file.length()} bytes, MIME: $mimeType")
            
            val requestFile = file.asRequestBody(mimeType.toMediaTypeOrNull())
            val body = MultipartBody.Part.createFormData("file", file.name, requestFile)
            
            // Create user_id as RequestBody
            val userIdBody = userId.toString().toRequestBody("text/plain".toMediaTypeOrNull())
            
            Log.d(TAG, "Sending request to AI service...")
            Log.d(TAG, "Base URL: ${com.example.finalyearproject.app.repository.network.AiRetrofitClient.getBaseUrl()}")
            Log.d(TAG, "File size: ${file.length()} bytes")
            
            // Test connection first with timeout
            try {
                Log.d(TAG, "Testing backend connection...")
                val healthCheck = withContext(Dispatchers.IO) {
                    withTimeout(10000) {
                        aiService.healthCheck()
                    }
                }
                if (!healthCheck.isSuccessful) {
                    Log.e(TAG, "Health check failed: ${healthCheck.code()} - ${healthCheck.message()}")
                    file.delete()
                    return Result.failure(Exception("Cannot connect to backend (HTTP ${healthCheck.code()}). Please check if server is running at ${com.example.finalyearproject.app.repository.network.AiRetrofitClient.getBaseUrl()}"))
                }
                Log.d(TAG, "✅ Health check passed")
            } catch (healthErr: Exception) {
                Log.e(TAG, "Health check error", healthErr)
                file.delete()
                val errorMsg = when {
                    healthErr.message?.contains("timeout", ignoreCase = true) == true ->
                        "Connection timeout. Please check if the backend server is running at ${com.example.finalyearproject.app.repository.network.AiRetrofitClient.getBaseUrl()}"
                    healthErr.message?.contains("connection", ignoreCase = true) == true ->
                        "Cannot connect to backend. Please ensure the server is running at ${com.example.finalyearproject.app.repository.network.AiRetrofitClient.getBaseUrl()}"
                    else ->
                        "Connection error: ${healthErr.message}. Backend URL: ${com.example.finalyearproject.app.repository.network.AiRetrofitClient.getBaseUrl()}"
                }
                return Result.failure(Exception(errorMsg))
            }
            
            // Retry logic for connection issues
            var lastError: Exception? = null
            var retryCount = 0
            val maxRetries = 2
            
            while (retryCount <= maxRetries) {
                try {
                    if (retryCount > 0) {
                        Log.d(TAG, "Retrying upload (attempt ${retryCount + 1}/${maxRetries + 1})...")
                        kotlinx.coroutines.delay(2000) // Wait 2 seconds before retry
                    }
                    
                    val response = aiService.uploadCV(userIdBody, body)
                    
                    if (response.isSuccessful && response.body() != null) {
                        val result = response.body()!!
                        Log.d(TAG, "CV upload successful: ${result.message}, CV ID: ${result.cv_id}")
                        // Clean up temp file
                        file.delete()
                        return Result.success(result)
                    } else {
                        val errorBody = try {
                            response.errorBody()?.string() ?: "Unknown error"
                        } catch (e: Exception) {
                            "Error reading response: ${e.message}"
                        }
                        Log.e(TAG, "Upload failed: HTTP ${response.code()}, $errorBody")
                        
                        // Don't retry on 4xx errors (client errors)
                        if (response.code() in 400..499) {
                            file.delete()
                            return Result.failure(Exception("Upload failed: ${response.code()} - $errorBody"))
                        }
                        
                        lastError = Exception("Upload failed: ${response.code()} - $errorBody")
                    }
                } catch (e: Exception) {
                    Log.e(TAG, "Upload attempt ${retryCount + 1} failed", e)
                    lastError = e
                    
                    // Check if it's a connection error that we should retry
                    val isConnectionError = e.message?.contains("connection reset", ignoreCase = true) == true ||
                                          e.message?.contains("connection refused", ignoreCase = true) == true ||
                                          e.message?.contains("timeout", ignoreCase = true) == true ||
                                          e.message?.contains("network", ignoreCase = true) == true ||
                                          e is java.net.SocketException ||
                                          e is java.net.ConnectException ||
                                          e is java.net.UnknownHostException
                    
                    if (!isConnectionError || retryCount >= maxRetries) {
                        file.delete()
                        val errorMsg = when {
                            e.message?.contains("connection reset", ignoreCase = true) == true ->
                                "Connection lost during upload. Please check your network connection and try again."
                            e.message?.contains("timeout", ignoreCase = true) == true ->
                                "Upload timeout. The file might be too large or connection too slow. Please try again."
                            e.message?.contains("connection refused", ignoreCase = true) == true ->
                                "Cannot connect to backend. Please ensure the server is running at ${com.example.finalyearproject.app.repository.network.AiRetrofitClient.getBaseUrl()}"
                            else ->
                                e.message ?: "Upload error: ${e.javaClass.simpleName}"
                        }
                        return Result.failure(Exception(errorMsg))
                    }
                }
                
                retryCount++
            }
            
            // All retries failed
            file.delete()
            val finalError = lastError?.message ?: "Upload failed after ${maxRetries + 1} attempts"
            Log.e(TAG, finalError)
            return Result.failure(Exception(finalError))
        } catch (e: Exception) {
            Log.e(TAG, "Error uploading CV", e)
            val errorMsg = when {
                e.message?.contains("connection reset", ignoreCase = true) == true ->
                    "Connection lost during upload. Please check your network and try again."
                e.message?.contains("timeout", ignoreCase = true) == true ->
                    "Upload timeout. Please try again or use a smaller file."
                else ->
                    e.message ?: "Unknown error occurred"
            }
            Result.failure(Exception(errorMsg))
        }
    }
    
    /**
     * Ingest CV text directly (without file upload)
     */
    suspend fun ingestCV(userId: Long, text: String): Result<CvIngestResponse> {
        return try {
            val request = CvIngestRequest(user_id = userId, text = text)
            val response = aiService.ingestCV(request)
            
            if (response.isSuccessful && response.body() != null) {
                Result.success(response.body()!!)
            } else {
                val errorBody = response.errorBody()?.string() ?: "Unknown error"
                Log.e(TAG, "Ingest CV failed: $errorBody")
                Result.failure(Exception("Failed to ingest CV: $errorBody"))
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error ingesting CV", e)
            Result.failure(e)
        }
    }
    
    // ==================== Session Management ====================
    
    /**
     * Start a new practice session
     * 
     * @param userId User ID for the session
     * @param major Optional major/field for the interview
     * @return StartSessionResponse with session_id
     */
    suspend fun startSession(userId: Long, major: String? = null): Result<StartSessionResponse> {
        return try {
            Log.d(TAG, "Starting session for user: $userId, major: $major")
            Log.d(TAG, "Backend URL: ${com.example.finalyearproject.app.repository.network.AiRetrofitClient.getBaseUrl()}")
            val request = StartSessionRequest(user_id = userId, major = major)
            val response = withContext(Dispatchers.IO) {
                withTimeout(15000) {
                    aiService.startSession(request)
                }
            }
            
            if (response.isSuccessful && response.body() != null) {
                val result = response.body()!!
                Log.d(TAG, "✅ Session started: ${result.session_id}")
                Result.success(result)
            } else {
                val errorBody = try {
                    response.errorBody()?.string() ?: "Unknown error"
                } catch (e: Exception) {
                    "Error reading response: ${e.message}"
                }
                Log.e(TAG, "❌ Start session failed: HTTP ${response.code()} - $errorBody")
                Result.failure(Exception("Failed to start session (HTTP ${response.code()}): $errorBody"))
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error starting session", e)
            Result.failure(e)
        }
    }
    
    /**
     * End the current practice session
     * 
     * @param sessionId The session ID to end
     * @param userId User ID for the session
     * @return EndSessionResponse with final grade and feedback
     */
    suspend fun endSession(sessionId: String, userId: Long): Result<EndSessionResponse> {
        return try {
            Log.d(TAG, "Ending session: $sessionId")
            val request = EndSessionRequest(session_id = sessionId, user_id = userId)
            val response = aiService.endSession(request)
            
            if (response.isSuccessful && response.body() != null) {
                val result = response.body()!!
                Log.d(TAG, "Session ended: Grade = ${result.grade}")
                Result.success(result)
            } else {
                val errorBody = response.errorBody()?.string() ?: "Unknown error"
                Log.e(TAG, "End session failed: $errorBody")
                Result.failure(Exception("Failed to end session: $errorBody"))
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error ending session", e)
            Result.failure(e)
        }
    }
    
    // ==================== Interview Questions ====================
    
    /**
     * Get next interview question
     * 
     * @param userId User ID 
     * @param domain Interview domain (e.g., "Software Engineering", "Marketing")
     * @param difficulty Question difficulty: "easy", "medium", "hard"
     */
    suspend fun getNextQuestion(
        userId: Long,
        domain: String,
        difficulty: String = "medium"
    ): Result<NextQuestionResponse> {
        return try {
            Log.d(TAG, "Getting next question - User: $userId, Domain: $domain, Difficulty: $difficulty")
            val request = NextQuestionRequest(
                user_id = userId,
                domain = domain,
                difficulty = difficulty
            )
            val response = withContext(Dispatchers.IO) {
                withTimeout(15000) {
                    aiService.getNextQuestion(request)
                }
            }
            
            if (response.isSuccessful && response.body() != null) {
                val result = response.body()!!
                Log.d(TAG, "✅ Question received: ${result.question}")
                Result.success(result)
            } else {
                val errorBody = try {
                    response.errorBody()?.string() ?: "Unknown error"
                } catch (e: Exception) {
                    "Error reading response: ${e.message}"
                }
                Log.e(TAG, "❌ Get question failed: HTTP ${response.code()} - $errorBody")
                Result.failure(Exception("Failed to get question (HTTP ${response.code()}): $errorBody"))
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error getting question", e)
            Result.failure(e)
        }
    }
    
    /**
     * Evaluate interview answer
     * 
     * @param sessionId Current session ID
     * @param userId User ID
     * @param domain Interview domain
     * @param question The question that was asked
     * @param answer User's answer to evaluate
     * @return EvaluateResponse with scores and feedback
     */
    suspend fun evaluateAnswer(
        sessionId: String,
        userId: Long,
        domain: String,
        question: String,
        answer: String
    ): Result<EvaluateResponse> {
        return try {
            val request = EvaluateRequest(
                session_id = sessionId,
                user_id = userId,
                domain = domain,
                question = question,
                answer = answer
            )
            val response = aiService.evaluateAnswer(request)
            
            if (response.isSuccessful && response.body() != null) {
                Result.success(response.body()!!)
            } else {
                val errorBody = response.errorBody()?.string() ?: "Unknown error"
                Log.e(TAG, "Evaluate failed: $errorBody")
                Result.failure(Exception("Evaluation failed: $errorBody"))
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error evaluating answer", e)
            Result.failure(e)
        }
    }
    
    // ==================== Feedback ====================
    
    /**
     * Submit interview feedback with scores
     * 
     * @param sessionId Session ID
     * @param userId User ID
     * @param overallScore Overall performance score (0-100)
     * @param technicalScore Technical knowledge score (0-100)
     * @param communicationScore Communication skills score (0-100)
     * @param confidenceScore Confidence level score (0-100)
     * @param textFeedback Optional text feedback
     */
    suspend fun submitFeedback(
        sessionId: String,
        userId: Long,
        overallScore: Double,
        technicalScore: Double,
        communicationScore: Double,
        confidenceScore: Double,
        textFeedback: String? = null
    ): Result<FeedbackResponse> {
        return try {
            Log.d(TAG, "Submitting feedback for session: $sessionId")
            val request = FeedbackRequest(
                session_id = sessionId,
                user_id = userId,
                overall_score = overallScore,
                technical_score = technicalScore,
                communication_score = communicationScore,
                confidence_score = confidenceScore,
                text_feedback = textFeedback
            )
            val response = aiService.submitFeedback(request)
            
            if (response.isSuccessful && response.body() != null) {
                val result = response.body()!!
                Log.d(TAG, "Feedback submitted: ${result.message}")
                Result.success(result)
            } else {
                val errorBody = response.errorBody()?.string() ?: "Unknown error"
                Log.e(TAG, "Submit feedback failed: $errorBody")
                Result.failure(Exception("Failed to submit feedback: $errorBody"))
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error submitting feedback", e)
            Result.failure(e)
        }
    }
    
    // ==================== Knowledge Base ====================
    
    /**
     * Seed knowledge base with domain-specific items
     */
    suspend fun seedKnowledgeBase(domain: String, items: List<String>): Result<SeedKBResponse> {
        return try {
            val request = SeedKBRequest(domain = domain, items = items)
            val response = aiService.seedKnowledgeBase(request)
            
            if (response.isSuccessful && response.body() != null) {
                Result.success(response.body()!!)
            } else {
                val errorBody = response.errorBody()?.string() ?: "Unknown error"
                Log.e(TAG, "Seed KB failed: $errorBody")
                Result.failure(Exception("Failed to seed KB: $errorBody"))
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error seeding KB", e)
            Result.failure(e)
        }
    }
    
    // ==================== User Scores ====================
    
    /**
     * Get user scores (technical, behavioral, CV analysis)
     */
    suspend fun getUserScores(userId: Long): Result<UserScoresResponse> {
        return try {
            Log.d(TAG, "Getting user scores for user: $userId")
            val response = withContext(Dispatchers.IO) {
                withTimeout(10000) {
                    aiService.getUserScores(userId)
                }
            }
            
            if (response.isSuccessful && response.body() != null) {
                val result = response.body()!!
                Log.d(TAG, "✅ User scores retrieved: Technical=${result.technical_score}, Behavioral=${result.behavioral_score}, CV=${result.cv_analysis_score}")
                Result.success(result)
            } else {
                val errorBody = try {
                    response.errorBody()?.string() ?: "Unknown error"
                } catch (e: Exception) {
                    "Error reading response: ${e.message}"
                }
                Log.e(TAG, "❌ Get user scores failed: HTTP ${response.code()} - $errorBody")
                Result.failure(Exception("Failed to get user scores (HTTP ${response.code()}): $errorBody"))
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error getting user scores", e)
            Result.failure(e)
        }
    }
    
    // ==================== Helper Methods ====================
    
    /**
     * Convert URI to File for upload
     * Returns Pair<File, MimeType> with proper file type detection
     */
    private fun uriToFile(uri: Uri): Pair<File, String> {
        // Check if URI is accessible
        val inputStream: InputStream? = try {
            context.contentResolver.openInputStream(uri)
        } catch (e: Exception) {
            Log.e(TAG, "Failed to open input stream from URI", e)
            throw Exception("Cannot access file. Please try selecting the file again.")
        }
        
        if (inputStream == null) {
            throw Exception("Cannot read file. Please check file permissions.")
        }
        
        // Get MIME type from content resolver
        var mimeType = try {
            context.contentResolver.getType(uri)
        } catch (e: Exception) {
            Log.w(TAG, "Failed to get MIME type", e)
            null
        }
        
        // Get file name and extension from URI
        var fileName = "cv_upload"
        var extension = "pdf"
        
        try {
            context.contentResolver.query(uri, null, null, null, null)?.use { cursor ->
                val nameIndex = cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME)
                if (cursor.moveToFirst() && nameIndex != -1) {
                    val originalName = cursor.getString(nameIndex)
                    if (originalName.isNotEmpty()) {
                        fileName = originalName.substringBeforeLast(".", originalName)
                        extension = originalName.substringAfterLast(".", "pdf").lowercase()
                    }
                }
            }
        } catch (e: Exception) {
            Log.w(TAG, "Failed to get file name from URI", e)
        }
        
        // Determine MIME type if not available from content resolver
        if (mimeType == null) {
            mimeType = when (extension) {
                "pdf" -> "application/pdf"
                "docx" -> "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
                "doc" -> "application/msword"
                "txt" -> "text/plain"
                else -> "application/pdf" // Default fallback
            }
        }
        
        // Ensure extension matches MIME type if MIME type was detected
        when {
            mimeType.contains("pdf") -> extension = "pdf"
            mimeType.contains("wordprocessingml") || mimeType.contains("docx") -> extension = "docx"
            mimeType.contains("msword") || mimeType.contains("doc") -> extension = "doc"
        }
        
        val file = File(context.cacheDir, "${fileName}_${System.currentTimeMillis()}.$extension")
        
        try {
            inputStream.use { input ->
                FileOutputStream(file).use { output ->
                    input.copyTo(output)
                }
            }
            
            // Verify file was created and has content
            if (!file.exists() || file.length() == 0L) {
                throw Exception("Failed to save file. Please try again.")
            }
            
            Log.d(TAG, "File created: ${file.name}, Size: ${file.length()} bytes, MIME type: $mimeType, Extension: $extension")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to copy file", e)
            // Clean up partial file
            file.delete()
            throw Exception("Failed to process file: ${e.message}")
        }
        
        return Pair(file, mimeType ?: "application/pdf")
    }
}
