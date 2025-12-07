plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.ksp)                       // ✅ keep only once
    id("org.jetbrains.kotlin.kapt")               // ✅ for Glide
    id("org.jetbrains.kotlin.plugin.serialization") version "1.9.0"
}

android {
    namespace = "com.example.finalyearproject"
    compileSdk = 35

    buildFeatures {
        viewBinding = true
    }

    defaultConfig {
        applicationId = "com.example.finalyearproject"
        minSdk = 24
        targetSdk = 35
        versionCode = 1
        versionName = "1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
    }

    signingConfigs {
        create("release") {
            storeFile = file("release.keystore")
            storePassword = "android123"
            keyAlias = "release"
            keyPassword = "android123"
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
            signingConfig = signingConfigs.getByName("release")
        }
        debug {
            signingConfig = signingConfigs.getByName("release")
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
    }

    kotlinOptions {
        jvmTarget = "11"
    }
}

dependencies {
    implementation("com.google.android.material:material:1.9.0")
    implementation("de.hdodenhof:circleimageview:3.1.0")

    // --- Room ---
    implementation(libs.room.runtime)
    implementation(libs.androidx.swiperefreshlayout)
    ksp(libs.room.compiler)
    implementation(libs.room.ktx)

    // --- Lifecycle ---
    implementation(libs.lifecycle.viewmodel.ktx)
    implementation(libs.lifecycle.livedata.ktx)

    // --- Fragment ---
    implementation(libs.fragment.ktx)

    // --- Coroutines ---
    implementation(libs.kotlinx.coroutines.android)

    // --- DataStore ---
    implementation(libs.datastore.preferences)

    // --- Base dependencies ---
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.appcompat)
    implementation(libs.material)
    implementation(libs.androidx.activity)
    implementation(libs.androidx.constraintlayout)
    testImplementation(libs.junit)
    androidTestImplementation(libs.androidx.junit)
    androidTestImplementation(libs.androidx.espresso.core)
    implementation(libs.logging.interceptor)
    implementation(libs.retrofit2.retrofit)
    implementation(libs.converter.gson)

    // --- Coil ---
    implementation("io.coil-kt:coil:2.7.0")
    implementation("androidx.lifecycle:lifecycle-runtime-ktx:2.8.3")

    // ✅ Supabase Dependencies with explicit versions
    implementation("io.github.jan-tennert.supabase:postgrest-kt:2.6.0")
    implementation("io.github.jan-tennert.supabase:gotrue-kt:2.6.0")
    implementation("io.github.jan-tennert.supabase:storage-kt:2.6.0")
    implementation("io.github.jan-tennert.supabase:realtime-kt:2.6.0")

    // Ktor Client - matching version (OkHttp engine for WebSocket support on Android)
    implementation("io.ktor:ktor-client-android:2.3.7")
    implementation("io.ktor:ktor-client-core:2.3.7")
    implementation("io.ktor:ktor-client-okhttp:2.3.7")  // OkHttp engine for WebSocket support
    implementation("io.ktor:ktor-client-websockets:2.3.7")
    implementation("io.ktor:ktor-utils:2.3.7")

    // Kotlinx Serialization
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.6.2")

    // ✅ Glide
    implementation("com.github.bumptech.glide:glide:4.16.0")
    kapt("com.github.bumptech.glide:compiler:4.16.0")

    // ✅ CameraX
    val cameraxVersion = "1.3.0"
    implementation("androidx.camera:camera-core:$cameraxVersion")
    implementation("androidx.camera:camera-camera2:$cameraxVersion")
    implementation("androidx.camera:camera-lifecycle:$cameraxVersion")
    implementation("androidx.camera:camera-view:$cameraxVersion")

    // ✅ ML Kit for Face and Pose Detection (Landmarks)
    implementation("com.google.mlkit:face-detection:16.1.5")
    implementation("com.google.mlkit:pose-detection:18.0.0-beta4")
    implementation("com.google.mlkit:pose-detection-accurate:18.0.0-beta4")

}
