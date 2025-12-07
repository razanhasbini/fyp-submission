package com.example.finalyearproject.app.practice.view

import android.Manifest
import android.content.pm.PackageManager
import android.os.Bundle
import android.speech.SpeechRecognizer
import android.speech.tts.TextToSpeech
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Button
import android.widget.ImageView
import android.widget.TextView
import android.widget.Toast
import androidx.activity.result.contract.ActivityResultContracts
import androidx.camera.core.CameraSelector
import androidx.camera.core.Preview
import androidx.camera.lifecycle.ProcessCameraProvider
import androidx.camera.view.PreviewView
import androidx.core.content.ContextCompat
import androidx.fragment.app.Fragment
import androidx.lifecycle.lifecycleScope
import com.example.finalyearproject.R
import com.example.finalyearproject.app.repository.AiRepository
import com.example.finalyearproject.app.repository.ProfileRepository
import com.example.finalyearproject.app.utils.NetworkUtils
import kotlinx.coroutines.*
import java.util.*
import java.util.concurrent.TimeUnit

class Practicefragment : Fragment(), TextToSpeech.OnInitListener {

    private var previewView: PreviewView? = null
    private var tvCameraPlaceholder: TextView? = null
    private var btnStartPractice: Button? = null
    private var btnEndSession: Button? = null
    private var tvCurrentQuestion: TextView? = null
    private var tvTimer: TextView? = null
    private var tvRecordingStatus: TextView? = null
    private var imgAvatar: ImageView? = null

    private var cameraProvider: ProcessCameraProvider? = null
    private var isInterviewActive = false
    private var sessionId: String? = null
    private var startTime: Long = 0
    private var timerJob: Job? = null
    private var questionCount = 0
    private var hasIntroduction = false

    private var tts: TextToSpeech? = null
    private var speechRecognizer: SpeechRecognizer? = null
    private var isListening = false
    private var isSpeaking = false
    private var isTtsReady = false

    private val aiRepository by lazy { AiRepository(requireContext()) }
    private val profileRepository by lazy { ProfileRepository(requireContext()) }

    private val cameraPermissionLauncher =
        registerForActivityResult(ActivityResultContracts.RequestPermission()) { isGranted ->
            if (isGranted) {
                startCameraPreview()
            } else {
                Toast.makeText(
                    requireContext(),
                    "Camera permission is required for practice.",
                    Toast.LENGTH_SHORT
                ).show()
            }
        }

    private val audioPermissionLauncher =
        registerForActivityResult(ActivityResultContracts.RequestPermission()) { isGranted ->
            if (isGranted) {
                initializeSpeechRecognition()
            } else {
                Toast.makeText(
                    requireContext(),
                    "Microphone permission is required for practice.",
                    Toast.LENGTH_SHORT
                ).show()
            }
        }

    companion object {
        private const val TAG = "PracticeFragment"
    }

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        return inflater.inflate(R.layout.fragment_practicefragment, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        NetworkUtils.configureAiServiceUrl()
        Log.d(TAG, "Network configured. Base URL: ${com.example.finalyearproject.app.repository.network.AiRetrofitClient.getBaseUrl()}")

        previewView = view.findViewById(R.id.previewView)
        tvCameraPlaceholder = view.findViewById(R.id.tvCameraPlaceholder)
        btnStartPractice = view.findViewById(R.id.btnStartPractice)
        btnEndSession = view.findViewById(R.id.btnEndSession)
        tvCurrentQuestion = view.findViewById(R.id.tvCurrentQuestion)
        tvTimer = view.findViewById(R.id.tvTimer)
        tvRecordingStatus = view.findViewById(R.id.tvRecordingStatus)
        imgAvatar = view.findViewById(R.id.imgAvatar)

        tts = TextToSpeech(requireContext(), this)

        btnStartPractice?.setOnClickListener {
            checkPermissionsAndStart()
        }

        btnEndSession?.setOnClickListener {
            endInterview()
        }

        btnEndSession?.isEnabled = false
    }

