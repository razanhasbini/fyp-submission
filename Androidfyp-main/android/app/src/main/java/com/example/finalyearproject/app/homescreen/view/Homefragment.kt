package com.example.finalyearproject.app.homescreen.view

import android.app.AlertDialog
import android.graphics.Color
import android.graphics.Typeface
import android.os.Bundle
import android.util.Log
import android.view.*
import android.widget.*
import androidx.fragment.app.Fragment
import androidx.lifecycle.lifecycleScope
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import com.bumptech.glide.Glide
import com.example.finalyearproject.R
import com.example.finalyearproject.app.homescreen.adapter.PostAdapter
import com.example.finalyearproject.app.homescreen.adapter.CommentAdapter
import com.example.finalyearproject.app.models.Post
import com.example.finalyearproject.app.profile.view.UserProfileFragment
import com.example.finalyearproject.app.repository.PostRepository
import com.example.finalyearproject.app.repository.models.Comment
import com.example.finalyearproject.app.repository.network.RetrofitClient
import com.example.finalyearproject.app.ui.home.*
import kotlinx.coroutines.launch

class Homefragment : Fragment() {

    private lateinit var recyclerView: RecyclerView
    private lateinit var adapter: PostAdapter
    private lateinit var viewModel: HomeViewModel
    private lateinit var progressBar: ProgressBar
    private lateinit var tvEmptyFeed: TextView
    private lateinit var repo: PostRepository
    private lateinit var llForYouTab: LinearLayout
    private lateinit var llFollowingTab: LinearLayout
    private lateinit var tvForYou: TextView
    private lateinit var tvFollowing: TextView
    private lateinit var vForYouIndicator: View
    private lateinit var vFollowingIndicator: View

    // NEW: reference to the "Create Post" clickable area
    private lateinit var llCreatePost: LinearLayout

    // NEW: reference to the search icon
    private lateinit var ivSearchIcon: ImageView

    override fun onCreateView(
        inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?
    ): View {
        val view = inflater.inflate(R.layout.fragment_homefragment, container, false)

        llForYouTab = view.findViewById(R.id.llForYouTab)
        llFollowingTab = view.findViewById(R.id.llFollowingTab)
        tvForYou = view.findViewById(R.id.tvForYou)
        tvFollowing = view.findViewById(R.id.tvFollowing)
        vForYouIndicator = view.findViewById(R.id.vForYouIndicator)
        vFollowingIndicator = view.findViewById(R.id.vFollowingIndicator)
        recyclerView = view.findViewById(R.id.recyclerView)
        progressBar = view.findViewById(R.id.progressBar)
        tvEmptyFeed = view.findViewById(R.id.tvEmptyFeed)

        // NEW: Initialize the Create Post clickable layout
        llCreatePost = view.findViewById(R.id.llCreatePost)

        // NEW: Initialize search icon
        ivSearchIcon = view.findViewById(R.id.ivSearchIcon)
        ivSearchIcon.setOnClickListener {
            // Navigate to SearchFragment
            parentFragmentManager.beginTransaction()
                .replace(R.id.fragment_container, UserProfileFragment()) // Replace with your container ID
                .addToBackStack(null)
                .commit()
        }

        recyclerView.layoutManager = LinearLayoutManager(requireContext())

        lifecycleScope.launch {
            try {
                val api = RetrofitClient.create(requireContext())
                repo = PostRepository(api)
                viewModel = HomeViewModelFactory(repo).create(HomeViewModel::class.java)

                adapter = PostAdapter(
                    emptyList(),
                    onLikeClicked = { handleLike(it) },
                    onCommentClicked = { handleComment(it) },
                    onFollowClicked = { handleFollow(it) },
                    onDeletePostClicked = { handleDeletePost(it) },
                    onViewLikesClicked = { handleViewLikes(it) }
                )

                recyclerView.adapter = adapter
                observeViewModel()
                switchTab("foryou")
            } catch (e: Exception) {
                Log.e("HomeFragmentLog", "❌ Failed init: ${e.message}")
                showErrorMessage("Initialization failed. Please restart the app.")
            }
        }

        llForYouTab.setOnClickListener { switchTab("foryou") }
        llFollowingTab.setOnClickListener { switchTab("following") }

        // NEW: Set click listener to open CreatePostFragment when clicking the input field
        llCreatePost.setOnClickListener {
            parentFragmentManager.beginTransaction()
                .replace(R.id.fragment_container, CreatePostFragment()) // Replace with your container ID
                .addToBackStack(null)
                .commit()
        }

        return view
    }

