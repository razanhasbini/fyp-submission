package com.example.finalyearproject.app.profile.view

import android.content.Intent
import com.example.datastore.saveToken
import android.graphics.Typeface
import android.os.Bundle
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.View.GONE
import android.view.View.VISIBLE
import android.view.ViewGroup
import android.widget.LinearLayout
import android.widget.ProgressBar
import android.widget.TextView
import android.widget.Toast
import androidx.core.content.ContextCompat
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.lifecycle.lifecycleScope
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import coil.load
import coil.request.CachePolicy
import com.example.finalyearproject.R
import com.example.finalyearproject.app.auth.view.AuthActivity
import com.example.finalyearproject.app.profile.viewmodels.ProfileViewModel
import com.example.finalyearproject.app.profile.viewmodels.MyPostsViewModel
import com.example.finalyearproject.app.profile_edit.view.ProfileEditFragment
import com.example.finalyearproject.app.feedback.view.CvAnalysisDetailFragment
import com.example.finalyearproject.app.feedback.view.TechnicalAnalysisDetailFragment
import com.example.finalyearproject.app.feedback.view.BehavioralAnalysisDetailFragment
import com.example.finalyearproject.app.repository.AiRepository
import com.example.finalyearproject.app.repository.ProfileRepository
import de.hdodenhof.circleimageview.CircleImageView
import kotlinx.coroutines.launch
import androidx.cardview.widget.CardView

class ProfileFragment : Fragment() {

    private val viewModel: ProfileViewModel by viewModels()
    private val postsViewModel: MyPostsViewModel by viewModels()

    // Containers
    private lateinit var loadingSpinner: ProgressBar
    private lateinit var scroll: View

    // Tabs
    private lateinit var tabRecentSessions: TextView
    private lateinit var tabYourPosts: TextView
    private lateinit var recentSessionsSection: LinearLayout
    private lateinit var yourPostsSection: LinearLayout
    private lateinit var postsRecyclerView: RecyclerView
    private lateinit var emptyPostsText: TextView

    // Profile info views
    private lateinit var userName: TextView
    private lateinit var userRole: TextView
    private lateinit var followersCount: TextView
    private lateinit var followingCount: TextView
    private lateinit var profileImage: CircleImageView
    private lateinit var followersContainer: View
    private lateinit var followingContainer: View

    // Score views
    private lateinit var tvTechnicalScore: TextView
    private lateinit var tvBehavioralScore: TextView
    private lateinit var tvCvAnalysisScore: TextView
    private lateinit var tvCvAnalysisScoreHeader: TextView
    private lateinit var cardTechnicalInterview: CardView
    private lateinit var cardBehavioralInterview: CardView
    private lateinit var cardCvAnalysis: CardView