    private fun checkPermissionsAndStart() {
        val hasCamera = ContextCompat.checkSelfPermission(
            requireContext(),
            Manifest.permission.CAMERA
        ) == PackageManager.PERMISSION_GRANTED

        val hasAudio = ContextCompat.checkSelfPermission(
            requireContext(),
            Manifest.permission.RECORD_AUDIO
        ) == PackageManager.PERMISSION_GRANTED

        when {
            !hasCamera -> cameraPermissionLauncher.launch(Manifest.permission.CAMERA)
            !hasAudio -> audioPermissionLauncher.launch(Manifest.permission.RECORD_AUDIO)
            else -> startInterview()
        }
    }

    private fun startCameraPreview() {
        val cameraProviderFuture = ProcessCameraProvider.getInstance(requireContext())
        cameraProviderFuture.addListener({
            try {
                cameraProvider = cameraProviderFuture.get()
                bindCameraUseCases()
            } catch (e: Exception) {
                Log.e(TAG, "Error setting up camera", e)
                Toast.makeText(requireContext(), "Error setting up camera", Toast.LENGTH_SHORT).show()
            }
        }, ContextCompat.getMainExecutor(requireContext()))
    }

    private fun bindCameraUseCases() {
        val cameraProvider = cameraProvider ?: return
        val previewView = previewView ?: return

        val preview = Preview.Builder().build().also {
            it.setSurfaceProvider(previewView.surfaceProvider)
        }

        val cameraSelector = CameraSelector.DEFAULT_FRONT_CAMERA

        try {
            cameraProvider.unbindAll()
            cameraProvider.bindToLifecycle(
                viewLifecycleOwner,
                cameraSelector,
                preview
            )
            previewView.visibility = View.VISIBLE
            tvCameraPlaceholder?.visibility = View.GONE
        } catch (e: Exception) {
            Log.e(TAG, "Error binding camera", e)
        }
    }

    private fun initializeSpeechRecognition() {
        if (SpeechRecognizer.isRecognitionAvailable(requireContext())) {
            speechRecognizer = SpeechRecognizer.createSpeechRecognizer(requireContext())
            Log.d(TAG, "✅ Speech recognition initialized")
        } else {
            Log.e(TAG, "❌ Speech recognition not available")
            Toast.makeText(requireContext(), "Speech recognition not available", Toast.LENGTH_SHORT).show()
        }
    }

    override fun onInit(status: Int) {
        if (status == TextToSpeech.SUCCESS) {
            val result = tts?.setLanguage(Locale.US)
            if (result == TextToSpeech.LANG_MISSING_DATA || result == TextToSpeech.LANG_NOT_SUPPORTED) {
                Log.e(TAG, "❌ TTS language not supported")
                isTtsReady = false
            } else {
                Log.d(TAG, "✅ TTS initialized successfully")
                isTtsReady = true
            }
        } else {
            Log.e(TAG, "❌ TTS initialization failed with status: $status")
            isTtsReady = false
        }
    }

    private fun startInterview() {
        if (isInterviewActive) return

        viewLifecycleOwner.lifecycleScope.launch {
            try {
                val profileResult = profileRepository.getMe()
                val userId = profileResult.getOrNull()?.id?.toLongOrNull()
                if (userId == null) {
                    Toast.makeText(requireContext(), "Please log in to start interview", Toast.LENGTH_SHORT).show()
                    return@launch
                }

                // Check if CV is uploaded
                val scoresResult = aiRepository.getUserScores(userId)
                scoresResult.fold(
                    onSuccess = { scores ->
                        if (scores.cv_analysis_score == 0.0) {
                            Toast.makeText(
                                requireContext(),
                                "Please upload and analyze your CV before starting the interview.",
                                Toast.LENGTH_LONG
                            ).show()
                            return@launch
                        }
                        proceedWithInterview(userId)
                    },
                    onFailure = { error ->
                        Log.e(TAG, "Failed to check CV status", error)
                        Toast.makeText(
                            requireContext(),
                            "Could not verify CV status. Please ensure your CV is uploaded.",
                            Toast.LENGTH_LONG
                        ).show()
                    }
                )
            } catch (e: Exception) {
                Log.e(TAG, "Error checking CV status", e)
                Toast.makeText(requireContext(), "Error: ${e.message}", Toast.LENGTH_LONG).show()
            }
        }
    }
    