    private fun observeViewModel() {
        viewLifecycleOwner.lifecycleScope.launch {
            viewModel.uiState.collect { state ->
                when (state) {
                    is HomeUiState.Loading -> showLoading(true)
                    is HomeUiState.Success -> {
                        showLoading(false)
                        if (state.posts.isEmpty()) showEmptyMessage("No posts available.")
                        else {
                            tvEmptyFeed.visibility = View.GONE
                            recyclerView.visibility = View.VISIBLE
                            adapter.updateData(state.posts)
                        }
                    }
                    is HomeUiState.Error -> {
                        showLoading(false)
                        showEmptyMessage(state.message)
                    }
                    is HomeUiState.Empty -> {
                        showLoading(false)
                        showEmptyMessage("No posts yet.")
                    }
                }
            }
        }
    }

    // ===== LIKE =====
    private fun handleLike(post: Post) {
        lifecycleScope.launch {
            val result = if (post.isLiked)
                repo.unlikePost(post.id)
            else repo.likePost(post.id)

            result.fold(
                onSuccess = {
                    post.isLiked = !post.isLiked
                    post.likesCount += if (post.isLiked) 1 else -1
                    adapter.notifyDataSetChanged()
                },
                onFailure = {
                    Toast.makeText(requireContext(), "Error liking post", Toast.LENGTH_SHORT).show()
                }
            )
        }
    }

    private fun handleComment(post: Post) {
        val dialogView = layoutInflater.inflate(R.layout.dialog_comments, null)
        val rvComments = dialogView.findViewById<RecyclerView>(R.id.rvComments)
        val etComment = dialogView.findViewById<EditText>(R.id.etComment)
        val btnSend = dialogView.findViewById<Button>(R.id.btnSend)
        val progressComments = dialogView.findViewById<ProgressBar>(R.id.progressComments)

        rvComments.layoutManager = LinearLayoutManager(requireContext())
        val comments = mutableListOf<Comment>()

        lateinit var commentAdapter: CommentAdapter

        commentAdapter = CommentAdapter(comments) { comment ->
            if (comment.isOwner) {
                AlertDialog.Builder(requireContext())
                    .setTitle("Delete Comment")
                    .setMessage("Are you sure you want to delete this comment?")
                    .setPositiveButton("Delete") { _, _ ->
                        lifecycleScope.launch {
                            repo.deleteComment(post.id, comment.id).fold(
                                onSuccess = {
                                    comments.remove(comment)
                                    post.commentsCount--
                                    commentAdapter.notifyDataSetChanged()
                                    adapter.notifyDataSetChanged()
                                    Toast.makeText(requireContext(), "Comment deleted", Toast.LENGTH_SHORT).show()
                                },
                                onFailure = {
                                    Toast.makeText(requireContext(), "Failed to delete comment", Toast.LENGTH_SHORT).show()
                                }
                            )
                        }
                    }
                    .setNegativeButton("Cancel", null)
                    .show()
            }
        }

        rvComments.adapter = commentAdapter

        progressComments.visibility = View.VISIBLE
        rvComments.visibility = View.GONE

        lifecycleScope.launch {
            repo.getComments(post.id).onSuccess {
                comments.clear()
                comments.addAll(it)
                commentAdapter.notifyDataSetChanged()
                progressComments.visibility = View.GONE
                rvComments.visibility = View.VISIBLE
            }.onFailure {
                progressComments.visibility = View.GONE
                Toast.makeText(requireContext(), "No comments", Toast.LENGTH_SHORT).show()
            }
        }

        val dialog = AlertDialog.Builder(requireContext())
            .setView(dialogView)
            .create()

        btnSend.setOnClickListener {
            val text = etComment.text.toString().trim()
            if (text.isBlank()) return@setOnClickListener

            btnSend.isEnabled = false
            if (text.isNotBlank()) {
                lifecycleScope.launch {
                    repo.addComment(post.id, text).fold(
                        onSuccess = {
                            post.commentsCount++
                            adapter.notifyDataSetChanged()

                            repo.getComments(post.id).onSuccess { newComments ->
                                comments.clear()
                                comments.addAll(newComments)
                                commentAdapter.notifyDataSetChanged()

                                rvComments.smoothScrollToPosition(comments.size - 1)
                            }

                            etComment.text.clear()
                        },
                        onFailure = {
                            Toast.makeText(requireContext(), "Failed to add comment", Toast.LENGTH_SHORT).show()
                        }
                    )
                }
            }
        }

        dialog.show()
    }

