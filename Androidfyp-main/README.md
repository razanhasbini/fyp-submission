# 📱 Interview AI Android App

Android application for practicing interviews with AI-powered feedback, CV analysis, and real-time speech recognition.

## 🚀 Quick Start

### Prerequisites

- Android Studio (or Gradle command line)
- Backend server running (see `wazzafak-ai-main/README.md`)
- Android device or emulator (API 24+)

### Setup

1. **Configure Network Settings:**

   **For Physical Device:**
   
   Find your computer's IP:
   ```bash
   # Windows
   ipconfig
   
   # Linux/Mac
   hostname -I
   ```
   
   Update `AiRetrofitClient.kt`:
   ```kotlin
   const val PHYSICAL_DEVICE = "http://YOUR_IP:8089/"
   ```
   
   Update `InterviewWebSocketClient.kt`:
   ```kotlin
   private val computerIP = "YOUR_IP"
   ```
   
   **For Emulator:**
   - Use `10.0.2.2:8089` (already configured)

2. **Build APK:**

   ```bash
   cd android
   
   # Windows
   .\gradlew.bat assembleDebug
   
   # Linux/Mac
   ./gradlew assembleDebug
   ```

3. **Install:**

   ```bash
   adb install app/build/outputs/apk/debug/app-debug.apk
   ```

## 📋 Requirements

- **Min SDK:** 24 (Android 7.0)
- **Target SDK:** 35 (Android 15)
- **Permissions:**
  - Camera (for interview practice)
  - Microphone (for speech recognition)
  - Internet (for API calls)

## 🔧 Configuration

### Network Settings

- **Emulator:** `http://10.0.2.2:8089`
- **Physical Device:** `http://YOUR_COMPUTER_IP:8089`

**Important:** Phone and computer must be on the same WiFi network for physical device testing.

### Backend Connection

Ensure backend is running before using the app:
```bash
cd ../wazzafak-ai-main
docker compose up -d
```

## 🎯 Features

- **CV Analysis:** Upload and analyze CV documents
- **Live Interview:** Practice interviews with AI interviewer
- **Speech Recognition:** Real-time voice input
- **Feedback:** Get detailed technical and behavioral feedback
- **Camera Analysis:** Eye contact and posture detection

## 🐛 Troubleshooting

**Can't connect to backend:**
- Verify backend is running
- Check network settings (IP address)
- Ensure same WiFi network (physical device)
- Test: Open `http://YOUR_IP:8089/v1/health` in phone browser

**Speech recognition not working:**
- Grant microphone permission
- Check Android version (requires API 24+)
- Verify no other app using microphone

**Build errors:**
- Sync Gradle: `./gradlew --refresh-dependencies`
- Clean build: `./gradlew clean`
- Check Android SDK is installed

## 📝 License

[Add your license]






