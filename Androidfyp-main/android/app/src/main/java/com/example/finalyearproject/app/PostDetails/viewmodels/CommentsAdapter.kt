package com.example.finalyearproject.app.postdetails.adapter

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import com.bumptech.glide.Glide
import com.example.finalyearproject.app.repository.models.Comment
import com.example.finalyearproject.databinding.ItemCommentBinding

class CommentsAdapter(
    private val onDeleteClick: (Comment) -> Unit
) : RecyclerView.Adapter<CommentsAdapter.CommentViewHolder>() {

    private var comments = listOf<Comment>()

    fun submitList(list: List<Comment>) {
        comments = list
        notifyDataSetChanged()
    }

    inner class CommentViewHolder(val binding: ItemCommentBinding) :
        RecyclerView.ViewHolder(binding.root)

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): CommentViewHolder {
        val binding = ItemCommentBinding.inflate(
            LayoutInflater.from(parent.context), parent, false
        )
        return CommentViewHolder(binding)
    }

    override fun onBindViewHolder(holder: CommentViewHolder, position: Int) {
        val comment = comments[position]

        holder.binding.tvUserName.text = comment.userName
        holder.binding.tvCommentContent.text = comment.content

        Glide.with(holder.itemView.context)
            .load(comment.userPhotoUrl)
            .into(holder.binding.ivUserPhoto)

        // Show delete only if user is owner
        if (comment.isOwner) {
            holder.binding.tvDeleteComment.visibility = View.VISIBLE
            holder.binding.tvDeleteComment.setOnClickListener {
                onDeleteClick(comment)
            }
        } else {
            holder.binding.tvDeleteComment.visibility = View.GONE
        }
    }

    override fun getItemCount() = comments.size
}
