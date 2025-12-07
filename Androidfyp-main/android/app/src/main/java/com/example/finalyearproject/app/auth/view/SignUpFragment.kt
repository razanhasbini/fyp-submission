package com.example.finalyearproject.app.auth.view

import android.nfc.Tag
import android.os.Bundle
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.AdapterView
import android.widget.Toast
import androidx.fragment.app.Fragment
import androidx.lifecycle.ViewModelProvider
import com.example.finalyearproject.R
import com.example.finalyearproject.app.auth.view.adapter.SignUpSpinnerAdapter
import com.example.finalyearproject.app.auth.viewmodels.SignUpViewModel
import com.example.finalyearproject.app.auth.viewmodels.SignUpViewModelFactory
import com.example.finalyearproject.app.repository.AuthRepository
import com.example.finalyearproject.databinding.FragmentSignUpBinding
import java.lang.Math.log

class SignUpFragment : Fragment() {

    private var _binding: FragmentSignUpBinding? = null
    private val binding get() = _binding!!

    private lateinit var viewModel: SignUpViewModel

    companion object {
        private const val TAG = "SignUpFragment"
    }

    override fun onCreateView(
        inflater: LayoutInflater, container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = FragmentSignUpBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        val authRepository = AuthRepository()
        val factory = SignUpViewModelFactory(authRepository)
        viewModel = ViewModelProvider(this, factory)[SignUpViewModel::class.java]

        setupSpinners()
        setupListeners()
    }

    private fun setupSpinners() {
        binding.jobPositionSpinner.adapter =
            SignUpSpinnerAdapter.getJobPositionAdapter(requireContext())

        binding.jobPositionSpinner.onItemSelectedListener =
            object : AdapterView.OnItemSelectedListener {
                override fun onItemSelected(
                    parent: AdapterView<*>?,
                    view: View?,
                    position: Int,
                    id: Long
                ) {
                    val selectedJob = parent?.getItemAtPosition(position).toString()
                    Log.d(TAG, "Job position selected: $selectedJob")
                    viewModel.onJobPositionSelected(selectedJob)

                    val specializationAdapter =
                        SignUpSpinnerAdapter.getSpecializationAdapter(requireContext(), selectedJob)
                    binding.spinnerSpecialization.adapter = specializationAdapter
                }

                override fun onNothingSelected(parent: AdapterView<*>?) {}
            }
    }

    private fun setupListeners() {
        binding.signInLink.setOnClickListener {
            requireActivity().supportFragmentManager.beginTransaction()
                .setCustomAnimations(R.anim.enter_from_right, R.anim.exit_to_left)
                .replace(R.id.fragment_container, SignInFragment())
                .addToBackStack(null)
                .commit()
        }

        binding.signUpBtn.setOnClickListener {
            performSignUp()
        }
    }

    private fun performSignUp() {
        val fullName = binding.fullNameField.text.toString()
        val email = binding.emailField.text.toString()
        val password = binding.passwordField.text.toString()
        val confirmPassword = binding.confirmPasswordField.text.toString()

        val validationError = validateSignUpInputs(fullName, email, password, confirmPassword)
        if (validationError != null) {
            Toast.makeText(requireContext(), validationError, Toast.LENGTH_LONG).show()
            return
        }

        showLoadingSpinner(true)
        binding.signUpBtn.isEnabled = false

        Log.d(TAG, "Sending verification code to: $email")

        viewModel.sendVerificationCode(email) { result ->
            requireActivity().runOnUiThread {
                showLoadingSpinner(false)
                binding.signUpBtn.isEnabled = true
                Log.w(TAG, "verifyResult = $result")
                when (result) {

                    is SignUpViewModel.SignUpResult.Success -> {
                        Log.d(TAG, "Verification code sent to $email")
                        Toast.makeText(requireContext(), "Verification code sent!", Toast.LENGTH_SHORT).show()
                        navigateToVerifyEmail(email, fullName, password)
                    }
                    is SignUpViewModel.SignUpResult.Error -> {
                        if (result.errorMessage.contains("sent", true)) {
                            Log.d(TAG, "Verification code sent to $email (error message fallback)")
                            Toast.makeText(requireContext(), "Verification code sent!", Toast.LENGTH_SHORT).show()
                            navigateToVerifyEmail(email, fullName, password)
                        } else {
                            Log.e(TAG, "Failed to send verification code: ${result.errorMessage}")
                            Toast.makeText(requireContext(), result.errorMessage, Toast.LENGTH_LONG).show()
                        }
                    }
                }
            }
        }
    }

    private fun validateSignUpInputs(fullName: String, email: String, password: String, confirmPassword: String): String? {
        return when {
            fullName.isBlank() -> "Please enter your full name"
            email.isBlank() -> "Please enter your email"
            !android.util.Patterns.EMAIL_ADDRESS.matcher(email).matches() -> "Invalid email"
            password.isBlank() -> "Please enter a password"
            password.length < 6 -> "Password must be at least 6 characters"
            confirmPassword != password -> "Passwords do not match"
            else -> null
        }
    }

    private fun navigateToVerifyEmail(email: String, fullName: String, password: String) {
        val jobPosition = binding.jobPositionSpinner.selectedItem?.toString() ?: ""
        val specialization = binding.spinnerSpecialization.selectedItem?.toString() ?: ""

        Log.d(TAG, "Navigating to VerifyEmailCodeFragment with email=$email, fullName=$fullName")

        val fragment = VerifyEmailCodeFragment.newInstance(
            email = email,
            fullName = fullName,
            password = password,
            jobPosition = jobPosition,
            specialization = specialization
        )

        requireActivity().supportFragmentManager.beginTransaction()
            .setCustomAnimations(R.anim.enter_from_right, R.anim.exit_to_left)
            .replace(R.id.fragment_container, fragment)
            .addToBackStack(null)
            .commit()
    }

    private fun showLoadingSpinner(show: Boolean) {
        binding.loadingSpinner?.visibility = if (show) View.VISIBLE else View.GONE
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