    private fun proceedWithInterview(userId: Long) {
        isInterviewActive = true
        btnStartPractice?.isEnabled = false
        btnEndSession?.isEnabled = true
        startTime = System.currentTimeMillis()
        startTimer()

        startCameraPreview()
        initializeSpeechRecognition()

        viewLifecycleOwner.lifecycleScope.launch {
            try {
                Log.d(TAG, "Attempting to start session for user: $userId")
                Log.d(TAG, "Backend URL: ${com.example.finalyearproject.app.repository.network.AiRetrofitClient.getBaseUrl()}")
                
                val sessionResult = aiRepository.startSession(userId, null)
                sessionResult.fold(
                    onSuccess = { response ->
                        sessionId = response.session_id
                        questionCount = 0
                        hasIntroduction = false
                        Log.d(TAG, "✅ Session started successfully: $sessionId")
                        speakIntroduction(userId)
                    },
                    onFailure = { error ->
                        Log.e(TAG, "❌ Failed to start session", error)
                        val errorMsg = "Failed to start session: ${error.message}\n\nBackend URL: ${com.example.finalyearproject.app.repository.network.AiRetrofitClient.getBaseUrl()}\n\nPlease ensure the backend server is running."
                        Toast.makeText(requireContext(), errorMsg, Toast.LENGTH_LONG).show()
                        resetInterview()
                    }
                )
            } catch (e: Exception) {
                Log.e(TAG, "Error starting interview", e)
                Toast.makeText(requireContext(), "Error: ${e.message}", Toast.LENGTH_LONG).show()
                resetInterview()
            }
        }
    }

    private fun speakIntroduction(userId: Long) {
        val introduction = "Hello! Thank you for taking the time today. I'll be conducting your interview. Let's begin with a brief introduction - could you tell me a bit about yourself?"
        tvCurrentQuestion?.text = introduction
        speakQuestion(introduction) {
            hasIntroduction = true
            viewLifecycleOwner.lifecycleScope.launch {
                delay(2000)
                getNextQuestion(userId)
            }
        }
    }
    
    private fun getNextQuestion(userId: Long) {
        viewLifecycleOwner.lifecycleScope.launch {
            try {
                questionCount++
                Log.d(TAG, "Getting next question for user: $userId (Question #$questionCount)")
                
                if (questionCount > 10) {
                    speakConclusion(userId)
                    return@launch
                }
                
                val questionResult = aiRepository.getNextQuestion(userId, "Software Engineering", "medium")
                questionResult.fold(
                    onSuccess = { response ->
                        Log.d(TAG, "✅ Question received: ${response.question}")
                        tvCurrentQuestion?.text = response.question
                        speakQuestion(response.question)
                    },
                    onFailure = { error ->
                        Log.e(TAG, "❌ Failed to get question", error)
                        val errorMsg = "Failed to get question: ${error.message}\n\nBackend URL: ${com.example.finalyearproject.app.repository.network.AiRetrofitClient.getBaseUrl()}"
                        Toast.makeText(requireContext(), errorMsg, Toast.LENGTH_LONG).show()
                        tvRecordingStatus?.text = "Error getting question"
                    }
                )
            } catch (e: Exception) {
                Log.e(TAG, "❌ Exception getting question", e)
                Toast.makeText(requireContext(), "Error: ${e.message}", Toast.LENGTH_LONG).show()
            }
        }
    }
    
    private fun speakConclusion(userId: Long) {
        val conclusion = "Thank you for your time today. We'll be in touch soon. Do you have any questions for us?"
        tvCurrentQuestion?.text = conclusion
        speakQuestion(conclusion) {
            viewLifecycleOwner.lifecycleScope.launch {
                delay(3000)
                endInterview()
            }
        }
    }

