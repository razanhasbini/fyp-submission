# 🚀 First-Time Setup Guide

Complete step-by-step guide for setting up the project after cloning from GitHub.

## 📋 Step-by-Step Setup

### Step 1: Clone the Repository

```bash
git clone <your-repo-url>
cd razzan
```

### Step 2: Set Up Supabase Database

1. **Create Supabase Account:**
   - Go to https://supabase.com
   - Sign up and create a new project
   - Wait for project to be ready (~2 minutes)

2. **Get Database Connection String:**
   - Go to: Settings → Database
   - Scroll to "Connection string"
   - Select "Connection pooling" tab
   - Copy the connection string
   - Format: `postgresql://postgres.[project-ref]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres`

3. **Initialize Database Schema:**
   - Go to: SQL Editor
   - Copy contents of `wazzafak-ai-main/init.sql`
   - Paste and run in SQL Editor

### Step 3: Get OpenAI API Key

1. **Create OpenAI Account:**
   - Go to https://platform.openai.com
   - Sign up and add payment method
   - Go to API Keys section

2. **Create API Key:**
   - Click "Create new secret key"
   - Copy the key (starts with `sk-`)
   - Save it securely (you won't see it again)

### Step 4: Configure Backend

1. **Edit `wazzafak-ai-main/docker-compose.yml`:**

   Find these lines and replace with your values:
   
   ```yaml
   environment:
     # Replace with your Supabase connection string
     - DATABASE_URL=postgresql://postgres.YOUR_PROJECT_REF:YOUR_PASSWORD@aws-0-YOUR_REGION.pooler.supabase.com:6543/postgres?sslmode=require
     
     # Replace with your OpenAI API key
     - LLM_API_KEY=sk-your-openai-api-key-here
   ```

   **Important:** URL-encode special characters in password:
   - `?` → `%3F`
   - `@` → `%40`
   - `#` → `%23`
   - `&` → `%26`

   **Example:**
   ```
   Password: "MyPass?123"
   Encoded: "MyPass%3F123"
   ```

2. **Save the file**

### Step 5: Start Backend

```bash
cd wazzafak-ai-main

# Start backend
docker compose up --build
```

**Wait for:**
```
✅ Application initialized with database connection
Interview-AI on :8089
```

**Test:**
```bash
# In another terminal
curl http://localhost:8089/v1/health
# Should return: {"ok":true}
```

### Step 6: Configure Android App

1. **Find Your Computer's IP Address:**

   **Windows:**
   ```powershell
   ipconfig
   # Look for "IPv4 Address" (e.g., 192.168.1.100)
   ```

   **Linux/Mac:**
   ```bash
   hostname -I
   ```

2. **Update Network Settings:**

   **File 1:** `Androidfyp-main/android/app/src/main/java/com/example/finalyearproject/app/repository/network/AiRetrofitClient.kt`
   
   ```kotlin
   object Urls {
       const val EMULATOR = "http://10.0.2.2:8089/"
       const val PHYSICAL_DEVICE = "http://YOUR_IP:8089/"  // Replace YOUR_IP
   }
   ```

   **File 2:** `Androidfyp-main/android/app/src/main/java/com/example/finalyearproject/app/interview/InterviewWebSocketClient.kt`
   
   ```kotlin
   private val computerIP = "YOUR_IP"  // Replace YOUR_IP
   private val baseUrl = "ws://$computerIP:8089"
   ```

### Step 7: Build Android App

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

### Step 8: Install on Device

**Option A: Using ADB**

```bash
# Enable USB Debugging on phone first
adb install app/build/outputs/apk/debug/app-debug.apk
```

**Option B: Manual Install**

1. Transfer APK to phone
2. Enable "Install from Unknown Sources"
3. Open APK and install

### Step 9: Test the System

1. **Start Backend** (if not running):
   ```bash
   cd wazzafak-ai-main
   docker compose up -d
   ```

2. **Open App on Phone:**
   - Ensure phone and computer on same WiFi
   - Open the app
   - Test CV upload
   - Test interview session

3. **Verify Connection:**
   - Check backend logs: `docker compose logs -f interview_backend`
   - Look for connection messages

---

## ✅ Verification Checklist

After setup, verify:

- [ ] Docker Desktop is running
- [ ] Backend container is running (`docker compose ps`)
- [ ] Health check works (`curl http://localhost:8089/v1/health`)
- [ ] Supabase database is accessible
- [ ] OpenAI API key is valid
- [ ] Android app network settings configured
- [ ] APK built successfully
- [ ] App installed on device
- [ ] Phone and computer on same WiFi (physical device)
- [ ] App can connect to backend

---

## 🔧 Common Setup Issues

### Issue: Docker not installed

**Solution:**
- Download Docker Desktop: https://www.docker.com/products/docker-desktop
- Install and restart computer
- Verify: `docker --version`

### Issue: Database connection fails

**Solution:**
- Verify Supabase project is active
- Check connection string format
- Ensure password is URL-encoded
- Verify using "Connection pooling" (port 6543)

### Issue: OpenAI API errors

**Solution:**
- Verify API key is correct
- Check account has credits/quota
- Verify payment method is added
- Check API key permissions

### Issue: Android app can't connect

**Solution:**
- Verify backend is running
- Check IP address is correct
- Ensure same WiFi network
- Test in phone browser: `http://YOUR_IP:8089/v1/health`
- Check Windows Firewall allows port 8089

---

## 📚 Next Steps

After successful setup:

1. **Read Documentation:**
   - `README.md` - Main project overview
   - `wazzafak-ai-main/BACKEND_RUN_GUIDE.md` - Detailed backend guide
   - `ARCHITECTURE_DIAGRAMS.md` - System architecture

2. **Test Features:**
   - Upload a CV
   - Start an interview
   - Check feedback generation

3. **Customize:**
   - Adjust AI models in `docker-compose.yml`
   - Modify interview prompts in `main.go`
   - Customize UI in Android app

---

## 🆘 Need Help?

1. Check troubleshooting sections in README files
2. Review backend logs: `docker compose logs interview_backend`
3. Check Android logs: `adb logcat`
4. Verify all prerequisites are installed
5. Ensure configuration is correct

---

**Happy Coding! 🚀**






