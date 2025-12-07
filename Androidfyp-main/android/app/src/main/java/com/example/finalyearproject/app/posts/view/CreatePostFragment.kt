package com.example.finalyearproject.app.homescreen.view

import android.net.Uri
import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.*
import androidx.activity.result.contract.ActivityResultContracts
import androidx.fragment.app.Fragment
import androidx.lifecycle.lifecycleScope
import coil.load
import com.example.finalyearproject.R
import com.example.finalyearproject.app.repository.PostRepository
import com.example.finalyearproject.app.repository.network.RetrofitClient
import com.example.finalyearproject.app.supabase.SupabaseProvider
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.io.File
import java.io.FileOutputStream
import android.provider.OpenableColumns
import io.github.jan.supabase.storage.storage

class CreatePostFragment : Fragment() {

    private lateinit var etPostContent: EditText
    private lateinit var ivSelectedImage: ImageView
    private lateinit var btnSelectImage: Button
    private lateinit var btnRemoveImage: ImageButton
    private lateinit var btnPost: Button
    private lateinit var btnCancel: Button
    private lateinit var progressBar: ProgressBar
    private lateinit var repo: PostRepository

    private var selectedImageUri: Uri? = null
    private var uploadedImageUrl: String? = null

    var onPostCreated: (() -> Unit)? = null

    private val pickImage =
        registerForActivityResult(ActivityResultContracts.GetContent()) { uri: Uri? ->
            uri ?: return@registerForActivityResult
            selectedImageUri = uri
            ivSelectedImage.load(uri)
            ivSelectedImage.visibility = View.VISIBLE
            btnRemoveImage.visibility = View.VISIBLE
        }

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        val view = inflater.inflate(R.layout.fragment_create_post, container, false)

        etPostContent = view.findViewById(R.id.etPostContent)
        ivSelectedImage = view.findViewById(R.id.ivSelectedImage)
        btnSelectImage = view.findViewById(R.id.btnSelectImage)
        btnRemoveImage = view.findViewById(R.id.btnRemoveImage)
        btnPost = view.findViewById(R.id.btnPost)
        btnCancel = view.findViewById(R.id.btnCancel)
        progressBar = view.findViewById(R.id.progressBar)

        lifecycleScope.launch {
            repo = PostRepository(RetrofitClient.create(requireContext()))
        }

        setupListeners()

        return view
    }

    private fun setupListeners() {
        btnSelectImage.setOnClickListener { pickImage.launch("image/*") }

        btnRemoveImage.setOnClickListener {
            selectedImageUri = null
            uploadedImageUrl = null
            ivSelectedImage.visibility = View.GONE
            btnRemoveImage.visibility = View.GONE
        }

        btnPost.setOnClickListener { uploadAndCreatePost() }

        btnCancel.setOnClickListener {
            parentFragmentManager.popBackStack()
        }
    }

    private fun uploadAndCreatePost() {
        val content = etPostContent.text.toString().trim()
        if (content.isEmpty()) {
            Toast.makeText(requireContext(), "Please write something", Toast.LENGTH_SHORT).show()
            return
        }

        setLoading(true)

        // If user selected an image, upload first
        if (selectedImageUri != null) {
            lifecycleScope.launch(Dispatchers.IO) {
                val file = copyToCache(selectedImageUri!!) ?: run {
                    withContext(Dispatchers.Main) {
                        Toast.makeText(requireContext(), "Failed to read image", Toast.LENGTH_SHORT).show()
                        setLoading(false)
                    }
                    return@launch
                }

                try {
                    val client = SupabaseProvider.client
                    val bucket = client.storage.from("missionvillage")
                    val path = "posts/${System.currentTimeMillis()}_${file.name}"

                    bucket.upload(path, file.readBytes(), upsert = true)
                    uploadedImageUrl = bucket.publicUrl(path)

                    withContext(Dispatchers.Main) { createPost(content) }

                } catch (e: Exception) {
                    withContext(Dispatchers.Main) {
                        Toast.makeText(requireContext(), "Image upload failed: ${e.message}", Toast.LENGTH_SHORT).show()
                        setLoading(false)
                    }
                }
            }
        } else {
            createPost(content)
        }
    }

    private fun createPost(content: String) {
        lifecycleScope.launch {
            val result = repo.createPost(content, uploadedImageUrl)

            result.fold(
                onSuccess = {
                    Toast.makeText(requireContext(), "Post created!", Toast.LENGTH_SHORT).show()
                    onPostCreated?.invoke()
                    parentFragmentManager.popBackStack()
                },
                onFailure = {
                    Toast.makeText(requireContext(), "Failed: ${it.message}", Toast.LENGTH_SHORT).show()
                }
            )

            setLoading(false)
        }
    }

    private fun setLoading(isLoading: Boolean) {
        progressBar.visibility = if (isLoading) View.VISIBLE else View.GONE
        btnPost.isEnabled = !isLoading
    }

    private fun copyToCache(uri: Uri): File? {
        val contentResolver = requireContext().contentResolver
        val fileName = run {
            var name = "upload.jpg"
            contentResolver.query(uri, null, null, null, null)?.use { cursor ->
                val nameIndex = cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME)
                if (cursor.moveToFirst() && nameIndex >= 0) {
                    name = cursor.getString(nameIndex)
                }
            }
            name
        }

        val outFile = File(requireContext().cacheDir, fileName)

        return try {
            contentResolver.openInputStream(uri)?.use { input ->
                FileOutputStream(outFile).use { output -> input.copyTo(output) }
            }
            outFile
        } catch (e: Exception) {
            null
        }
    }
}
