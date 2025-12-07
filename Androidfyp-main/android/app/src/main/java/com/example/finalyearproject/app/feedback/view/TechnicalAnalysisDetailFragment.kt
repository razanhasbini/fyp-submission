package com.example.finalyearproject.app.feedback.view

import android.os.Build
import android.os.Bundle
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.TextView
import androidx.fragment.app.Fragment
import com.example.finalyearproject.R
import com.example.finalyearproject.app.repository.network.AiRetrofitClient
import com.example.finalyearproject.app.utils.NetworkUtils

class TechnicalAnalysisDetailFragment : Fragment() {
    
    private lateinit var tvTechnicalAnalysis: TextView
    
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
        
        configureUrl()
        
        val technicalFeedback = arguments?.getString("technical_feedback")
        if (technicalFeedback != null && technicalFeedback.isNotEmpty()) {
            tvTechnicalAnalysis.text = technicalFeedback
        } else {
            tvTechnicalAnalysis.text = "No technical analysis available. Complete an interview to get feedback."
        }
    }
    
    private fun configureUrl() {
        NetworkUtils.configureAiServiceUrl()
    }
}
