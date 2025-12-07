package com.example.datastore

import android.content.Context
import android.util.Base64
import androidx.datastore.preferences.preferencesDataStore
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.core.edit
import kotlinx.coroutines.flow.first
import java.security.KeyStore
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties

// DataStore instance
val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "user_prefs")

// Key to store JWT
val tokenKey = stringPreferencesKey("token")

// Encryption constants
private const val ANDROID_KEYSTORE = "AndroidKeyStore"
private const val ALIAS = "jwt_encryption_key"
private const val TRANSFORMATION = "AES/GCM/NoPadding"
private const val IV_SIZE = 12 // GCM standard IV size
private const val TAG_SIZE = 128 // GCM tag size in bits

// Generate or retrieve AES encryption key from Android Keystore
private fun getOrCreateSecretKey(): SecretKey {
    val keyStore = KeyStore.getInstance(ANDROID_KEYSTORE)
    keyStore.load(null)

    // If key already exists, return it
    keyStore.getKey(ALIAS, null)?.let {
        return it as SecretKey
    }

    // Create new AES key with proper KeyGenParameterSpec
    val keyGenerator = KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, ANDROID_KEYSTORE)
    val spec = KeyGenParameterSpec.Builder(
        ALIAS,
        KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
    )
        .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
        .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
        .setRandomizedEncryptionRequired(true)
        .setKeySize(256)
        .build()

    keyGenerator.init(spec)
    return keyGenerator.generateKey()
}

// Encrypt a string using AES/GCM/NoPadding
private fun encrypt(input: String): String {
    val cipher = Cipher.getInstance(TRANSFORMATION)
    val secretKey = getOrCreateSecretKey()
    cipher.init(Cipher.ENCRYPT_MODE, secretKey)

    val iv = cipher.iv
    val encrypted = cipher.doFinal(input.toByteArray(Charsets.UTF_8))

    val combined = iv + encrypted
    return Base64.encodeToString(combined, Base64.DEFAULT)
}

// Decrypt a string using AES/GCM/NoPadding
private fun decrypt(encryptedData: String): String {
    val combined = Base64.decode(encryptedData, Base64.DEFAULT)
    val iv = combined.sliceArray(0 until IV_SIZE)
    val encrypted = combined.sliceArray(IV_SIZE until combined.size)

    val cipher = Cipher.getInstance(TRANSFORMATION)
    val secretKey = getOrCreateSecretKey()
    val spec = GCMParameterSpec(TAG_SIZE, iv)
    cipher.init(Cipher.DECRYPT_MODE, secretKey, spec)

    val decrypted = cipher.doFinal(encrypted)
    return String(decrypted, Charsets.UTF_8)
}

// Save JWT token securely
suspend fun saveToken(context: Context, token: String) {
    context.dataStore.edit { prefs ->
        prefs[tokenKey] = encrypt(token)
    }
}

// Get JWT token securely
suspend fun getToken(context: Context): String {
    val prefs = context.dataStore.data.first()
    val encrypted = prefs[tokenKey] ?: return ""
    return try {
        decrypt(encrypted)
    } catch (e: Exception) {
        ""
    }
}