    private lateinit var postsAdapter: MyPostsAdapter
    private lateinit var aiRepository: AiRepository
    private lateinit var profileRepository: ProfileRepository

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        return inflater.inflate(R.layout.fragment_profilefragment, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        aiRepository = AiRepository(requireContext())
        profileRepository = ProfileRepository(requireContext())
        
        initViews(view)
        setupRecyclerView()
        setupToggleListeners()
        setupScoreCardListeners()

        // ✅ Set default tab to Recent Sessions
        selectRecentSessions()
        
        // Load user scores
        loadUserScores()

        // Edit profile click
        val editIcon: View = view.findViewById(R.id.editIcon)
        editIcon.setOnClickListener { navigateToProfileEdit() }

        val logoutButton: View = view.findViewById(R.id.logoutButton)
        logoutButton.setOnClickListener {
            lifecycleScope.launch {
                try {
                    saveToken(requireContext(), "")
                    Toast.makeText(requireContext(), "Logged out successfully", Toast.LENGTH_SHORT).show()
                    val intent = Intent(requireContext(), AuthActivity::class.java)
                    intent.flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
                    startActivity(intent)
                } catch (e: Exception) {
                    Log.e("ProfileFragment", "Error during logout: ${e.message}", e)
                    Toast.makeText(requireContext(), "Logout failed", Toast.LENGTH_SHORT).show()
                }
            }
        }

        // Start loading profile data
        showLoading(true)
        viewModel.loadProfile()

        // Observe Profile ViewModel
        viewLifecycleOwner.lifecycleScope.launch {
            viewModel.state.collect { state ->
                if (state.loading) {
                    showLoading(true)
                    return@collect
                }

                showLoading(false)

                state.error?.let { msg ->
                    Toast.makeText(requireContext(), msg, Toast.LENGTH_SHORT).show()
                }

                state.profile?.let { profile ->
                    userName.text = profile.name.orEmpty()

                    val roleText = buildString {
                        if (!profile.jobPosition.isNullOrBlank()) append(profile.jobPosition)
                        if (!profile.jobPosition.isNullOrBlank() && !profile.jobPositionType.isNullOrBlank()) append(" • ")
                        if (!profile.jobPositionType.isNullOrBlank()) append(profile.jobPositionType)
                    }

                    userRole.text = roleText.ifBlank { profile.role.orEmpty() }

                    followersCount.text = "..."
                    followingCount.text = "..."

                    val url = profile.photoUrl
                    Log.d("ProfileFragment", "Photo URL: $url")

                    if (!url.isNullOrBlank()) {
                        try {
                            val cleanUrl = url.trim()
                            val freshUrl = "$cleanUrl?t=${System.currentTimeMillis()}"

                            profileImage.load(freshUrl) {
                                crossfade(true)
                                placeholder(R.drawable.profile)
                                error(R.drawable.profile)
                                memoryCachePolicy(CachePolicy.ENABLED)
                                diskCachePolicy(CachePolicy.ENABLED)
                                networkCachePolicy(CachePolicy.ENABLED)
                                addHeader("Cache-Control", "no-cache")
                                target { drawable ->
                                    profileImage.setImageDrawable(drawable)
                                }
                                listener(
                                    onSuccess = { _, result ->
                                        Log.d("ProfileFragment", "✅ Image loaded successfully from: ${result.dataSource}")
                                    },
                                    onError = { _, result ->
                                        Log.e("ProfileFragment", "❌ Image load failed: ${result.throwable.message}")
                                    }
                                )
                            }
                        } catch (e: Exception) {
                            Log.e("ProfileFragment", "Exception loading image: ${e.message}", e)
                            profileImage.setImageResource(R.drawable.profile)
                        }
                    } else {
                        profileImage.setImageResource(R.drawable.profile)
                    }

                    lifecycleScope.launch {
                        try {
                            val (followers, following) = viewModel.getFollowerAndFollowingCounts()
                            followersCount.text = followers.toString()
                            followingCount.text = following.toString()
                        } catch (e: Exception) {
                            Log.e("ProfileFragment", "Error loading follower/following counts: ${e.message}", e)
                            followersCount.text = "0"
                            followingCount.text = "0"
                        }
                    }
                }
            }
        }

        // ✅ Observe Posts ViewModel
        viewLifecycleOwner.lifecycleScope.launch {
            postsViewModel.posts.collect { posts ->
                Log.d("ProfileFragment", "📬 Posts collected: ${posts.size}")

                // ✅ Only update UI if posts section is visible

                    if (posts.isEmpty()) {
                        emptyPostsText.visibility = VISIBLE
                        postsRecyclerView.visibility = GONE
                    } else {
                        emptyPostsText.visibility = GONE
                        postsRecyclerView.visibility = VISIBLE
                        postsAdapter.submitList(posts)
                    }

            }
        }
        viewLifecycleOwner.lifecycleScope.launch {
            postsViewModel.loading.collect { isLoading ->
                if (isLoading) {
                    emptyPostsText.visibility = GONE
                    postsRecyclerView.visibility = GONE
                } else {
                    val posts = postsViewModel.posts.value
                    if (posts.isEmpty()) {
                        emptyPostsText.visibility = VISIBLE
                        postsRecyclerView.visibility = GONE
                    } else {
                        emptyPostsText.visibility = GONE
                        postsRecyclerView.visibility = VISIBLE
                    }
                }
            }
        }


        followersContainer.setOnClickListener {
            navigateToUserList(com.example.finalyearproject.app.profile.viewmodels.ListType.FOLLOWERS)
        }
        followingContainer.setOnClickListener {
            navigateToUserList(com.example.finalyearproject.app.profile.viewmodels.ListType.FOLLOWING)
        }
    }

