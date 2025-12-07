package com.example.finalyearproject.app.homescreen.adapter

import android.graphics.Color
import android.os.Build
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import androidx.annotation.RequiresApi
import androidx.core.content.ContextCompat
import androidx.recyclerview.widget.RecyclerView
import com.bumptech.glide.Glide
import com.example.finalyearproject.R
import com.example.finalyearproject.app.models.Post
import java.time.Instant
import java.time.ZoneId
import java.time.ZonedDateTime
import java.time.format.DateTimeFormatter
import java.time.Duration
import java.util.Locale

class PostAdapter(
    private var posts: List<Post>,
    private val onLikeClicked: (Post) -> Unit,
    private val onCommentClicked: (Post) -> Unit,
    private val onFollowClicked: (Post) -> Unit,
    private val onDeletePostClicked: (Post) -> Unit,
    private val onViewLikesClicked: (Post) -> Unit
) : RecyclerView.Adapter<PostAdapter.PostViewHolder>() {

    inner class PostViewHolder(view: View) : RecyclerView.ViewHolder(view) {
        val ivUserPhoto: ImageView = view.findViewById(R.id.ivUserPhoto)
        val tvUserName: TextView = view.findViewById(R.id.tvUserName)
        val tvFollow: TextView = view.findViewById(R.id.tvFollow)
        val tvContent: TextView = view.findViewById(R.id.tvContent)
        val ivPhoto: ImageView = view.findViewById(R.id.ivPhoto)
        val ivLike: ImageView = view.findViewById(R.id.ivLike)
        val tvLikeCount: TextView = view.findViewById(R.id.tvLikeCount)
        val ivComment: ImageView = view.findViewById(R.id.ivComment)
        val tvCommentCount: TextView = view.findViewById(R.id.tvCommentCount)
        val btnComment: LinearLayout = view.findViewById(R.id.btnComment)
        val tvPostTime: TextView = view.findViewById(R.id.tvPostTime)
        val btnDeletePost: ImageView = view.findViewById(R.id.btnDeletePost)
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): PostViewHolder {
        val view = LayoutInflater.from(parent.context)
            .inflate(R.layout.item_post, parent, false)
        return PostViewHolder(view)
    }

    override fun getItemCount(): Int = posts.size

    @RequiresApi(Build.VERSION_CODES.O)
    override fun onBindViewHolder(holder: PostViewHolder, position: Int) {
        val post = posts[position]

        // === USER INFO ===
        holder.tvUserName.text = post.userName
        holder.tvContent.text = post.content

        // === TIMESTAMP DEBUG ===
        Log.d("PostAdapter", "🕒 CreatedAt for post ${post.id}: ${post.createdAt}")

        // === FORMAT TIME ===
        holder.tvPostTime.text = formatPostTime(post.createdAt)

        // === USER PHOTO ===
        Glide.with(holder.itemView.context)
            .load(post.userPhotoUrl)
            .placeholder(R.drawable.ic_profile_placeholder)
            .error(R.drawable.ic_profile_placeholder)
            .circleCrop()
            .into(holder.ivUserPhoto)

        // === FOLLOW / FOLLOWING ===
        if (post.isOwner) {
            holder.tvFollow.visibility = View.GONE
        } else {
            holder.tvFollow.visibility = View.VISIBLE
            holder.tvFollow.text = if (post.isFollowing) "Following" else "+ Follow"
            holder.tvFollow.setTextColor(
                if (post.isFollowing)
                    Color.GRAY
                else
                    ContextCompat.getColor(holder.itemView.context, R.color.blue)
            )
            holder.tvFollow.setOnClickListener { onFollowClicked(post) }
        }

        // === POST IMAGE ===
        if (!post.photoUrl.isNullOrEmpty()) {
            holder.ivPhoto.visibility = View.VISIBLE
            Glide.with(holder.itemView.context)
                .load(post.photoUrl)
                .placeholder(R.drawable.image_placeholder)
                .into(holder.ivPhoto)
        } else {
            holder.ivPhoto.visibility = View.GONE
        }

        // === LIKE ===
        holder.ivLike.setImageResource(
            if (post.isLiked) R.drawable.ic_like_filled else R.drawable.ic_like_outline
        )
        holder.tvLikeCount.text = post.likesCount.toString()
        holder.tvLikeCount.setTextColor(
            if (post.isLiked)
                ContextCompat.getColor(holder.itemView.context, R.color.red)
            else
                ContextCompat.getColor(holder.itemView.context, R.color.black)
        )

        // === COMMENTS ===
        holder.ivComment.setImageResource(R.drawable.ic_comment_outline)
        holder.tvCommentCount.text = post.commentsCount.toString()
        holder.tvCommentCount.setTextColor(Color.DKGRAY)

        // === DELETE BUTTON (Visible only if it's your post) ===
        if (post.isOwner) {
            holder.btnDeletePost.visibility = View.VISIBLE
            holder.btnDeletePost.setOnClickListener { onDeletePostClicked(post) }
        } else {
            holder.btnDeletePost.visibility = View.GONE
        }

        // === CLICK HANDLERS ===
        holder.ivLike.setOnClickListener { onLikeClicked(post) }
        holder.tvLikeCount.setOnClickListener { onViewLikesClicked(post) }
        holder.ivComment.setOnClickListener { onCommentClicked(post) }
        holder.tvCommentCount.setOnClickListener { onCommentClicked(post) }

        // === LONG PRESS DELETE (Backup) ===
        holder.itemView.setOnLongClickListener {
            if (post.isOwner) {
                onDeletePostClicked(post)
                true
            } else false
        }
    }

    fun updateData(newPosts: List<Post>) {
        posts = newPosts
        notifyDataSetChanged()
    }

    // ✅ TIME FORMATTER (UTC safe)
    @RequiresApi(Build.VERSION_CODES.O)
    private fun formatPostTime(isoString: String?): String {
        if (isoString.isNullOrBlank()) return "Just now"
        return try {
            val instant = Instant.parse(isoString)
            val postUtc = ZonedDateTime.ofInstant(instant, ZoneId.of("UTC"))
            val postLocal = postUtc.withZoneSameInstant(ZoneId.systemDefault())
            val nowLocal = ZonedDateTime.now(ZoneId.systemDefault())

            val diff = Duration.between(postLocal, nowLocal)
            val minutes = diff.toMinutes()
            val hours = diff.toHours()
            val days = diff.toDays()

            when {
                minutes < 1 -> "Just now"
                minutes < 60 -> "$minutes min ago"
                hours < 24 -> "$hours hr${if (hours > 1) "s" else ""} ago"
                days == 1L -> "Yesterday"
                days < 7 -> "$days days ago"
                else -> postLocal.format(
                    DateTimeFormatter.ofPattern("MMM d, yyyy", Locale.getDefault())
                )
            }
        } catch (e: Exception) {
            Log.e("PostAdapter", "❌ Time parse error: ${e.message}")
            "Just now"
        }
    }
}
