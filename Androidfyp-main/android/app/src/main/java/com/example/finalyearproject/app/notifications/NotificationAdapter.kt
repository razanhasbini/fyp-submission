package com.example.finalyearproject.app.notifications.adapter

import android.util.Log
import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import com.example.finalyearproject.app.repository.models.Notification
import com.example.finalyearproject.databinding.ItemNotificationCardBinding

class NotificationAdapter(
    private val items: List<Notification>,
    private val onClick: (Notification) -> Unit
) : RecyclerView.Adapter<NotificationAdapter.NotificationViewHolder>() {

    inner class NotificationViewHolder(private val binding: ItemNotificationCardBinding) :
        RecyclerView.ViewHolder(binding.root) {

        fun bind(notification: Notification) {
            binding.tvNotificationTitle.text = notification.title
            binding.tvNotificationMessage.text = notification.message
            binding.tvNotificationTime.text = notification.timestamp
            binding.viewUnreadIndicator.visibility =
                if (notification.isRead) RecyclerView.GONE else RecyclerView.VISIBLE

            binding.root.setOnClickListener {
                Log.d("NotificationAdapter", "Item clicked: ${notification.id}")
                onClick(notification)
            }
        }
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): NotificationViewHolder {
        val binding = ItemNotificationCardBinding.inflate(
            LayoutInflater.from(parent.context), parent, false
        )
        return NotificationViewHolder(binding)
    }

    override fun getItemCount() = items.size

    override fun onBindViewHolder(holder: NotificationViewHolder, position: Int) {
        holder.bind(items[position])
    }
}
