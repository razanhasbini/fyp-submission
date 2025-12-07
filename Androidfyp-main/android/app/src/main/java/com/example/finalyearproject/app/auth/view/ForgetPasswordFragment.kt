// app/auth/view/ForgetPasswordFragment.kt
package com.example.finalyearproject.app.auth.view

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Toast
import androidx.core.os.bundleOf
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import com.example.finalyearproject.R
import com.example.finalyearproject.app.auth.viewmodels.ForgetPasswordViewModel
import com.example.finalyearproject.databinding.FragmentForgetPasswordBinding

class ForgetPasswordFragment : Fragment() {

    private var _binding: FragmentForgetPasswordBinding? = null
    private val binding get() = _binding!!
    private val viewModel: ForgetPasswordViewModel by viewModels()

    override fun onCreateView(
        inflater: LayoutInflater, container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = FragmentForgetPasswordBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        setupObservers()
        setupListeners()
    }

    private fun setupObservers() {
        viewModel.sendResetCodeState.observe(viewLifecycleOwner) { result ->
            hideLoading()

            result.onSuccess { message ->
                Toast.makeText(requireContext(), message, Toast.LENGTH_SHORT).show()

                // Navigate to ChangePass fragment with email
                val email = binding.emailField.text.toString()
                navigateToChangePassword(email)
            }
            result.onFailure { error ->
                Toast.makeText(requireContext(), error.message ?: "An error occurred", Toast.LENGTH_SHORT).show()
            }
        }
    }

    private fun setupListeners() {
        binding.resetPasswordBtn.setOnClickListener {
            val email = binding.emailField.text.toString().trim()

            if (email.isEmpty()) {
                binding.emailLayout.error = "Email is required"
                return@setOnClickListener
            }

            if (!android.util.Patterns.EMAIL_ADDRESS.matcher(email).matches()) {
                binding.emailLayout.error = "Invalid email format"
                return@setOnClickListener
            }

            binding.emailLayout.error = null
            showLoading()
            viewModel.sendPasswordResetCode(email)
        }

        binding.signInLink.setOnClickListener {
            requireActivity().supportFragmentManager.beginTransaction()
                .setCustomAnimations(
                    R.anim.enter_from_right,
                    R.anim.exit_to_left,
                    R.anim.enter_from_left,
                    R.anim.exit_to_right
                )
                .replace(R.id.fragment_container, SignInFragment())
                .addToBackStack(null)
                .commit()
        }
    }

    private fun showLoading() {
        binding.resetPasswordBtn.isEnabled = false
        binding.resetPasswordBtn.text = "Sending..."
        binding.emailField.isEnabled = false
    }

    private fun hideLoading() {
        binding.resetPasswordBtn.isEnabled = true
        binding.resetPasswordBtn.text = "Reset Password"
        binding.emailField.isEnabled = true
    }

    private fun navigateToChangePassword(email: String) {
        val bundle = bundleOf("email" to email)
        val changePassFragment = ChangePass().apply {
            arguments = bundle
        }

        requireActivity().supportFragmentManager.beginTransaction()
            .setCustomAnimations(
                R.anim.enter_from_right,
                R.anim.exit_to_left,
                R.anim.enter_from_left,
                R.anim.exit_to_right
            )
            .replace(R.id.fragment_container, changePassFragment)
            .addToBackStack(null)
            .commit()
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}