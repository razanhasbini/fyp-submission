package com.example.finalyearproject.app.auth.view

import android.content.Intent
import android.os.Bundle
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Toast
import androidx.fragment.app.Fragment
import androidx.lifecycle.lifecycleScope
import com.example.datastore.getToken
import com.example.datastore.saveToken
import com.example.finalyearproject.R
import com.example.finalyearproject.app.auth.viewmodels.SignInResult
import com.example.finalyearproject.app.auth.viewmodels.SignInViewModel
import com.example.finalyearproject.databinding.FragmentSignInBinding
import com.example.finalyearproject.app.homescreen.view.Home
import kotlinx.coroutines.launch

private const val ARG_EMAIL = "email"
private const val ARG_PASSWORD = "password"

class SignInFragment : Fragment() {

    private var _binding: FragmentSignInBinding? = null
    private val binding get() = _binding!!

    private var prefilledEmail: String? = null
    private var prefilledPassword: String? = null

    private lateinit var viewModel: SignInViewModel

    companion object {
        fun newInstance(email: String? = null, password: String? = null) = SignInFragment().apply {
            arguments = Bundle().apply {
                putString(ARG_EMAIL, email)
                putString(ARG_PASSWORD, password)
            }
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        arguments?.let {
            prefilledEmail = it.getString(ARG_EMAIL)
            prefilledPassword = it.getString(ARG_PASSWORD)
        }
    }

    override fun onCreateView(
        inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?
    ) = FragmentSignInBinding.inflate(inflater, container, false).also { _binding = it }.root

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        viewModel = SignInViewModel(requireContext())

        prefilledEmail?.let { binding.emailField.setText(it) }
        prefilledPassword?.let { binding.passwordField.setText(it) }

        setupListeners()
    }

    private fun setupListeners() {

        // Navigate to Sign Up
        binding.signUpLink.setOnClickListener {
            requireActivity().supportFragmentManager.beginTransaction()
                .setCustomAnimations(
                    R.anim.enter_from_right, R.anim.exit_to_left,
                    R.anim.enter_from_left, R.anim.exit_to_right
                )
                .replace(R.id.fragment_container, SignUpFragment())
                .addToBackStack(null)
                .commit()
        }

        // Navigate to Forgot Password
        binding.forgotPassword.setOnClickListener {
            requireActivity().supportFragmentManager.beginTransaction()
                .setCustomAnimations(
                    R.anim.enter_from_right, R.anim.exit_to_left,
                    R.anim.enter_from_left, R.anim.exit_to_right
                )
                .replace(R.id.fragment_container, ForgetPasswordFragment())
                .addToBackStack(null)
                .commit()
        }

        // Sign In
        binding.signInBtn.setOnClickListener {
            val email = binding.emailField.text.toString().trim()
            val password = binding.passwordField.text.toString().trim()

            if (email.isEmpty() || password.isEmpty()) {
                Toast.makeText(requireContext(), "Please enter email and password", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }

            // Disable button and show spinner
            binding.signInBtn.isEnabled = false
            binding.signInProgressBar?.visibility = View.VISIBLE

            viewModel.signIn(email, password) { result ->
                requireActivity().runOnUiThread {
                    // Enable button and hide spinner
                    binding.signInBtn.isEnabled = true
                    binding.signInProgressBar?.visibility = View.GONE

                    when (result) {
                        is SignInResult.Success -> {
                            val token = result.response.token
                            Log.d("TOKEN_TEST", "Token returned from API: $token")

                            // Save token securely
                            lifecycleScope.launch {
                                try {
                                    saveToken(requireContext(), token)
                                    Log.d("TOKEN_TEST", "Token saved successfully")
                                } catch (e: Exception) {
                                    Log.e("TOKEN_TEST", "Error saving token", e)
                                }
                            }

                            Toast.makeText(requireContext(), "Sign in successful!", Toast.LENGTH_SHORT).show()
                            val intent = Intent(requireContext(), Home::class.java)
                            startActivity(intent)
                            requireActivity().finish()
                        }
                        is SignInResult.Error -> {
                            val errorMessage = result.errorMessage.lowercase()
                            if ("invalid" in errorMessage || "incorrect" in errorMessage || "wrong" in errorMessage) {
                                Toast.makeText(requireContext(), "Email or password is incorrect", Toast.LENGTH_LONG).show()
                            } else {
                                Toast.makeText(requireContext(), "Error: ${result.errorMessage}", Toast.LENGTH_LONG).show()
                            }
                        }
                    }
                }
            }
        }



    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