    private fun speakQuestion(question: String, onComplete: (() -> Unit)? = null) {
        if (isSpeaking) {
            Log.w(TAG, "Already speaking, skipping")
            return
        }

        if (!isTtsReady || tts == null) {
            Log.e(TAG, "❌ TTS not ready, displaying question as text")
            tvRecordingStatus?.text = "Question displayed (TTS not available)"
            tvRecordingStatus?.setTextColor(ContextCompat.getColor(requireContext(), android.R.color.holo_orange_dark))
            viewLifecycleOwner.lifecycleScope.launch {
                delay(2000)
                onComplete?.invoke() ?: startListening()
            }
            return
        }

        isSpeaking = true
        tvRecordingStatus?.text = "AI is speaking..."
        tvRecordingStatus?.setTextColor(ContextCompat.getColor(requireContext(), android.R.color.holo_blue_dark))
        Log.d(TAG, "🎤 Starting to speak question: $question")

        tts?.setOnUtteranceProgressListener(object : android.speech.tts.UtteranceProgressListener() {
            override fun onStart(utteranceId: String?) {
                Log.d(TAG, "✅ TTS started speaking question")
            }

            override fun onDone(utteranceId: String?) {
                Log.d(TAG, "✅ TTS finished speaking")
                isSpeaking = false
                viewLifecycleOwner.lifecycleScope.launch {
                    delay(2000)
                    onComplete?.invoke() ?: startListening()
                }
            }

            override fun onError(utteranceId: String?) {
                Log.e(TAG, "❌ TTS error")
                isSpeaking = false
                viewLifecycleOwner.lifecycleScope.launch {
                    delay(2000)
                    onComplete?.invoke() ?: startListening()
                }
            }
        })

        val result = tts?.speak(question, TextToSpeech.QUEUE_FLUSH, null, "question")
        if (result == TextToSpeech.ERROR) {
            Log.e(TAG, "❌ TTS speak returned ERROR")
            isSpeaking = false
            onComplete?.invoke() ?: startListening()
        } else {
            Log.d(TAG, "✅ TTS speak initiated successfully")
        }
    }

    private fun startListening() {
        if (isListening || isSpeaking) return

        isListening = true
        tvRecordingStatus?.text = "Listening..."
        tvRecordingStatus?.setTextColor(ContextCompat.getColor(requireContext(), android.R.color.holo_green_dark))

        val speechRecognizer = speechRecognizer ?: run {
            isListening = false
            return
        }

        val intent = android.content.Intent(android.speech.RecognizerIntent.ACTION_RECOGNIZE_SPEECH).apply {
            putExtra(android.speech.RecognizerIntent.EXTRA_LANGUAGE_MODEL, android.speech.RecognizerIntent.LANGUAGE_MODEL_FREE_FORM)
            putExtra(android.speech.RecognizerIntent.EXTRA_LANGUAGE, Locale.getDefault())
        }

        speechRecognizer.setRecognitionListener(object : android.speech.RecognitionListener {
            override fun onReadyForSpeech(params: Bundle?) {
                Log.d(TAG, "Ready for speech")
            }

            override fun onBeginningOfSpeech() {
                Log.d(TAG, "Beginning of speech")
            }

            override fun onRmsChanged(rmsdB: Float) {}

            override fun onBufferReceived(buffer: ByteArray?) {}

            override fun onEndOfSpeech() {
                Log.d(TAG, "End of speech")
            }

            override fun onError(error: Int) {
                Log.e(TAG, "Speech recognition error: $error")
                isListening = false
                tvRecordingStatus?.text = "Error listening"
                tvRecordingStatus?.setTextColor(ContextCompat.getColor(requireContext(), android.R.color.holo_red_dark))
            }

            override fun onResults(results: Bundle?) {
                val matches = results?.getStringArrayList(android.speech.SpeechRecognizer.RESULTS_RECOGNITION)
                val answer = matches?.get(0) ?: ""
                Log.d(TAG, "Recognized: $answer")
                isListening = false
                processAnswer(answer)
            }

            override fun onPartialResults(partialResults: Bundle?) {}

            override fun onEvent(eventType: Int, params: Bundle?) {}
        })

        try {
            speechRecognizer.startListening(intent)
        } catch (e: Exception) {
            Log.e(TAG, "Error starting speech recognition", e)
            isListening = false
        }
    }

