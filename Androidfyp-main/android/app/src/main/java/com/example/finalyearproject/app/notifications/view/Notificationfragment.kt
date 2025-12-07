package com.example.finalyearproject.app.notifications.view

import android.os.Bundle
import android.util.Log
import android.view.View
import android.widget.Toast
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.recyclerview.widget.LinearLayoutManager
import com.example.finalyearproject.R
import com.example.finalyearproject.app.notifications.adapter.NotificationAdapter
import com.example.finalyearproject.app.notifications.repository.NotificationRepository
import com.example.finalyearproject.app.notifications.viewmodel.NotificationViewModel
import com.example.finalyearproject.app.notifications.viewmodel.NotificationViewModelFactory
import com.example.finalyearproject.app.postdetails.view.PostDetailsFragment
import com.example.finalyearproject.app.profile.view.UserListFragment
import com.example.finalyearproject.app.profile.viewmodels.ListType
import com.example.finalyearproject.app.repository.network.RetrofitClient
import com.example.finalyearproject.databinding.FragmentNotificationfragmentBinding

class NotificationFragment : Fragment(R.layout.fragment_notificationfragment) {

    private var _binding: FragmentNotificationfragmentBinding? = null
    private val binding get() = _binding!!

    private val viewModel by viewModels<NotificationViewModel> {
        NotificationViewModelFactory(
            NotificationRepository(RetrofitClient.create(requireContext()))
        )
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        _binding = FragmentNotificationfragmentBinding.bind(view)

        // RecyclerView
        binding.rvNotifications.layoutManager = LinearLayoutManager(requireContext())

        // Swipe-to-refresh
        binding.swipeRefreshNotifications.setOnRefreshListener {
            Log.d("NotificationFragment", "Swipe refresh triggered")
            refreshNotifications()
        }

        observeNotifications()

        // Initial load
        refreshNotifications()
    }

    private fun refreshNotifications() {
        binding.swipeRefreshNotifications.isRefreshing = true
        viewModel.loadNotifications()
    }

    private fun observeNotifications() {
        viewModel.notifications.observe(viewLifecycleOwner) { notifications ->
            binding.swipeRefreshNotifications.isRefreshing = false
            Log.d("NotificationFragment", "Notifications received: ${notifications.size}")

            if (notifications.isEmpty()) {
                binding.tvEmptyNotifications.visibility = View.VISIBLE
                binding.rvNotifications.visibility = View.GONE
            } else {
                binding.tvEmptyNotifications.visibility = View.GONE
                binding.rvNotifications.visibility = View.VISIBLE

                binding.rvNotifications.adapter = NotificationAdapter(notifications) { notification ->
                    Log.d("NotificationFragment",
                        "Clicked notification: ${notification.id} / ${notification.type}"
                    )

                    // Mark as read
                    viewModel.markAsRead(notification.id).observe(viewLifecycleOwner) {
                        Log.d("NotificationFragment", "Marked as read: ${notification.id}")
                        binding.rvNotifications.adapter?.notifyDataSetChanged()
                    }

                    // Handle navigation by type
                    when (notification.type.uppercase()) {
                        "FOLLOW" -> navigateToFollowers()

                        "LIKE", "COMMENT" -> {
                            if (notification.relatedId != null) {
                                navigateToPostDetails(notification.relatedId)
                            } else {
                                Toast.makeText(requireContext(),
                                    "Post ID missing in notification",
                                    Toast.LENGTH_SHORT
                                ).show()
                            }
                        }

                        else -> Log.d("NotificationFragment",
                            "Unknown type: ${notification.type}"
                        )
                    }
                }
            }
        }
    }

    private fun navigateToFollowers() {
        Log.d("NotificationFragment", "Navigating to followers fragment")
        val fragment = UserListFragment().apply {
            arguments = Bundle().apply {
                putSerializable("type", ListType.FOLLOWERS)
            }
        }

        requireActivity().supportFragmentManager.beginTransaction()
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

    private fun navigateToPostDetails(postId: String) {
        Log.d("NotificationFragment", "Navigating to post details: $postId")

        val fragment = PostDetailsFragment().apply {
            arguments = Bundle().apply {
                putString("postId", postId)
            }
        }

        requireActivity().supportFragmentManager.beginTransaction()
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

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
