# 🎤 Interview AI System - Complete Setup Guide

An AI-powered interview practice system with real-time speech recognition, CV analysis, and behavioral feedback.

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Backend Setup](#backend-setup)
- [Android App Setup](#android-app-setup)
- [Configuration](#configuration)
- [Running the System](#running-the-system)
- [Troubleshooting](#troubleshooting)
- [Project Structure](#project-structure)

---

## 🛠️ Prerequisites

Before you begin, ensure you have:

1. **Docker Desktop** installed and running
   - Download: https://www.docker.com/products/docker-desktop
   - Verify: `docker --version`

2. **Android Studio** (for building the app)
   - Download: https://developer.android.com/studio
   - Or use command line with Gradle

3. **Supabase Account** (for database)
   - Sign up: https://supabase.com
   - Create a new project
   - Get your database password

4. **OpenAI API Key** (for AI features)
   - Sign up: https://platform.openai.com
   - Create an API key

---

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone <your-repo-url>
cd razzan
```

**First time setup?** See [SETUP.md](SETUP.md) for detailed step-by-step instructions.

### 2. Set Up Backend

```bash
cd wazzafak-ai-main

# Edit docker-compose.yml and add your credentials:
# - DATABASE_URL (Supabase connection string)
# - LLM_API_KEY (OpenAI API key)

# Start backend
docker compose up --build
```

### 3. Set Up Android App

```bash
cd Androidfyp-main/android

# Build APK
./gradlew assembleDebug  # Linux/Mac
.\gradlew.bat assembleDebug  # Windows
```

### 4. Test the System

- Backend: `http://localhost:8089/v1/health`
- Install APK on device/emulator
- Configure network settings (see below)

---

## 🔧 Backend Setup

### Step 1: Configure Database Connection

1. **Get Supabase Connection String:**
   - Go to Supabase Dashboard → Settings → Database
   - Copy "Connection string" → "Connection pooling"
   - Format: `postgresql://postgres.[project-ref]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres`

2. **Update `docker-compose.yml`:**
   
   Open `wazzafak-ai-main/docker-compose.yml` and update:
   
   ```yaml
   environment:
     - DATABASE_URL=postgresql://postgres.YOUR_PROJECT_REF:YOUR_PASSWORD@aws-0-YOUR_REGION.pooler.supabase.com:6543/postgres?sslmode=require
     - LLM_API_KEY=sk-your-openai-api-key-here
   ```

   **Important:** URL-encode special characters in password:
   - `?` becomes `%3F`
   - `@` becomes `%40`
   - `#` becomes `%23`

### Step 2: Start Backend

```bash
cd wazzafak-ai-main

# Start backend (foreground - see logs)
docker compose up --build

# OR start in background
docker compose up -d --build
```

### Step 3: Verify Backend

```bash
# Check if running
docker compose ps

# Test health endpoint
curl http://localhost:8089/v1/health

# View logs
docker compose logs -f interview_backend
```

**Expected output:**
```json
{"ok":true}
```

### Step 4: Initialize Database Schema

The backend will automatically create tables on first run. If you need to manually initialize:

1. Go to Supabase Dashboard → SQL Editor
2. Run the SQL from `wazzafak-ai-main/init.sql`

---

## 📱 Android App Setup

### Step 1: Configure Network Settings

**For Physical Device Testing:**

1. Find your computer's IP address:
   
   **Windows:**
   ```powershell
   ipconfig
   # Look for "IPv4 Address" (e.g., 192.168.1.100)
   ```
   
   **Linux/Mac:**
   ```bash
   hostname -I
   # or
   ip route get 8.8.8.8 | awk '{print $7}'
   ```

2. Update `AiRetrofitClient.kt`:
   
   File: `Androidfyp-main/android/app/src/main/java/com/example/finalyearproject/app/repository/network/AiRetrofitClient.kt`
   
   ```kotlin
   object Urls {
       const val EMULATOR = "http://10.0.2.2:8089/"
       const val PHYSICAL_DEVICE = "http://YOUR_COMPUTER_IP:8089/"  // Replace with your IP
   }
   ```

3. Update `InterviewWebSocketClient.kt`:
   
   File: `Androidfyp-main/android/app/src/main/java/com/example/finalyearproject/app/interview/InterviewWebSocketClient.kt`
   
   ```kotlin
   private val computerIP = "YOUR_COMPUTER_IP"  // Replace with your IP
   private val baseUrl = "ws://$computerIP:8089"
   ```

**For Emulator Testing:**
- Use `10.0.2.2:8089` (already configured by default)
- No changes needed

### Step 2: Build the APK

**Option A: Command Line**

```bash
cd Androidfyp-main/android

# Windows
.\gradlew.bat assembleDebug

# Linux/Mac
./gradlew assembleDebug
```

**APK Location:**
```
Androidfyp-main/android/app/build/outputs/apk/debug/app-debug.apk
```

**Option B: Android Studio**

1. Open Android Studio
2. Open project: `Androidfyp-main/android`
3. Wait for Gradle sync
4. Build → Build Bundle(s) / APK(s) → Build APK(s)

### Step 3: Install on Device

**Method 1: Using ADB**

```bash
# Enable USB Debugging on phone first
adb install app/build/outputs/apk/debug/app-debug.apk
```

**Method 2: Manual Install**

1. Transfer APK to phone
2. Enable "Install from Unknown Sources"
3. Open APK and install

---

## ⚙️ Configuration

### Backend Configuration

All configuration is in `wazzafak-ai-main/docker-compose.yml`:

```yaml
environment:
  - DATABASE_URL=...          # Supabase connection
  - LLM_API_KEY=...           # OpenAI API key
  - USE_MOCK=0                # Set to 1 for testing without API
  - PORT=8089                 # Backend port
  - CHAT_MODEL=gpt-4o-mini    # Chat model
  - EVAL_MODEL=gpt-4o         # Evaluation model
```

### Mock Mode (Testing Without OpenAI)

To test without OpenAI API calls:

```yaml
- USE_MOCK=1
- LLM_API_KEY=  # Can be empty
```

### Android App Configuration

**Network Settings:**
- Emulator: `http://10.0.2.2:8089`
- Physical Device: `http://YOUR_IP:8089`

**Permissions:**
- Camera (for interview practice)
- Microphone (for speech recognition)
- Internet (for API calls)

---

## 🏃 Running the System

### Daily Workflow

**1. Start Backend:**
```bash
cd wazzafak-ai-main
docker compose up -d
```

**2. Verify Backend:**
```bash
curl http://localhost:8089/v1/health
docker compose logs interview_backend
```

**3. Run Android App:**
- Install APK on device/emulator
- Ensure phone and computer on same WiFi (physical device)
- Open app and test features

**4. Stop Backend:**
```bash
cd wazzafak-ai-main
docker compose down
```

### Testing Checklist

- [ ] Backend running (`docker compose ps`)
- [ ] Health check works (`curl http://localhost:8089/v1/health`)
- [ ] Android app installed
- [ ] Network settings configured
- [ ] Phone and computer on same WiFi (physical device)
- [ ] Permissions granted (camera, microphone)

---

## 🐛 Troubleshooting

### Backend Issues

**Problem: Backend won't start**

```bash
# Check Docker is running
docker ps

# Check logs
docker compose logs interview_backend

# Common fixes:
# - Verify DATABASE_URL is correct
# - Check port 8089 is not in use
# - Ensure Docker Desktop is running
```

**Problem: Database connection error**

- Verify Supabase database is accessible
- Check password is URL-encoded (`?` → `%3F`)
- Verify connection string format
- Check Supabase project is active

**Problem: Port 8089 already in use**

```bash
# Find process using port
# Windows
netstat -ano | findstr :8089

# Linux/Mac
lsof -i :8089

# Change port in docker-compose.yml
- PORT=8089
ports:
  - "8089:8089"  # Change both
```

### Android App Issues

**Problem: Can't connect to backend (Physical Device)**

1. Verify phone and computer on same WiFi
2. Check computer's IP address is correct
3. Test in phone browser: `http://YOUR_IP:8089/v1/health`
4. Check Windows Firewall allows port 8089:
   ```powershell
   New-NetFirewallRule -DisplayName "Interview Backend" -Direction Inbound -LocalPort 8089 -Protocol TCP -Action Allow
   ```

**Problem: Can't connect to backend (Emulator)**

- Use `10.0.2.2:8089` (not `localhost`)
- Verify backend is running
- Check emulator network settings

**Problem: Speech recognition not working**

- Grant microphone permission
- Check Android version (requires API 24+)
- Verify no other app is using microphone
- Check logs: `adb logcat | grep SpeechRecognizer`

**Problem: CV upload fails**

- Check backend logs for errors
- Verify file format (PDF or DOCX)
- Check file size (should be < 10MB)
- Verify backend is running

---

## 📁 Project Structure

```
razzan/
├── wazzafak-ai-main/          # Go Backend
│   ├── main.go                 # Main server, HTTP handlers, WebSocket
│   ├── cv_parser.go            # CV upload handler, text extraction
│   ├── docker-compose.yml      # Docker configuration
│   ├── Dockerfile              # Docker build file
│   ├── go.mod                  # Go dependencies
│   ├── init.sql                # Database schema
│   └── BACKEND_RUN_GUIDE.md    # Detailed backend guide
│
└── Androidfyp-main/           # Android Application
    └── android/
        ├── app/
        │   └── src/main/
        │       ├── java/...    # Kotlin source code
        │       └── AndroidManifest.xml
        └── build.gradle.kts    # Gradle configuration
```

---

## 📡 API Endpoints

### Health Check
```
GET /v1/health
Response: {"ok":true}
```

### CV Upload
```
POST /v1/cv/upload
Content-Type: multipart/form-data
Body: user_id, file
```

### Live Interview (WebSocket)
```
WS /v1/live-interview?user_id=1
```

### User Scores
```
GET /v1/user/scores?user_id=1
```

### Feedback
```
GET /v1/user/feedback/cv?user_id=1
GET /v1/user/feedback/technical?user_id=1
GET /v1/user/feedback/behavioral?user_id=1
```

---

## 🔐 Security Notes

**For Production:**

1. **Never commit secrets:**
   - Add `.env` to `.gitignore`
   - Use environment variables or secrets management
   - Don't hardcode API keys

2. **Database:**
   - Use connection pooling (already configured)
   - Enable SSL (`sslmode=require`)
   - Use strong passwords

3. **API Keys:**
   - Rotate keys regularly
   - Use environment variables
   - Don't expose in client code

---

## 📚 Additional Resources

- **Backend Guide:** `wazzafak-ai-main/BACKEND_RUN_GUIDE.md`
- **Architecture:** `ARCHITECTURE_DIAGRAMS.md`
- **Technical Report:** `COMPLETE_TECHNICAL_REPORT.md`

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

---

## 📝 License

[Add your license here]

---

## 🆘 Support

If you encounter issues:

1. Check the troubleshooting section
2. Review backend logs: `docker compose logs interview_backend`
3. Check Android logs: `adb logcat`
4. Verify all prerequisites are installed
5. Ensure configuration is correct

---

## ✅ Quick Reference

```bash
# Backend
cd wazzafak-ai-main
docker compose up -d              # Start
docker compose down               # Stop
docker compose logs -f            # View logs
curl http://localhost:8089/v1/health  # Test

# Android
cd Androidfyp-main/android
./gradlew assembleDebug           # Build APK
adb install app/build/outputs/apk/debug/app-debug.apk  # Install
```

---

**Happy Coding! 🚀**

