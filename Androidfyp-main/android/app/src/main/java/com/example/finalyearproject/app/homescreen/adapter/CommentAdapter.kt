package com.example.finalyearproject.app.homescreen.adapter

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ImageView
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView
import com.bumptech.glide.Glide
import com.example.finalyearproject.R
import com.example.finalyearproject.app.repository.models.Comment

/**
 * Adapter for displaying comments inside the comments dialog.
 * Supports real-time updates and delete button for user-owned comments.
 */
class CommentAdapter(
    private val comments: MutableList<Comment>,
    private val onDeleteComment: (Comment) -> Unit
) : RecyclerView.Adapter<CommentAdapter.CommentViewHolder>() {

    inner class CommentViewHolder(view: View) : RecyclerView.ViewHolder(view) {
        val ivUserPhoto: ImageView = view.findViewById(R.id.ivUserPhoto)
        val tvUserName: TextView = view.findViewById(R.id.tvUserName)
        val tvCommentContent: TextView = view.findViewById(R.id.tvCommentContent)
        val tvDeleteComment: TextView = view.findViewById(R.id.tvDeleteComment)
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): CommentViewHolder {
        val view = LayoutInflater.from(parent.context)
            .inflate(R.layout.item_comment, parent, false)
        return CommentViewHolder(view)
    }

    override fun onBindViewHolder(holder: CommentViewHolder, position: Int) {
        val comment = comments[position]

        // 💬 Bind text data
        holder.tvUserName.text = comment.userName
        holder.tvCommentContent.text = comment.content

        // 👤 Load user image safely with Glide
        Glide.with(holder.itemView.context)
            .load(comment.userPhotoUrl)
            .placeholder(R.drawable.ic_profile_placeholder)
            .error(R.drawable.ic_profile_placeholder)
            .circleCrop()
            .into(holder.ivUserPhoto)

        // 🗑️ Show delete button only for user-owned comments
        if (comment.isOwner) {
            holder.tvDeleteComment.visibility = View.VISIBLE
            holder.tvDeleteComment.setOnClickListener {
                onDeleteComment(comment)
            }
        } else {
            holder.tvDeleteComment.visibility = View.GONE
        }

        // 🪄 Optional: Allow long press delete as backup
        holder.itemView.setOnLongClickListener {
            if (comment.isOwner) {
                onDeleteComment(comment)
                true
            } else false
        }
    }

    override fun getItemCount(): Int = comments.size

    // ✅ Add a single new comment dynamically
    fun addComment(comment: Comment) {
        comments.add(comment)
        notifyItemInserted(comments.size - 1)
    }

    // ✅ Update all comments at once
    fun updateComments(newComments: List<Comment>) {
        comments.clear()
        comments.addAll(newComments)
        notifyDataSetChanged()
    }

    // ✅ Remove a specific comment by reference
    fun removeComment(comment: Comment) {
        val position = comments.indexOf(comment)
        if (position != -1) {
            comments.removeAt(position)
            notifyItemRemoved(position)
        }
    }

    // ✅ Optional: Clear all comments
    fun clearComments() {
        comments.clear()
        notifyDataSetChanged()
    }
}
