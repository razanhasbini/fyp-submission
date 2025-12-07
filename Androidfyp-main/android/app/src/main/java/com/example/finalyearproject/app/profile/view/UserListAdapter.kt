package com.example.finalyearproject.app.profile.view

import android.content.Context
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Button
import android.widget.ImageView
import android.widget.ProgressBar
import android.widget.TextView
import android.widget.Toast
import androidx.recyclerview.widget.RecyclerView
import coil.load
import com.example.finalyearproject.R
import com.example.finalyearproject.app.profile.viewmodels.ListType
import com.example.finalyearproject.app.repository.network.RetrofitClient
import com.example.finalyearproject.app.repository.PostRepository
import com.example.finalyearproject.app.repository.models.UserSummary
import com.example.finalyearproject.app.profile.viewmodels.ProfileViewModel
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

class UserListAdapter(
    private val context: Context,
    private val users: MutableList<UserSummary>,
    private val type: ListType
) : RecyclerView.Adapter<UserListAdapter.UserViewHolder>() {

    private val repo = PostRepository(RetrofitClient.create(context))

    inner class UserViewHolder(view: View) : RecyclerView.ViewHolder(view) {
        val ivPhoto: ImageView = view.findViewById(R.id.ivUserPhoto)
        val tvName: TextView = view.findViewById(R.id.tvUserName)
        val btnFollow: Button = view.findViewById(R.id.btnFollow)
        val progressFollow: ProgressBar = view.findViewById(R.id.progressFollow)
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): UserViewHolder {
        val v = LayoutInflater.from(parent.context)
            .inflate(R.layout.item_user_row, parent, false)
        return UserViewHolder(v)
    }

    override fun getItemCount() = users.size

    override fun onBindViewHolder(holder: UserViewHolder, position: Int) {
        val user = users[position]

        holder.tvName.text = user.username ?: user.name ?: "Unknown"
        holder.ivPhoto.load(user.photoUrl) {
            placeholder(R.drawable.ic_profile_placeholder)
            error(R.drawable.ic_profile_placeholder)
        }

        holder.progressFollow.visibility = View.GONE
        holder.btnFollow.isEnabled = true
        holder.btnFollow.alpha = 1f

        // ✅ Now always show follow/unfollow button in both lists
        holder.btnFollow.visibility = View.VISIBLE

        holder.btnFollow.text = if (user.isFollowing == true) "Unfollow" else "Follow"

        holder.btnFollow.setOnClickListener {
            holder.btnFollow.isEnabled = false
            holder.progressFollow.visibility = View.VISIBLE

            // ✨ Fade animation
            holder.progressFollow.alpha = 0f
            holder.progressFollow.animate().alpha(1f).setDuration(200).start()
            holder.btnFollow.animate().alpha(0.5f).setDuration(200).start()

            CoroutineScope(Dispatchers.Main).launch {
                try {
                    val result = if (user.isFollowing == true)
                        repo.unfollowUser(user.id.toString())
                    else
                        repo.followUser(user.id.toString())

                    result.fold(
                        onSuccess = {
                            user.isFollowing = !(user.isFollowing ?: false)
                            holder.btnFollow.text =
                                if (user.isFollowing == true) "Unfollow" else "Follow"
                            Toast.makeText(context, "Action successful", Toast.LENGTH_SHORT).show()
                            ProfileViewModel.triggerRefresh = true
                        },
                        onFailure = {
                            Toast.makeText(context, "Failed: ${it.message}", Toast.LENGTH_SHORT)
                                .show()
                        }
                    )
                } finally {
                    // Restore UI
                    holder.progressFollow.animate().alpha(0f).setDuration(200)
                        .withEndAction { holder.progressFollow.visibility = View.GONE }.start()
                    holder.btnFollow.animate().alpha(1f).setDuration(200).start()
                    holder.btnFollow.isEnabled = true
                }
            }
        }
    }

    fun updateUsers(newUsers: List<UserSummary>) {
        users.clear()
        users.addAll(newUsers)
        notifyDataSetChanged()
    }
}
