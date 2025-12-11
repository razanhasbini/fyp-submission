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

class BehavioralAnalysisDetailFragment : Fragment() {
    
    private lateinit var tvBehavioralAnalysis: TextView
    private lateinit var aiRepository: AiRepository
    private lateinit var profileRepository: ProfileRepository
    
    companion object {
        private const val TAG = "BehavioralAnalysis"
    }
    
    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View? {
        return inflater.inflate(R.layout.fragment_behavioral_analysis_detail, container, false)
    }
    
    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        
        tvBehavioralAnalysis = view.findViewById(R.id.tvBehavioralAnalysis)
        aiRepository = AiRepository(requireContext())
        profileRepository = ProfileRepository(requireContext())
        
        configureUrl()
        loadBehavioralFeedback()
    }
    
    private fun configureUrl() {
        NetworkUtils.configureAiServiceUrl()
    }
    
    private fun loadBehavioralFeedback() {
        viewLifecycleOwner.lifecycleScope.launch {
            try {
                val profileResult = profileRepository.getMe()
                val userId = profileResult.getOrNull()?.id?.toLongOrNull()
                
                if (userId == null) {
                    tvBehavioralAnalysis.text = "Please log in to view your behavioral feedback."
                    return@launch
                }
                
                tvBehavioralAnalysis.text = "Loading behavioral feedback..."
                
                val feedbackResult = aiRepository.getBehavioralFeedback(userId)
                feedbackResult.fold(
                    onSuccess = { feedback ->
                        val feedbackText = feedback.feedback ?: "No detailed feedback available."
                        val communicationText = "Communication Score: ${feedback.communication_score.toInt()}%\n"
                        val confidenceText = "Confidence Score: ${feedback.confidence_score.toInt()}%\n\n"
                        val overallText = "Overall Score: ${feedback.overall_score.toInt()}%\n\n"
                        val behavioralScore = ((feedback.communication_score + feedback.confidence_score) / 2.0).toInt()
                        val scoreText = "Behavioral Score: $behavioralScore%\n\n"
                        val fullText = scoreText + communicationText + confidenceText + overallText + "Detailed Feedback:\n\n" + feedbackText
                        tvBehavioralAnalysis.text = fullText
                        Log.d(TAG, "✅ Behavioral feedback loaded successfully")
                    },
                    onFailure = { error ->
                        Log.e(TAG, "Failed to load behavioral feedback: ${error.message}", error)
                        tvBehavioralAnalysis.text = "Failed to load behavioral feedback. ${error.message}\n\nPlease complete an interview to get feedback."
                    }
                )
            } catch (e: Exception) {
                Log.e(TAG, "Error loading behavioral feedback", e)
                tvBehavioralAnalysis.text = "Error loading behavioral feedback: ${e.message}"
            }
        }
    }
}
