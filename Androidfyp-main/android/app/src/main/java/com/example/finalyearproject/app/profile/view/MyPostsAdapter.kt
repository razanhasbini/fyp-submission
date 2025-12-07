package com.example.finalyearproject.app.profile.view

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ImageButton
import android.widget.ImageView
import android.widget.TextView
import androidx.recyclerview.widget.DiffUtil
import androidx.recyclerview.widget.ListAdapter
import androidx.recyclerview.widget.RecyclerView
import coil.load
import com.example.finalyearproject.R
import com.example.finalyearproject.app.repository.models.PostResponse

class MyPostsAdapter(
    private val onDeleteClick: (String) -> Unit
) : ListAdapter<PostResponse, MyPostsAdapter.PostViewHolder>(PostDiffCallback()) {

    inner class PostViewHolder(view: View) : RecyclerView.ViewHolder(view) {
        val postImage: ImageView = view.findViewById(R.id.postImage)
        val postCaption: TextView = view.findViewById(R.id.postCaption)
        val postDate: TextView = view.findViewById(R.id.postDate)
        val deleteButton: ImageButton = view.findViewById(R.id.deleteButton)
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): PostViewHolder {
        val view = LayoutInflater.from(parent.context)
            .inflate(R.layout.item_my_post, parent, false)
        return PostViewHolder(view)
    }

    override fun onBindViewHolder(holder: PostViewHolder, position: Int) {
        val post = getItem(position)

        holder.postCaption.text = post.content ?: "No caption"
        holder.postDate.text = post.created_at ?: ""

        if (!post.photo_url.isNullOrBlank()) {
            holder.postImage.load(post.photo_url) {
                placeholder(R.drawable.image_placeholder)
                error(R.drawable.image_placeholder)
                crossfade(true)
            }
        } else {
            holder.postImage.setImageResource(R.drawable.image_placeholder)
        }

        // 🔥 Added confirmation dialog (ONLY addition)
        holder.deleteButton.setOnClickListener {
            val context = holder.itemView.context

            androidx.appcompat.app.AlertDialog.Builder(context)
                .setTitle("Delete Post?")
                .setMessage("Are you sure you want to delete this post?")
                .setPositiveButton("Delete") { _, _ ->
                    post.id?.let { onDeleteClick(it) }
                }
                .setNegativeButton("Cancel", null)
                .show()
        }
    }

    class PostDiffCallback : DiffUtil.ItemCallback<PostResponse>() {
        override fun areItemsTheSame(oldItem: PostResponse, newItem: PostResponse): Boolean =
            oldItem.id == newItem.id

        override fun areContentsTheSame(oldItem: PostResponse, newItem: PostResponse): Boolean =
            oldItem == newItem
    }
}
