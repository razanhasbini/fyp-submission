// app/auth/view/ChangePass.kt
package com.example.finalyearproject.app.auth.view

import android.os.Bundle
import androidx.fragment.app.Fragment
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Toast
import androidx.fragment.app.viewModels
import com.example.finalyearproject.R
import com.example.finalyearproject.app.auth.viewmodels.ChangePassViewModel
import com.example.finalyearproject.databinding.FragmentChangePasswordBinding

class ChangePass : Fragment() {

    private var _binding: FragmentChangePasswordBinding? = null
    private val binding get() = _binding!!
    private val viewModel: ChangePassViewModel by viewModels()
    private var email: String? = null

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        arguments?.let {
            email = it.getString("email")
        }
    }

    override fun onCreateView(
        inflater: LayoutInflater, container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = FragmentChangePasswordBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        setupObservers()
        setupListeners()
    }

    private fun setupObservers() {
        viewModel.resetPasswordState.observe(viewLifecycleOwner) { result ->
            hideLoading()

            result.onSuccess { message ->
                Toast.makeText(requireContext(), message, Toast.LENGTH_LONG).show()
                navigateToSignIn()
            }
            result.onFailure { error ->
                Toast.makeText(requireContext(), error.message ?: "An error occurred", Toast.LENGTH_SHORT).show()
            }
        }
    }

    private fun setupListeners() {
        binding.resetPasswordBtn.setOnClickListener {
            val code = binding.codeField.text.toString().trim()
            val newPassword = binding.newPasswordField.text.toString().trim()
            val confirmPassword = binding.confirmPasswordField.text.toString().trim()

            // Validation
            if (code.isEmpty()) {
                binding.codeLayout.error = "Code is required"
                return@setOnClickListener
            }

            if (code.length != 6) {
                binding.codeLayout.error = "Code must be 6 digits"
                return@setOnClickListener
            }

            if (newPassword.isEmpty()) {
                binding.newPasswordLayout.error = "Password is required"
                return@setOnClickListener
            }

            if (newPassword.length < 8) {
                binding.newPasswordLayout.error = "Password must be at least 8 characters"
                return@setOnClickListener
            }

            if (confirmPassword.isEmpty()) {
                binding.confirmPasswordLayout.error = "Please confirm your password"
                return@setOnClickListener
            }

            if (newPassword != confirmPassword) {
                binding.confirmPasswordLayout.error = "Passwords do not match"
                return@setOnClickListener
            }

            if (email == null) {
                Toast.makeText(requireContext(), "Email not found", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }

            // Clear errors
            binding.codeLayout.error = null
            binding.newPasswordLayout.error = null
            binding.confirmPasswordLayout.error = null

            // Show loading and call API
            showLoading()
            viewModel.resetPassword(email!!, code, newPassword)
        }

        binding.signInLink.setOnClickListener {
            navigateToSignIn()
        }
    }

    private fun showLoading() {
        binding.resetPasswordBtn.isEnabled = false
        binding.resetPasswordBtn.text = "Resetting..."
        binding.codeField.isEnabled = false
        binding.newPasswordField.isEnabled = false
        binding.confirmPasswordField.isEnabled = false
    }

    private fun hideLoading() {
        binding.resetPasswordBtn.isEnabled = true
        binding.resetPasswordBtn.text = "Reset Password"
        binding.codeField.isEnabled = true
        binding.newPasswordField.isEnabled = true
        binding.confirmPasswordField.isEnabled = true
    }

    private fun navigateToSignIn() {
        requireActivity().supportFragmentManager.beginTransaction()
            .setCustomAnimations(
                R.anim.enter_from_right,
                R.anim.exit_to_left,
                R.anim.enter_from_left,
                R.anim.exit_to_right
            )
            .replace(R.id.fragment_container, SignInFragment())
            .commit()
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}