    override fun onResume() {
        super.onResume()
        // ✅ Refresh profile after editing
        viewModel.loadProfile()
        
        // ✅ Reload scores when returning to this fragment
        loadUserScores()

        // ✅ Reload posts when returning to this fragment
        if (ProfileViewModel.triggerRefresh) {
            postsViewModel.loadMyPosts()
            ProfileViewModel.triggerRefresh = false
        }
    }

    private fun initViews(view: View) {
        loadingSpinner = view.findViewById(R.id.profileLoadingSpinner)
        scroll = view.findViewById(R.id.profileScroll)

        tabRecentSessions = view.findViewById(R.id.tabRecentSessions)
        tabYourPosts = view.findViewById(R.id.tabYourPosts)
        recentSessionsSection = view.findViewById(R.id.recentSessionsSection)
        yourPostsSection = view.findViewById(R.id.yourPostsSection)
        postsRecyclerView = view.findViewById(R.id.postsRecyclerView)
        emptyPostsText = view.findViewById(R.id.emptyPostsText)

        userName = view.findViewById(R.id.userName)
        userRole = view.findViewById(R.id.userRole)
        followersCount = view.findViewById(R.id.followersCount)
        followingCount = view.findViewById(R.id.followingCount)
        profileImage = view.findViewById(R.id.profileImage)
        followersContainer = view.findViewById(R.id.followersContainer)
        followingContainer = view.findViewById(R.id.followingContainer)
        
        // Score views
        tvTechnicalScore = view.findViewById(R.id.tvTechnicalScore)
        tvBehavioralScore = view.findViewById(R.id.tvBehavioralScore)
        tvCvAnalysisScore = view.findViewById(R.id.tvCvAnalysisScore)
        tvCvAnalysisScoreHeader = view.findViewById(R.id.tvCvAnalysisScoreHeader)
        cardTechnicalInterview = view.findViewById(R.id.cardTechnicalInterview)
        cardBehavioralInterview = view.findViewById(R.id.cardBehavioralInterview)
        cardCvAnalysis = view.findViewById(R.id.cardCvAnalysis)
    }
    
    private fun loadUserScores() {
        lifecycleScope.launch {
            try {
                val profileResult = profileRepository.getMe()
                val userId = profileResult.getOrNull()?.id?.toLongOrNull()
                if (userId == null) {
                    Log.e("ProfileFragment", "User not logged in, cannot load scores")
                    return@launch
                }
                
                val scoresResult = aiRepository.getUserScores(userId)
                scoresResult.fold(
                    onSuccess = { scores ->
                        val technicalScore = scores.technical_score.toInt()
                        val behavioralScore = scores.behavioral_score.toInt()
                        val cvScore = scores.cv_analysis_score.toInt()
                        
                        tvTechnicalScore.text = "$technicalScore%"
                        tvBehavioralScore.text = "$behavioralScore%"
                        tvCvAnalysisScore.text = "$cvScore%"
                        tvCvAnalysisScoreHeader.text = "Score: $cvScore%"
                        
                        Log.d("ProfileFragment", "✅ Scores loaded: Technical=$technicalScore%, Behavioral=$behavioralScore%, CV=$cvScore%")
                    },
                    onFailure = { error ->
                        Log.e("ProfileFragment", "Failed to load scores: ${error.message}", error)
                        // Keep default 0% values
                    }
                )
            } catch (e: Exception) {
                Log.e("ProfileFragment", "Error loading scores: ${e.message}", e)
            }
        }
    }
    