    private fun handleViewLikes(post: Post) {
        lifecycleScope.launch {
            repo.getPostLikes(post.id).onSuccess { users ->
                val dialogView = layoutInflater.inflate(R.layout.dialog_likes_list, null)
                val rvLikes = dialogView.findViewById<RecyclerView>(R.id.rvLikes)
                rvLikes.layoutManager = LinearLayoutManager(requireContext())

                rvLikes.adapter = object : RecyclerView.Adapter<RecyclerView.ViewHolder>() {
                    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): RecyclerView.ViewHolder {
                        val v = LayoutInflater.from(parent.context)
                            .inflate(R.layout.item_like_user, parent, false)
                        return object : RecyclerView.ViewHolder(v) {}
                    }

                    override fun onBindViewHolder(holder: RecyclerView.ViewHolder, position: Int) {
                        val user = users[position]
                        val iv = holder.itemView.findViewById<ImageView>(R.id.ivUserPhoto)
                        val tv = holder.itemView.findViewById<TextView>(R.id.tvUserName)
                        tv.text = user.user_name
                        Glide.with(holder.itemView.context)
                            .load(user.photo_url)
                            .placeholder(R.drawable.ic_profile_placeholder)
                            .circleCrop()
                            .into(iv)
                    }

                    override fun getItemCount() = users.size
                }

                AlertDialog.Builder(requireContext())
                    .setView(dialogView)
                    .create()
                    .show()
            }.onFailure {
                Toast.makeText(requireContext(), "Failed to load likes", Toast.LENGTH_SHORT).show()
            }
        }
    }

    private fun handleFollow(post: Post) {
        if (post.isOwner) {
            Toast.makeText(requireContext(), "You cannot follow yourself.", Toast.LENGTH_SHORT).show()
            return
        }

        lifecycleScope.launch {
            val result = if (post.isFollowing)
                repo.unfollowUser(post.userId)
            else repo.followUser(post.userId)

            result.fold(
                onSuccess = {
                    post.isFollowing = !post.isFollowing
                    adapter.notifyDataSetChanged()

                    com.example.finalyearproject.app.profile.viewmodels.ProfileViewModel.triggerRefresh = true
                    Log.d("Homefragment", "🔔 triggerRefresh set to true after follow/unfollow of ${post.userId}")
                },
                onFailure = {
                    Toast.makeText(requireContext(), "Follow action failed", Toast.LENGTH_SHORT).show()
                }
            )
        }
    }

    private fun handleDeletePost(post: Post) {
        AlertDialog.Builder(requireContext())
            .setTitle("Delete Post")
            .setMessage("Are you sure you want to delete this post?")
            .setPositiveButton("Delete") { _, _ ->
                lifecycleScope.launch {
                    repo.deletePost(post.id)
                    viewModel.refresh()
                }
            }
            .setNegativeButton("Cancel", null)
            .show()
    }

    private fun showLoading(isLoading: Boolean) {
        progressBar.visibility = if (isLoading) View.VISIBLE else View.GONE
    }

    private fun showEmptyMessage(message: String) {
        tvEmptyFeed.visibility = View.VISIBLE
        tvEmptyFeed.text = message
        recyclerView.visibility = View.GONE
    }

    private fun showErrorMessage(message: String) = showEmptyMessage(message)

    private fun switchTab(tab: String) {
        if (tab == "foryou") {
            styleTab(true)
            viewModel.onTabSelected(FeedTab.ForYou)
        } else {
            styleTab(false)
            viewModel.onTabSelected(FeedTab.Following)
        }
    }

    private fun styleTab(isForYou: Boolean) {
        if (isForYou) {
            vForYouIndicator.visibility = View.VISIBLE
            vFollowingIndicator.visibility = View.INVISIBLE
            tvForYou.setTextColor(Color.parseColor("#2196F3"))
            tvForYou.setTypeface(null, Typeface.BOLD)
            tvFollowing.setTextColor(Color.GRAY)
        } else {
            vForYouIndicator.visibility = View.INVISIBLE
            vFollowingIndicator.visibility = View.VISIBLE
            tvFollowing.setTextColor(Color.parseColor("#2196F3"))
            tvFollowing.setTypeface(null, Typeface.BOLD)
            tvForYou.setTextColor(Color.GRAY)
        }
    }
}
