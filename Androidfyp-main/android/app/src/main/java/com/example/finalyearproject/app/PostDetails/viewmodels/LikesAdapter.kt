package com.example.finalyearproject.app.postdetails.adapter

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView
import com.example.finalyearproject.R
import com.example.finalyearproject.app.repository.models.LikeUserResponse

class LikesAdapter(
    private val likes: List<LikeUserResponse>
) : RecyclerView.Adapter<LikesAdapter.LikeViewHolder>() {

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): LikeViewHolder {
        val view = LayoutInflater.from(parent.context)
            .inflate(android.R.layout.simple_list_item_1, parent, false)
        return LikeViewHolder(view)
    }

    override fun onBindViewHolder(holder: LikeViewHolder, position: Int) {
        holder.bind(likes[position])
    }

    override fun getItemCount(): Int = likes.size

    class LikeViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        private val tvName: TextView = itemView.findViewById(android.R.id.text1)
        fun bind(user: LikeUserResponse) {
            tvName.text = user.user_name ?: "Unknown User"
        }
    }
}
