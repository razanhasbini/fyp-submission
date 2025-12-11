package com.example.finalyearproject.app.feedback.view

import android.os.Bundle
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.TextView
import androidx.fragment.app.Fragment
import androidx.lifecycle.lifecycleScope
import com.example.finalyearproject.R
import com.example.finalyearproject.app.repository.AiRepository
import com.example.finalyearproject.app.repository.ProfileRepository
import com.example.finalyearproject.app.utils.NetworkUtils
import kotlinx.coroutines.launch

class TechnicalAnalysisDetailFragment : Fragment() {
    
    private lateinit var tvTechnicalAnalysis: TextView
    private lateinit var aiRepository: AiRepository
    private lateinit var profileRepository: ProfileRepository
    
    companion object {
        private const val TAG = "TechnicalAnalysis"
    }
    
    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View? {
        return inflater.inflate(R.layout.fragment_technical_analysis_detail, container, false)
    }
    
    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        
        tvTechnicalAnalysis = view.findViewById(R.id.tvTechnicalAnalysis)
        aiRepository = AiRepository(requireContext())
        profileRepository = ProfileRepository(requireContext())
        
        configureUrl()
        loadTechnicalFeedback()
    }
    
    private fun configureUrl() {
        NetworkUtils.configureAiServiceUrl()
    }
    
    private fun loadTechnicalFeedback() {
        viewLifecycleOwner.lifecycleScope.launch {
            try {
                val profileResult = profileRepository.getMe()
                val userId = profileResult.getOrNull()?.id?.toLongOrNull()
                
                if (userId == null) {
                    tvTechnicalAnalysis.text = "Please log in to view your technical feedback."
                    return@launch
                }
                
                tvTechnicalAnalysis.text = "Loading technical feedback..."
                
                val feedbackResult = aiRepository.getTechnicalFeedback(userId)
                feedbackResult.fold(
                    onSuccess = { feedback ->
                        val feedbackText = feedback.feedback ?: "No detailed feedback available."
                        val scoreText = "Technical Score: ${feedback.technical_score.toInt()}%\n\n"
                        val overallText = "Overall Score: ${feedback.overall_score.toInt()}%\n\n"
                        val fullText = scoreText + overallText + "Detailed Feedback:\n\n" + feedbackText
                        tvTechnicalAnalysis.text = fullText
                        Log.d(TAG, "✅ Technical feedback loaded successfully")
                    },
                    onFailure = { error ->
                        Log.e(TAG, "Failed to load technical feedback: ${error.message}", error)
                        tvTechnicalAnalysis.text = "Failed to load technical feedback. ${error.message}\n\nPlease complete an interview to get feedback."
                    }
                )
            } catch (e: Exception) {
                Log.e(TAG, "Error loading technical feedback", e)
                tvTechnicalAnalysis.text = "Error loading technical feedback: ${e.message}"
            }
        }
    }
}
