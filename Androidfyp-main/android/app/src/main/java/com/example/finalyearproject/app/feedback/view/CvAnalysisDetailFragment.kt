package com.example.finalyearproject.app.feedback.view

import android.os.Build
import android.os.Bundle
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.TextView
import android.widget.Toast
import androidx.fragment.app.Fragment
import com.example.finalyearproject.R
import com.example.finalyearproject.app.repository.ProfileRepository
import com.example.finalyearproject.app.repository.network.AiRetrofitClient
import com.example.finalyearproject.app.repository.network.CvFeedbackResponse
import com.example.finalyearproject.app.utils.NetworkUtils
import kotlinx.coroutines.*
import retrofit2.Response

class CvAnalysisDetailFragment : Fragment() {
    
    private lateinit var tvCvAnalysis: TextView
    private lateinit var tvCvSuggestions: TextView
    private lateinit var tvCvGrade: TextView
    
    companion object {
        private const val TAG = "CvAnalysisDetail"
    }
    
    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View? {
        return inflater.inflate(R.layout.fragment_cv_analysis_detail, container, false)
    }
    
    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        
        tvCvAnalysis = view.findViewById(R.id.tvCvAnalysis)
        tvCvSuggestions = view.findViewById(R.id.tvCvSuggestions)
        tvCvGrade = view.findViewById(R.id.tvCvGrade)
        
        configureUrl()
        loadCvAnalysis()
    }
    
    private fun configureUrl() {
        NetworkUtils.configureAiServiceUrl()
    }
    
    private fun loadCvAnalysis() {
        Log.d(TAG, "Loading CV analysis")
        Log.d(TAG, "Backend URL: ${AiRetrofitClient.getBaseUrl()}")
        
        CoroutineScope(Dispatchers.Main).launch {
            try {
                val profileRepo = ProfileRepository(requireContext())
                val profileResult = withContext(Dispatchers.IO) {
                    profileRepo.getMe()
                }
                
                val userId = profileResult.getOrNull()?.id?.toLongOrNull()
                if (userId == null) {
                    tvCvAnalysis.text = "Please log in to view CV analysis"
                    return@launch
                }
                
                Log.d(TAG, "Loading CV analysis for user: $userId")
                
                val response = withContext(Dispatchers.IO) {
                    AiRetrofitClient.aiService.getCvFeedback(userId)
                }
                
                Log.d(TAG, "API call completed, status: ${response.code()}")
                
                if (response.isSuccessful && response.body() != null) {
                    val feedback = response.body()!!
                    Log.d(TAG, "CV feedback received - Grade: ${feedback.grade}, Response length: ${feedback.ai_response?.length}, Suggestions length: ${feedback.ai_suggestion?.length}")
                    
                    tvCvGrade.text = feedback.grade?.let { "Grade: $it%" } ?: "Grade: N/A"
                    tvCvAnalysis.text = feedback.ai_response ?: "No analysis available"
                    tvCvSuggestions.text = feedback.ai_suggestion ?: "No suggestions available"
                } else {
                    Log.e(TAG, "API call failed: ${response.code()}, ${response.message()}")
                    tvCvAnalysis.text = "No CV analysis found. Upload a CV to get feedback."
                    tvCvSuggestions.text = ""
                    tvCvGrade.text = ""
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error loading CV analysis", e)
                tvCvAnalysis.text = "Error loading CV analysis: ${e.message}"
                tvCvSuggestions.text = ""
                tvCvGrade.text = ""
                Toast.makeText(requireContext(), "Failed to load CV analysis", Toast.LENGTH_SHORT).show()
            }
        }
    }
}
