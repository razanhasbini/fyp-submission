package com.example.finalyearproject.app.auth.view

import android.os.Bundle
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ProgressBar
import android.widget.Toast
import androidx.fragment.app.Fragment
import androidx.lifecycle.ViewModelProvider
import com.example.finalyearproject.R
import com.example.finalyearproject.app.auth.viewmodels.SignUpViewModel
import com.example.finalyearproject.app.auth.viewmodels.SignUpViewModelFactory
import com.example.finalyearproject.app.repository.AuthRepository
import com.google.android.material.button.MaterialButton
import com.google.android.material.textfield.TextInputEditText

private const val ARG_EMAIL = "user_email"
private const val ARG_FULL_NAME = "full_name"
private const val ARG_PASSWORD = "password"
private const val ARG_JOB_POSITION = "job_position"
private const val ARG_SPECIALIZATION = "specialization"

class VerifyEmailCodeFragment : Fragment() {

    private lateinit var viewModel: SignUpViewModel
    private lateinit var codeField: TextInputEditText
    private lateinit var verifyButton: MaterialButton
    private var loadingSpinner: ProgressBar? = null

    private var email: String? = null
    private var fullName: String? = null
    private var password: String? = null
    private var jobPosition: String? = null
    private var specialization: String? = null

    companion object {
        private const val TAG = "VerifyEmailCodeFragment"

        fun newInstance(
            email: String,
            fullName: String,
            password: String,
            jobPosition: String,
            specialization: String
        ) = VerifyEmailCodeFragment().apply {
            arguments = Bundle().apply {
                putString(ARG_EMAIL, email)
                putString(ARG_FULL_NAME, fullName)
                putString(ARG_PASSWORD, password)
                putString(ARG_JOB_POSITION, jobPosition)
                putString(ARG_SPECIALIZATION, specialization)
            }
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        arguments?.let {
            email = it.getString(ARG_EMAIL)
            fullName = it.getString(ARG_FULL_NAME)
            password = it.getString(ARG_PASSWORD)
            jobPosition = it.getString(ARG_JOB_POSITION)
            specialization = it.getString(ARG_SPECIALIZATION)
        }
    }

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ) = inflater.inflate(R.layout.fragment_verifyemailcode, container, false)

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        viewModel = ViewModelProvider(
            this,
            SignUpViewModelFactory(AuthRepository())
        )[SignUpViewModel::class.java]

        codeField = view.findViewById(R.id.codeField)
        verifyButton = view.findViewById(R.id.verifyButton)
        loadingSpinner = view.findViewById(R.id.loadingSpinner)

        verifyButton.setOnClickListener {
            val code = codeField.text.toString().trim()
            if (code.isEmpty()) {
                Toast.makeText(requireContext(), "Enter verification code", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            verifyEmailAndCompleteRegistration(code)
        }

        view.findViewById<View>(R.id.resendCodeText)?.setOnClickListener {
            email?.let { userEmail ->
                Toast.makeText(requireContext(), "Resending code...", Toast.LENGTH_SHORT).show()
                viewModel.resendVerificationCode(userEmail) { result ->
                    requireActivity().runOnUiThread {
                        when (result) {
                            is SignUpViewModel.SignUpResult.Success ->
                                Toast.makeText(requireContext(), "Verification code sent!", Toast.LENGTH_SHORT).show()
                            is SignUpViewModel.SignUpResult.Error ->
                                Toast.makeText(requireContext(), result.errorMessage, Toast.LENGTH_LONG).show()
                        }
                    }
                }
            }
        }
    }

    private fun verifyEmailAndCompleteRegistration(code: String) {
        val emailVal = email ?: return
        val fullNameVal = fullName ?: return
        val passwordVal = password ?: return
        val jobVal = jobPosition ?: ""
        val specVal = specialization ?: ""

        showLoadingSpinner(true)
        verifyButton.isEnabled = false

        Log.d(TAG, "Verifying email code for: $emailVal")

        viewModel.verifyEmailCode(emailVal, code) { verifyResult ->
            requireActivity().runOnUiThread {
                if (verifyResult is SignUpViewModel.SignUpResult.Success) {
                    Log.d(TAG, "Email verified successfully")
                    Toast.makeText(requireContext(), "Email verified!", Toast.LENGTH_SHORT).show()

                    // Complete registration
                    viewModel.completeRegistration(
                        fullNameVal, emailVal, passwordVal, jobVal, specVal, "", specVal,
                        requireContext()
                    ) { result ->
                        requireActivity().runOnUiThread {
                            showLoadingSpinner(false)
                            verifyButton.isEnabled = true

                            when (result) {
                                is SignUpViewModel.SignUpResult.Success -> {
                                    Log.d(TAG, "Registration completed successfully")
                                    Toast.makeText(
                                        requireContext(),
                                        "Registration completed! Please sign in.",
                                        Toast.LENGTH_SHORT
                                    ).show()

                                    // Navigate to SignIn with prefilled credentials
                                    navigateToSignIn(emailVal, passwordVal)
                                }
                                is SignUpViewModel.SignUpResult.Error -> {
                                    Log.e(TAG, "Registration failed: ${result.errorMessage}")
                                    Toast.makeText(
                                        requireContext(),
                                        result.errorMessage,
                                        Toast.LENGTH_LONG
                                    ).show()
                                }
                            }
                        }
                    }
                } else {
                    showLoadingSpinner(false)
                    verifyButton.isEnabled = true
                    val errorMsg = (verifyResult as SignUpViewModel.SignUpResult.Error).errorMessage
                    Log.e(TAG, "Email verification failed: $errorMsg")
                    Toast.makeText(requireContext(), errorMsg, Toast.LENGTH_LONG).show()
                }
            }
        }
    }

    private fun showLoadingSpinner(show: Boolean) {
        loadingSpinner?.visibility = if (show) View.VISIBLE else View.GONE
    }

    private fun navigateToSignIn(email: String, password: String) {
        Log.d(TAG, "Navigating to SignIn with prefilled credentials for: $email")

        val signInFragment = SignInFragment.newInstance(
            email = email,
            password = password
        )

        requireActivity().supportFragmentManager.beginTransaction()
            .setCustomAnimations(
                R.anim.enter_from_right,
                R.anim.exit_to_left,
                R.anim.enter_from_left,
                R.anim.exit_to_right
            )
            .replace(R.id.fragment_container, signInFragment)
            .commit()
    }
}