    private fun setupScoreCardListeners() {
        cardTechnicalInterview.setOnClickListener {
            parentFragmentManager.beginTransaction()
                .setCustomAnimations(
                    R.anim.enter_from_right,
                    R.anim.exit_to_left,
                    R.anim.enter_from_left,
                    R.anim.exit_to_right
                )
                .replace(R.id.fragment_container, TechnicalAnalysisDetailFragment())
                .addToBackStack(null)
                .commit()
        }
        
        cardBehavioralInterview.setOnClickListener {
            parentFragmentManager.beginTransaction()
                .setCustomAnimations(
                    R.anim.enter_from_right,
                    R.anim.exit_to_left,
                    R.anim.enter_from_left,
                    R.anim.exit_to_right
                )
                .replace(R.id.fragment_container, BehavioralAnalysisDetailFragment())
                .addToBackStack(null)
                .commit()
        }
        
        cardCvAnalysis.setOnClickListener {
            parentFragmentManager.beginTransaction()
                .setCustomAnimations(
                    R.anim.enter_from_right,
                    R.anim.exit_to_left,
                    R.anim.enter_from_left,
                    R.anim.exit_to_right
                )
                .replace(R.id.fragment_container, CvAnalysisDetailFragment())
                .addToBackStack(null)
                .commit()
        }
    }

    private fun setupRecyclerView() {
        postsRecyclerView.layoutManager = LinearLayoutManager(requireContext())

        // ✅ Initialize adapter with delete callback
        postsAdapter = MyPostsAdapter { postId ->
            postsViewModel.deletePost(postId)
        }
        postsRecyclerView.adapter = postsAdapter
    }

    private fun setupToggleListeners() {
        tabRecentSessions.setOnClickListener { selectRecentSessions() }
        tabYourPosts.setOnClickListener {
            selectYourPosts()
            // ✅ Load posts when tab is clicked
            postsViewModel.loadMyPosts()
        }
    }

    private fun selectRecentSessions() {
        tabRecentSessions.background =
            ContextCompat.getDrawable(requireContext(), R.drawable.toggle_selected)
        tabRecentSessions.setTextColor(ContextCompat.getColor(requireContext(), R.color.blue))
        tabRecentSessions.setTypeface(null, Typeface.BOLD)

        tabYourPosts.background =
            ContextCompat.getDrawable(requireContext(), R.drawable.toggle_unselected)
        tabYourPosts.setTextColor(
            ContextCompat.getColor(requireContext(), android.R.color.darker_gray)
        )
        tabYourPosts.setTypeface(null, Typeface.NORMAL)

        recentSessionsSection.visibility = VISIBLE
        yourPostsSection.visibility = GONE
    }

    private fun selectYourPosts() {
        tabYourPosts.background =
            ContextCompat.getDrawable(requireContext(), R.drawable.toggle_selected)
        tabYourPosts.setTextColor(ContextCompat.getColor(requireContext(), R.color.blue))
        tabYourPosts.setTypeface(null, Typeface.BOLD)

        tabRecentSessions.background =
            ContextCompat.getDrawable(requireContext(), R.drawable.toggle_unselected)
        tabRecentSessions.setTextColor(
            ContextCompat.getColor(requireContext(), android.R.color.darker_gray)
        )
        tabRecentSessions.setTypeface(null, Typeface.NORMAL)

        yourPostsSection.visibility = VISIBLE
        recentSessionsSection.visibility = GONE
    }

    private fun showLoading(isLoading: Boolean) {
        loadingSpinner.visibility = if (isLoading) VISIBLE else GONE
        scroll.visibility = if (isLoading) GONE else VISIBLE
    }

    private fun navigateToProfileEdit() {
        parentFragmentManager.beginTransaction()
            .setCustomAnimations(
                R.anim.enter_from_right,
                R.anim.exit_to_left,
                R.anim.enter_from_left,
                R.anim.exit_to_right
            )
            .replace(R.id.fragment_container, ProfileEditFragment())
            .addToBackStack(null)
            .commit()
    }

    private fun navigateToUserList(type: com.example.finalyearproject.app.profile.viewmodels.ListType) {
        val fragment = UserListFragment().apply {
            arguments = Bundle().apply {
                putSerializable("type", type)
            }
        }
        parentFragmentManager.beginTransaction()
            .setCustomAnimations(
                R.anim.enter_from_right,
                R.anim.exit_to_left,
                R.anim.enter_from_left,
                R.anim.exit_to_right
            )
            .replace(R.id.fragment_container, fragment)
            .addToBackStack(null)
            .commit()
    }
}