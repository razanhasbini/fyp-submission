package com.example.finalyearproject.app.notifications.viewmodel

import androidx.lifecycle.LiveData
import androidx.lifecycle.MutableLiveData
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.finalyearproject.app.notifications.repository.NotificationRepository
import com.example.finalyearproject.app.repository.models.Notification
import kotlinx.coroutines.launch

class NotificationViewModel(private val repository: NotificationRepository) : ViewModel() {

    private val _notifications = MutableLiveData<List<Notification>>(emptyList())
    val notifications: LiveData<List<Notification>> get() = _notifications

    /** Load all notifications from the repository */
    fun loadNotifications() {
        viewModelScope.launch {
            try {
                val data = repository.getNotifications()
                _notifications.value = data
            } catch (e: Exception) {
                e.printStackTrace()
                _notifications.value = emptyList()
            }
        }
    }

    /** Mark a notification as read */
    fun markAsRead(notificationId: Long): LiveData<Boolean> {
        val result = MutableLiveData<Boolean>()
        viewModelScope.launch {
            try {
                repository.markNotificationRead(notificationId)
                result.value = true
            } catch (e: Exception) {
                e.printStackTrace()
                result.value = false
            }
        }
        return result
    }
}
