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

class BehavioralAnalysisDetailFragment : Fragment() {
    
    private lateinit var tvBehavioralAnalysis: TextView
    
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
        
        configureUrl()
        
        val behavioralFeedback = arguments?.getString("behavioral_feedback")
        if (behavioralFeedback != null && behavioralFeedback.isNotEmpty()) {
            tvBehavioralAnalysis.text = behavioralFeedback
        } else {
            tvBehavioralAnalysis.text = "No behavioral analysis available. Complete an interview to get feedback."
        }
    }
    
    private fun configureUrl() {
        NetworkUtils.configureAiServiceUrl()
    }
}
