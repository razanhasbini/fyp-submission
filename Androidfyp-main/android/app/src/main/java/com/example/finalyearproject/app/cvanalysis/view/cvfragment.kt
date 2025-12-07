package com.example.finalyearproject.app.cvanalysis.view

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.provider.OpenableColumns
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.*
import androidx.activity.result.contract.ActivityResultContracts
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.lifecycleScope
import com.example.finalyearproject.R
import com.example.finalyearproject.app.cvanalysis.viewmodels.CvAnalysisViewModel
import com.example.finalyearproject.app.cvanalysis.viewmodels.CvAnalysisViewModelFactory
import com.example.finalyearproject.app.profile.viewmodels.ProfileViewModel
import com.example.finalyearproject.app.utils.NetworkUtils
import kotlinx.coroutines.launch
import kotlinx.coroutines.delay

class cvfragment : Fragment() {

    private val viewModel: CvAnalysisViewModel by viewModels {
        CvAnalysisViewModelFactory(requireActivity().application)
    }
    private var selectedFileUri: Uri? = null

    // Activity Result API to pick a document (CV file)
    private val pickCvFileLauncher =
        registerForActivityResult(ActivityResultContracts.OpenDocument()) { uri: Uri? ->
            if (uri != null) {
                handleSelectedFile(uri)
            }
        }
    
    companion object {
        private const val TAG = "CvFragment"
    }

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        return inflater.inflate(R.layout.fragment_cvfragment, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        NetworkUtils.configureAiServiceUrl()
        Log.d(TAG, "CV Fragment - Network configured. Base URL: ${com.example.finalyearproject.app.repository.network.AiRetrofitClient.getBaseUrl()}")

        val btnSelectCvFile = view.findViewById<Button>(R.id.btnSelectCvFile)
        val btnAnalyzeCv = view.findViewById<Button>(R.id.btnAnalyzeCv)
        val progressAnalyzing = view.findViewById<ProgressBar>(R.id.progressAnalyzing)
        val layoutAiResponse = view.findViewById<LinearLayout>(R.id.layoutAiResponse)
        val tvAiSummary = view.findViewById<TextView>(R.id.tvAiSummary)
        val tvAiSuggestions = view.findViewById<TextView>(R.id.tvAiSuggestions)

        // Initially, analyze button is disabled (no file yet)
        btnAnalyzeCv.isEnabled = false

        // Choose CV button
        btnSelectCvFile.setOnClickListener {
            // Only allow PDF and DOCX
            pickCvFileLauncher.launch(
                arrayOf(
                    "application/pdf",
                    "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
                )
            )
        }

        // Analyze button - now using ViewModel
        btnAnalyzeCv.setOnClickListener {
            val uri = selectedFileUri
            if (uri == null) {
                Toast.makeText(requireContext(), "Please choose a CV file first", Toast.LENGTH_SHORT)
                    .show()
                return@setOnClickListener
            }

            progressAnalyzing.visibility = View.VISIBLE
            layoutAiResponse.visibility = View.GONE
            btnAnalyzeCv.isEnabled = false

            viewModel.uploadAndAnalyzeCV(uri)
        }

        // Observe ViewModel state
        viewLifecycleOwner.lifecycleScope.launch {
            viewModel.state.collect { state ->
                progressAnalyzing.visibility = if (state.loading) View.VISIBLE else View.GONE
                btnAnalyzeCv.isEnabled = !state.loading

                if (state.uploadSuccess) {
                    layoutAiResponse.visibility = View.VISIBLE
                    if (state.analysisResult.isNullOrBlank()) {
                        tvAiSummary.text = "CV analysis completed successfully! Check your profile for details."
                    } else {
                        tvAiSummary.text = state.analysisResult
                    }
                    if (state.suggestions.isNullOrBlank()) {
                        tvAiSuggestions.text = "No specific suggestions provided."
                    } else {
                        tvAiSuggestions.text = state.suggestions
                    }
                    
                    ProfileViewModel.triggerRefresh = true
                    
                    viewLifecycleOwner.lifecycleScope.launch {
                        delay(2000)
                        val profileResult = com.example.finalyearproject.app.repository.ProfileRepository(requireContext()).getMe()
                        val userId = profileResult.getOrNull()?.id?.toLongOrNull()
                        if (userId != null) {
                            com.example.finalyearproject.app.repository.AiRepository(requireContext()).getUserScores(userId)
                        }
                    }
                }

                state.error?.let { error ->
                    Toast.makeText(requireContext(), "Error: $error", Toast.LENGTH_LONG).show()
                    Log.e(TAG, "CV upload error: $error")
                }
            }
        }
    }

    private fun handleSelectedFile(uri: Uri) {
        try {
            val tvSelectedFileName = view?.findViewById<TextView>(R.id.tvSelectedFileName)
            val btnAnalyzeCv = view?.findViewById<Button>(R.id.btnAnalyzeCv) ?: return

            val contentResolver = requireContext().contentResolver
            val mimeType = try {
                contentResolver.getType(uri)
            } catch (e: Exception) {
                Log.e(TAG, "Error getting MIME type", e)
                null
            }

        // Only allow PDF and DOCX
        val isAllowed = mimeType == "application/pdf" ||
                mimeType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

        if (!isAllowed) {
            Toast.makeText(
                requireContext(),
                "Only PDF or DOCX files are allowed.",
                Toast.LENGTH_LONG
            ).show()
            selectedFileUri = null
            btnAnalyzeCv.isEnabled = false
            tvSelectedFileName?.text = "Invalid file type"
            return
        }

            // Get file name (for UI)
            var displayName = "CV file"
            try {
                contentResolver.query(uri, null, null, null, null)?.use { cursor ->
                    val nameIndex = cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME)
                    if (cursor.moveToFirst() && nameIndex != -1) {
                        displayName = cursor.getString(nameIndex)
                    }
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error getting file name", e)
            }

            selectedFileUri = uri
            tvSelectedFileName?.text = displayName
            btnAnalyzeCv.isEnabled = true

            // Persist permission so you can read this URI later (e.g. in ViewModel / upload)
            val flags = Intent.FLAG_GRANT_READ_URI_PERMISSION or
                    Intent.FLAG_GRANT_PERSISTABLE_URI_PERMISSION
            try {
                contentResolver.takePersistableUriPermission(uri, flags)
            } catch (e: SecurityException) {
                Log.w(TAG, "Could not persist URI permission", e)
            } catch (e: Exception) {
                Log.e(TAG, "Error persisting URI permission", e)
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error handling selected file", e)
            Toast.makeText(requireContext(), "Error selecting file: ${e.message}", Toast.LENGTH_LONG).show()
        }
    }
}
