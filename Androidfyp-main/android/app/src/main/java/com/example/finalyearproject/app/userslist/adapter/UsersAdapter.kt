package com.example.finalyearproject.app.userslist.view

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ImageView
import android.widget.TextView
import androidx.recyclerview.widget.DiffUtil
import androidx.recyclerview.widget.ListAdapter
import androidx.recyclerview.widget.RecyclerView
import coil.load
import com.example.finalyearproject.R
import com.example.finalyearproject.app.repository.models.UserSummary

class UsersAdapter(
    private val onClick: (UserSummary) -> Unit
) : ListAdapter<UserSummary, UsersAdapter.VH>(DIFF) {

    object DIFF: DiffUtil.ItemCallback<UserSummary>() {
        override fun areItemsTheSame(oldItem: UserSummary, newItem: UserSummary) = oldItem.id == newItem.id
        override fun areContentsTheSame(oldItem: UserSummary, newItem: UserSummary) = oldItem == newItem
    }

    inner class VH(itemView: View): RecyclerView.ViewHolder(itemView) {
        val photo: ImageView = itemView.findViewById(R.id.ivUserPhoto)
        val name: TextView = itemView.findViewById(R.id.userName)
        val meta: TextView = itemView.findViewById(R.id.userRole)
        init {
            itemView.setOnClickListener {
                val pos = adapterPosition
                if (pos != RecyclerView.NO_POSITION) {
                    getItem(pos)?.let(onClick)
                }
            }
        }

    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): VH {
        val v = LayoutInflater.from(parent.context).inflate(R.layout.item_user_row, parent, false)
        return VH(v)
    }

    override fun onBindViewHolder(holder: VH, position: Int) {
        val u = getItem(position)
        if (!u.photoUrl.isNullOrBlank()) holder.photo.load(u.photoUrl) { placeholder(R.drawable.profile) }
        else holder.photo.setImageResource(R.drawable.profile)
        holder.name.text = u.name
        holder.meta.text = listOfNotNull("@${u.username}".takeIf { !u.username.isNullOrBlank() }, u.role).joinToString(" • ")
    }
}
