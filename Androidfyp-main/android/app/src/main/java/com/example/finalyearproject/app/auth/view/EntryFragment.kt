package com.example.finalyearproject.app.auth.view

import android.animation.AnimatorSet
import android.animation.ObjectAnimator
import android.content.Intent
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.view.animation.AccelerateDecelerateInterpolator
import android.view.animation.AnimationUtils
import android.widget.Button
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import androidx.fragment.app.Fragment
import androidx.lifecycle.lifecycleScope
import com.example.datastore.getToken
import com.example.finalyearproject.R
import com.example.finalyearproject.app.homescreen.view.Home
import kotlinx.coroutines.launch

class EntryFragment : Fragment() {

    private lateinit var appLogo: ImageView
    private lateinit var welcomeTitle: TextView
    private lateinit var subTitle: TextView
    private lateinit var featuresContainer: LinearLayout
    private lateinit var getStartedButton: Button
    private lateinit var floatingDot1: View
    private lateinit var floatingDot2: View
    private lateinit var floatingDot3: View
    private lateinit var pulseIndicator: View

    private val floatingAnimators = mutableListOf<ObjectAnimator>()

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View? {
        return inflater.inflate(R.layout.fragment_entryfragment, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        // ✅ Check JWT Token First
        lifecycleScope.launch {
            val token = getToken(requireContext())

            if (!token.isNullOrEmpty()) {
                // ✅ Token exists — navigate directly to Home
                val intent = Intent(requireContext(), Home::class.java)
                startActivity(intent)
                requireActivity().finish()
                return@launch
            } else {
                // 🔹 No token — show entry animations
                initViews(view)
                setupAnimations()
                setupClickListeners()
            }
        }
    }

    private fun initViews(view: View) {
        appLogo = view.findViewById(R.id.appLogo)
        welcomeTitle = view.findViewById(R.id.welcomeTitle)
        subTitle = view.findViewById(R.id.subTitle)
        featuresContainer = view.findViewById(R.id.featuresContainer)
        getStartedButton = view.findViewById(R.id.getStartedButton)
        floatingDot1 = view.findViewById(R.id.floatingDot1)
        floatingDot2 = view.findViewById(R.id.floatingDot2)
        floatingDot3 = view.findViewById(R.id.floatingDot3)
        pulseIndicator = view.findViewById(R.id.pulseIndicator)
    }

    private fun setupAnimations() {
        Handler(Looper.getMainLooper()).postDelayed({
            val bounceAnim = AnimationUtils.loadAnimation(context, R.anim.bounce)
            appLogo.startAnimation(bounceAnim)
        }, 300)

        Handler(Looper.getMainLooper()).postDelayed({
            val fadeInAnim = AnimationUtils.loadAnimation(context, R.anim.fade_in_up)
            welcomeTitle.startAnimation(fadeInAnim)
        }, 800)

        Handler(Looper.getMainLooper()).postDelayed({
            val fadeInAnim = AnimationUtils.loadAnimation(context, R.anim.fade_in_up)
            subTitle.startAnimation(fadeInAnim)
        }, 1200)

        Handler(Looper.getMainLooper()).postDelayed({
            val slideInAnim = AnimationUtils.loadAnimation(context, R.anim.fade_in_up)
            featuresContainer.startAnimation(slideInAnim)
        }, 1600)

        Handler(Looper.getMainLooper()).postDelayed({
            startButtonPulseAnimation()
        }, 2000)

        startFloatingDotsAnimation()
        startLogoRotationAnimation()
    }

    private fun startButtonPulseAnimation() {
        val pulseAnimator = ObjectAnimator.ofFloat(pulseIndicator, "alpha", 0.3f, 0.0f).apply {
            duration = 2000
            repeatCount = ObjectAnimator.INFINITE
            repeatMode = ObjectAnimator.RESTART
            interpolator = AccelerateDecelerateInterpolator()
        }

        val scaleXAnimator = ObjectAnimator.ofFloat(pulseIndicator, "scaleX", 1.0f, 1.5f).apply {
            duration = 2000
            repeatCount = ObjectAnimator.INFINITE
            repeatMode = ObjectAnimator.RESTART
            interpolator = AccelerateDecelerateInterpolator()
        }

        val scaleYAnimator = ObjectAnimator.ofFloat(pulseIndicator, "scaleY", 1.0f, 1.5f).apply {
            duration = 2000
            repeatCount = ObjectAnimator.INFINITE
            repeatMode = ObjectAnimator.RESTART
            interpolator = AccelerateDecelerateInterpolator()
        }

        AnimatorSet().apply {
            playTogether(pulseAnimator, scaleXAnimator, scaleYAnimator)
            start()
        }
    }

    private fun startFloatingDotsAnimation() {
        val dot1YAnim = ObjectAnimator.ofFloat(floatingDot1, "translationY", 0f, -30f, 0f).apply {
            duration = 3000
            repeatCount = ObjectAnimator.INFINITE
            interpolator = AccelerateDecelerateInterpolator()
        }

        val dot1XAnim = ObjectAnimator.ofFloat(floatingDot1, "translationX", 0f, 15f, 0f).apply {
            duration = 4000
            repeatCount = ObjectAnimator.INFINITE
            interpolator = AccelerateDecelerateInterpolator()
        }

        val dot2YAnim = ObjectAnimator.ofFloat(floatingDot2, "translationY", 0f, 25f, 0f).apply {
            duration = 3500
            repeatCount = ObjectAnimator.INFINITE
            interpolator = AccelerateDecelerateInterpolator()
        }

        val dot2XAnim = ObjectAnimator.ofFloat(floatingDot2, "translationX", 0f, -20f, 0f).apply {
            duration = 2800
            repeatCount = ObjectAnimator.INFINITE
            interpolator = AccelerateDecelerateInterpolator()
        }

        val dot3YAnim = ObjectAnimator.ofFloat(floatingDot3, "translationY", 0f, -20f, 0f).apply {
            duration = 4200
            repeatCount = ObjectAnimator.INFINITE
            interpolator = AccelerateDecelerateInterpolator()
        }

        val dot3XAnim = ObjectAnimator.ofFloat(floatingDot3, "translationX", 0f, 25f, 0f).apply {
            duration = 3200
            repeatCount = ObjectAnimator.INFINITE
            interpolator = AccelerateDecelerateInterpolator()
        }

        floatingAnimators.addAll(listOf(dot1YAnim, dot1XAnim, dot2YAnim, dot2XAnim, dot3YAnim, dot3XAnim))
        floatingAnimators.forEach { it.start() }
    }

    private fun startLogoRotationAnimation() {
        ObjectAnimator.ofFloat(appLogo, "rotation", 0f, 5f, -5f, 0f).apply {
            duration = 6000
            repeatCount = ObjectAnimator.INFINITE
            interpolator = AccelerateDecelerateInterpolator()
            start()
        }
    }

    private fun setupClickListeners() {
        getStartedButton.setOnClickListener { v ->
            val scaleDownX = ObjectAnimator.ofFloat(v, "scaleX", 1.0f, 0.95f).apply { duration = 100 }
            val scaleUpX = ObjectAnimator.ofFloat(v, "scaleX", 0.95f, 1.0f).apply { duration = 100 }
            val scaleDownY = ObjectAnimator.ofFloat(v, "scaleY", 1.0f, 0.95f).apply { duration = 100 }
            val scaleUpY = ObjectAnimator.ofFloat(v, "scaleY", 0.95f, 1.0f).apply { duration = 100 }

            AnimatorSet().apply {
                play(scaleDownX).with(scaleDownY)
                play(scaleUpX).with(scaleUpY).after(scaleDownX)
                start()
            }

            Handler(Looper.getMainLooper()).postDelayed({
                navigateToNextScreen()
            }, 200)
        }
    }

    private fun navigateToNextScreen() {
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

    override fun onDestroyView() {
        super.onDestroyView()
        floatingAnimators.forEach { it.cancel() }
        floatingAnimators.clear()
    }
}