    private fun processAnswer(answer: String) {
        if (answer.isBlank()) {
            startListening()
            return
        }

        tvRecordingStatus?.text = "Processing answer..."
        tvRecordingStatus?.setTextColor(ContextCompat.getColor(requireContext(), android.R.color.holo_orange_dark))

        viewLifecycleOwner.lifecycleScope.launch {
            try {
                val profileResult = profileRepository.getMe()
                val userId = profileResult.getOrNull()?.id?.toLongOrNull() ?: return@launch
                val currentSessionId = sessionId ?: return@launch
                val currentQuestion = tvCurrentQuestion?.text?.toString() ?: return@launch

                val evaluateResult = aiRepository.evaluateAnswer(
                    currentSessionId,
                    userId,
                    "Software Engineering",
                    currentQuestion,
                    answer
                )

                evaluateResult.fold(
                    onSuccess = { response ->
                        Log.d(TAG, "Answer evaluated: ${response.score}")
                        tvRecordingStatus?.text = "Answer evaluated. Preparing next question..."
                        tvRecordingStatus?.setTextColor(ContextCompat.getColor(requireContext(), android.R.color.holo_green_dark))
                        delay(3000)
                        getNextQuestion(userId)
                    },
                    onFailure = { error ->
                        Log.e(TAG, "Failed to evaluate answer", error)
                        Toast.makeText(requireContext(), "Error evaluating answer", Toast.LENGTH_SHORT).show()
                        delay(2000)
                        getNextQuestion(userId)
                    }
                )
            } catch (e: Exception) {
                Log.e(TAG, "Error processing answer", e)
                delay(1000)
                startListening()
            }
        }
    }

    private fun startTimer() {
        timerJob?.cancel()
        timerJob = viewLifecycleOwner.lifecycleScope.launch {
            while (isInterviewActive) {
                val elapsed = System.currentTimeMillis() - startTime
                val minutes = TimeUnit.MILLISECONDS.toMinutes(elapsed)
                val seconds = TimeUnit.MILLISECONDS.toSeconds(elapsed) % 60
                tvTimer?.text = String.format("%02d:%02d", minutes, seconds)
                delay(1000)
            }
        }
    }

    private fun endInterview() {
        if (!isInterviewActive) return

        isInterviewActive = false
        timerJob?.cancel()
        btnStartPractice?.isEnabled = true
        btnEndSession?.isEnabled = false

        tts?.stop()
        speechRecognizer?.stopListening()
        speechRecognizer?.cancel()
        isListening = false
        isSpeaking = false

        viewLifecycleOwner.lifecycleScope.launch {
            try {
                val profileResult = profileRepository.getMe()
                val userId = profileResult.getOrNull()?.id?.toLongOrNull()
                val currentSessionId = sessionId

                if (userId != null && currentSessionId != null) {
                    val endResult = aiRepository.endSession(currentSessionId, userId)
                    endResult.fold(
                        onSuccess = { response ->
                            Toast.makeText(requireContext(), "Session ended. Grade: ${response.grade}", Toast.LENGTH_LONG).show()
                            Log.d(TAG, "Session ended with grade: ${response.grade}")
                        },
                        onFailure = { error ->
                            Log.e(TAG, "Failed to end session", error)
                        }
                    )
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error ending session", e)
            }
        }

        resetInterview()
    }

    private fun resetInterview() {
        isInterviewActive = false
        questionCount = 0
        hasIntroduction = false
        sessionId = null
        startTime = 0
        timerJob?.cancel()
        tvTimer?.text = "00:00"
        tvCurrentQuestion?.text = "Your question will appear here."
        tvRecordingStatus?.text = "Not recording"
        tvRecordingStatus?.setTextColor(ContextCompat.getColor(requireContext(), android.R.color.holo_red_dark))
        btnStartPractice?.isEnabled = true
        btnEndSession?.isEnabled = false
        previewView?.visibility = View.GONE
        tvCameraPlaceholder?.visibility = View.VISIBLE
    }

    override fun onDestroyView() {
        super.onDestroyView()
        endInterview()
        cameraProvider?.unbindAll()
        tts?.stop()
        tts?.shutdown()
        speechRecognizer?.destroy()
        previewView = null
        tvCameraPlaceholder = null
        btnStartPractice = null
        btnEndSession = null
        tvCurrentQuestion = null
        tvTimer = null
        tvRecordingStatus = null
        imgAvatar = null
    }
}
