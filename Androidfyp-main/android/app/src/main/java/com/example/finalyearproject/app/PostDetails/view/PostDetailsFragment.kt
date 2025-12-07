package com.example.finalyearproject.app.postdetails.view

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import androidx.appcompat.app.AlertDialog
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.recyclerview.widget.LinearLayoutManager
import com.bumptech.glide.Glide
import com.example.finalyearproject.R
import com.example.finalyearproject.app.postdetails.adapter.CommentsAdapter
import com.example.finalyearproject.app.postdetails.adapter.LikesAdapter
import com.example.finalyearproject.app.postdetails.viewmodel.PostDetailsViewModel
import com.example.finalyearproject.app.postdetails.viewmodel.PostDetailsViewModelFactory
import com.example.finalyearproject.app.repository.PostRepository
import com.example.finalyearproject.app.repository.network.RetrofitClient
import com.example.finalyearproject.databinding.FragmentPostDetailsBinding

class PostDetailsFragment : Fragment(R.layout.fragment_post_details) {

    private var _binding: FragmentPostDetailsBinding? = null
    private val binding get() = _binding!!

    private lateinit var commentsAdapter: CommentsAdapter

    private val viewModel by viewModels<PostDetailsViewModel> {
        PostDetailsViewModelFactory(PostRepository(RetrofitClient.create(requireContext())))
    }

    private var currentPostId: String? = null

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        _binding = FragmentPostDetailsBinding.bind(view)

        commentsAdapter = CommentsAdapter { comment ->
            viewModel.deleteComment(comment.id)
        }

        binding.rvComments.layoutManager = LinearLayoutManager(requireContext())
        binding.rvComments.adapter = commentsAdapter

        currentPostId = arguments?.getString("postId")
        currentPostId?.let { viewModel.loadPost(it) }

        setupListeners()
        observeViewModel()
    }

    private fun setupListeners() {
        // Send comment
        binding.btnSendComment.setOnClickListener {
            val text = binding.etComment.text.toString()
            if (text.isNotBlank() && currentPostId != null) {
                viewModel.addComment(text)
                binding.etComment.text.clear()
            }
        }

        // Show list of users who liked (view-only)
        binding.likesLayout.setOnClickListener {
            val postId = currentPostId ?: return@setOnClickListener
            viewModel.fetchLikesList(postId)
        }
    }

    private fun observeViewModel() {
        viewModel.post.observe(viewLifecycleOwner) { post ->
            if (post == null) return@observe

            // Update post info
            binding.tvUserName.text = post.userName
            binding.tvContent.text = post.content

            // Handle post image
            if (!post.photoUrl.isNullOrEmpty()) {
                binding.ivPostImage.visibility = View.VISIBLE
                Glide.with(this).load(post.photoUrl).into(binding.ivPostImage)
            } else {
                binding.ivPostImage.visibility = View.GONE
            }

        }

        viewModel.comments.observe(viewLifecycleOwner) { comments ->
            commentsAdapter.submitList(comments)
        }

        viewModel.likesList.observe(viewLifecycleOwner) { users ->
            if (users.isEmpty()) return@observe

            val dialogView = LayoutInflater.from(requireContext())
                .inflate(R.layout.dialog_likes_list, null)

            val rvLikes = dialogView.findViewById<androidx.recyclerview.widget.RecyclerView>(R.id.rvLikes)
            rvLikes.layoutManager = LinearLayoutManager(requireContext())
            rvLikes.adapter = LikesAdapter(users)

            AlertDialog.Builder(requireContext())
                .setView(dialogView)
                .setCancelable(true)
                .show()
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
