package com.example.finalyearproject.app.profile.view

import android.os.Bundle
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.EditText
import android.widget.ImageView
import android.widget.TextView
import android.widget.Toast
import androidx.fragment.app.Fragment
import androidx.lifecycle.lifecycleScope
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import coil.load
import coil.transform.CircleCropTransformation
import com.example.finalyearproject.R
import com.example.finalyearproject.app.repository.models.UserProfileResponse
import com.example.finalyearproject.app.repository.network.ApiService
import com.example.finalyearproject.app.repository.network.RetrofitClient
import com.google.android.material.button.MaterialButton
import kotlinx.coroutines.launch
import retrofit2.HttpException

class UserProfileFragment : Fragment() {

    private lateinit var etUsernameSearch: EditText
    private lateinit var btnSearchUser: MaterialButton
    private lateinit var cardUserProfile: View

    private lateinit var ivProfilePhoto: ImageView
    private lateinit var tvFullName: TextView
    private lateinit var tvUsername: TextView
    private lateinit var tvEmail: TextView
    private lateinit var tvJobPosition: TextView
    private lateinit var btnFollow: MaterialButton

    private lateinit var rvUserPosts: RecyclerView

    private var currentUserProfile: UserProfileResponse? = null
    private var isFollowing: Boolean = false

    private val defaultPhotoUrl =
        "https://img.freepik.com/free-vector/blue-circle-with-white-user_78370-4707.jpg?semt=ais_hybrid&w=740&q=80"

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        return inflater.inflate(R.layout.fragment_user_profile, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        etUsernameSearch = view.findViewById(R.id.etUsernameSearch)
        btnSearchUser = view.findViewById(R.id.btnSearchUser)
        cardUserProfile = view.findViewById(R.id.cardUserProfile)

        ivProfilePhoto = view.findViewById(R.id.ivProfilePhoto)
        tvFullName = view.findViewById(R.id.tvFullName)
        tvUsername = view.findViewById(R.id.tvUsername)
        tvEmail = view.findViewById(R.id.tvEmail)
        tvJobPosition = view.findViewById(R.id.tvJobPosition)
        btnFollow = view.findViewById(R.id.btnFollow)

        rvUserPosts = view.findViewById(R.id.rvUserPosts)
        rvUserPosts.layoutManager = LinearLayoutManager(requireContext())

        cardUserProfile.visibility = View.GONE

        btnSearchUser.setOnClickListener {
            val username = etUsernameSearch.text.toString().trim()
            if (username.isEmpty()) {
                Toast.makeText(requireContext(), "Enter a username", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            loadUserProfile(username)
        }

        btnFollow.setOnClickListener {
            currentUserProfile?.let { user ->
                if (!isFollowing) {
                    followUser(user.id)
                }
            }
        }
    }

    private fun getUserApi(): ApiService {
        return RetrofitClient.create(requireContext())
    }

    private fun loadUserProfile(username: String) {
        lifecycleScope.launch {
            try {
                Log.d("UserProfileFragment", "Fetching user profile for: $username")

                val userApi = getUserApi()
                val user = userApi.getUserByUsername(username)

                Log.d("UserProfileFragment", "User fetched: $user")

                currentUserProfile = user
                cardUserProfile.visibility = View.VISIBLE

                tvFullName.text = user.name
                tvUsername.text = "@${user.username ?: "unknown"}"
                tvEmail.text = user.email ?: "N/A"
                tvJobPosition.text = user.jobPosition ?: "N/A"

                // Use user's photoUrl or default if null/empty
                val photoUrl = if (!user.photoUrl.isNullOrEmpty()) user.photoUrl else defaultPhotoUrl
                ivProfilePhoto.load(photoUrl) {
                    placeholder(R.drawable.ic_user_placeholder)
                    error(R.drawable.ic_user_placeholder)
                    transformations(CircleCropTransformation())
                }

                btnFollow.text = if (isFollowing) "Following" else "Follow"
                btnFollow.isEnabled = !isFollowing

                // TODO: Load user posts into rvUserPosts

            } catch (e: HttpException) {
                cardUserProfile.visibility = View.GONE
                Log.e("UserProfileFragment", "HTTP error: ${e.code()}, message: ${e.message()}")
                when (e.code()) {
                    401 -> Toast.makeText(requireContext(), "Unauthorized: Invalid token", Toast.LENGTH_LONG).show()
                    404 -> Toast.makeText(requireContext(), "User not found", Toast.LENGTH_SHORT).show()
                    else -> Toast.makeText(requireContext(), "Server error: ${e.code()}", Toast.LENGTH_SHORT).show()
                }
            } catch (e: Exception) {
                cardUserProfile.visibility = View.GONE
                Log.e("UserProfileFragment", "Error fetching user: ${e.localizedMessage}", e)
                Toast.makeText(requireContext(), "Error fetching user", Toast.LENGTH_SHORT).show()
            }
        }
    }

    private fun followUser(userId: String) {
        lifecycleScope.launch {
            try {
                val userApi = getUserApi()
                Log.d("UserProfileFragment", "Following user ID: $userId")

                userApi.followUser(userId)

                Toast.makeText(requireContext(), "Followed successfully", Toast.LENGTH_SHORT).show()
                isFollowing = true
                btnFollow.text = "Following"
                btnFollow.isEnabled = false

            } catch (e: HttpException) {
                Log.e("UserProfileFragment", "Follow HTTP error: ${e.code()}, message: ${e.message()}")
                if (e.code() == 401) {
                    Toast.makeText(requireContext(), "Unauthorized: Invalid token", Toast.LENGTH_LONG).show()
                } else {
                    Toast.makeText(requireContext(), "Failed to follow user", Toast.LENGTH_SHORT).show()
                }
            } catch (e: Exception) {
                Log.e("UserProfileFragment", "Follow exception: ${e.localizedMessage}", e)
                Toast.makeText(requireContext(), "Error following user", Toast.LENGTH_SHORT).show()
            }
        }
    }
}
