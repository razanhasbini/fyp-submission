package com.example.finalyearproject.app.profile_edit.view

import android.net.Uri
import android.os.Bundle
import android.provider.OpenableColumns
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Button
import android.widget.EditText
import android.widget.ImageView
import android.widget.Toast
import androidx.activity.result.contract.ActivityResultContracts
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.lifecycle.lifecycleScope
import coil.load
import com.example.finalyearproject.R
import com.example.finalyearproject.app.profile_edit.viewmodels.ProfileEditViewModel
import com.example.finalyearproject.app.supabase.SupabaseProvider
import io.github.jan.supabase.storage.storage
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.io.File
import java.io.FileOutputStream

class ProfileEditFragment : Fragment() {

    private val vm: ProfileEditViewModel by viewModels()

    private lateinit var nameInput: EditText
    private lateinit var changePhotoButton: ImageView
    private lateinit var profileImage: ImageView
    private lateinit var saveButton: Button
    private lateinit var cancelButton: Button
    private lateinit var backButton: ImageView

    private val pickImage =
        registerForActivityResult(ActivityResultContracts.GetContent()) { uri: Uri? ->
            uri ?: return@registerForActivityResult
            val file = copyToCache(uri) ?: return@registerForActivityResult
            uploadToSupabase(file)
        }

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        return inflater.inflate(R.layout.fragment_profile_edit, container, false)
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        nameInput = view.findViewById(R.id.nameInput)
        changePhotoButton = view.findViewById(R.id.changePhotoButton)
        profileImage = view.findViewById(R.id.profileImage)
        saveButton = view.findViewById(R.id.saveButton)
        cancelButton = view.findViewById(R.id.cancelButton)
        backButton = view.findViewById(R.id.backButton)

        // 🔙 Navigation
        backButton.setOnClickListener { parentFragmentManager.popBackStack() }
        cancelButton.setOnClickListener { parentFragmentManager.popBackStack() }

        // 📸 Upload new photo
        changePhotoButton.setOnClickListener { pickImage.launch("image/*") }

        // 💾 Save name
        saveButton.setOnClickListener {
            val newName = nameInput.text.toString().trim()
            if (newName.isEmpty()) {
                Toast.makeText(requireContext(), "The name cannot be empty", Toast.LENGTH_SHORT).show()
            } else {
                vm.updateName(newName)
            }
        }

        // 🔹 Load user data initially
        vm.loadUserProfile()

        // 👀 Observe ViewModel state
        viewLifecycleOwner.lifecycleScope.launch {
            vm.state.collect { s ->
                s.user?.let { user ->
                    nameInput.setText(user.name ?: "")
                }

                s.error?.let {
                    Toast.makeText(requireContext(), it, Toast.LENGTH_SHORT).show()
                }

                s.photoUrl?.let { url ->
                    profileImage.load(url)
                }

                s.successMessage?.let {
                    Toast.makeText(requireContext(), it, Toast.LENGTH_SHORT).show()
                    parentFragmentManager.popBackStack()
                }
            }
        }
    }

    // ✅ Upload image to Supabase Storage
    private fun uploadToSupabase(file: File) {
        viewLifecycleOwner.lifecycleScope.launch(Dispatchers.IO) {
            try {
                val client = SupabaseProvider.client
                val storage = client.storage
                val bucket = storage.from("missionvillage")

                val path = "users/${System.currentTimeMillis()}_${file.name}"

                // Upload image
                bucket.upload(path, file.readBytes(), upsert = true)

                // Get public URL
                val publicUrl = bucket.publicUrl(path)

                // Switch to main thread
                withContext(Dispatchers.Main) {
                    profileImage.load(publicUrl)
                    Toast.makeText(requireContext(), "Photo uploaded!", Toast.LENGTH_SHORT).show()

                    // ✅ Update the photo URL in backend
                    vm.updatePhotoUrl(publicUrl)
                }

            } catch (e: Exception) {
                withContext(Dispatchers.Main) {
                    Toast.makeText(
                        requireContext(),
                        "Upload failed: ${e.message}",
                        Toast.LENGTH_SHORT
                    ).show()
                }
            }
        }
    }

    // ✅ Copy selected image to cache
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
            Toast.makeText(requireContext(), "Failed to read image", Toast.LENGTH_SHORT).show()
            null
        }
    }
}
