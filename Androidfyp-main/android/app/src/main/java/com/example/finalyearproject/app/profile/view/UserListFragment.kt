package com.example.finalyearproject.app.profile.view

import android.os.Bundle
import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ProgressBar
import android.widget.TextView
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.lifecycle.lifecycleScope
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import com.example.finalyearproject.R
import com.example.finalyearproject.app.profile.viewmodels.ListType
import com.example.finalyearproject.app.profile.viewmodels.UserListViewModel
import kotlinx.coroutines.launch

class UserListFragment : Fragment() {

    private val viewModel: UserListViewModel by viewModels()
    private lateinit var recyclerView: RecyclerView
    private lateinit var progressBar: ProgressBar
    private lateinit var emptyText: TextView
    private lateinit var adapter: UserListAdapter
    private var listType: ListType = ListType.FOLLOWERS

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        listType = arguments?.getSerializable("type") as? ListType ?: ListType.FOLLOWERS
        Log.d("UserListFragment", "📥 Opened for $listType")
    }

    override fun onCreateView(
        inflater: LayoutInflater, container: ViewGroup?, savedInstanceState: Bundle?
    ): View {
        val view = inflater.inflate(R.layout.fragment_user_list, container, false)
        recyclerView = view.findViewById(R.id.rvUserList)
        progressBar = view.findViewById(R.id.progressUserList)
        emptyText = view.findViewById(R.id.tvEmptyUserList)

        recyclerView.layoutManager = LinearLayoutManager(requireContext())
        adapter = UserListAdapter(requireContext(), mutableListOf(), listType)
        recyclerView.adapter = adapter

        observeViewModel()
        viewModel.load(listType)
        return view
    }

    private fun observeViewModel() {
        viewLifecycleOwner.lifecycleScope.launch {
            viewModel.state.collect { state ->
                when {
                    state.loading -> {
                        progressBar.visibility = View.VISIBLE
                        recyclerView.visibility = View.GONE
                        emptyText.visibility = View.GONE
                    }

                    state.error != null -> {
                        progressBar.visibility = View.GONE
                        recyclerView.visibility = View.GONE
                        emptyText.visibility = View.VISIBLE
                        emptyText.text = "Error: ${state.error}"
                    }

                    state.users.isEmpty() -> {
                        progressBar.visibility = View.GONE
                        recyclerView.visibility = View.GONE
                        emptyText.visibility = View.VISIBLE
                        emptyText.text = "No users found."
                    }

                    else -> {
                        progressBar.visibility = View.GONE
                        emptyText.visibility = View.GONE
                        recyclerView.visibility = View.VISIBLE

                        // ✅ Update adapter with new list
                        adapter.updateUsers(state.users)
                    }
                }
            }
        }
    }
}
