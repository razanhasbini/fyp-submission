# 🚀 How to Run the Backend and Android App

## 📋 Prerequisites

Before starting, make sure you have:
- ✅ Docker Desktop installed and running
- ✅ Android Studio installed (for building the app)
- ✅ Supabase database password (or connection string)
- ✅ OpenAI API key (already configured)

---

## 🔧 Part 1: Running the Backend

### Step 1: Configure Database Connection

**Option A: Using .env file (Recommended)**

1. Navigate to backend folder:
   ```bash
   cd wazzafak-ai-main
   ```

2. Create `.env` file:
   ```bash
   # On Windows (PowerShell)
   Copy-Item env.example .env
   
   # Or create manually: create a file named .env
   ```

3. Edit `.env` file and add your Supabase connection string:
   ```env
   DATABASE_URL=postgresql://postgres.npeusanizvcyjwsgbhfn:YOUR_PASSWORD@aws-0-eu-north-1.pooler.supabase.com:6543/postgres?sslmode=require
   LLM_API_KEY=YOUR_OPENAI_API_KEY
   USE_MOCK=0
   PORT=8089
   ```

**Option B: Edit docker-compose.yml directly**

Edit `docker-compose.yml` line 21 and replace `[YOUR-PASSWORD]` with your actual password.

### Step 2: Start the Backend

1. Open terminal/PowerShell in `wazzafak-ai-main` folder

2. Start Docker containers:
   ```bash
   docker compose up -d --build
   ```

   This will:
   - Build the Go backend
   - Start the backend container
   - Connect to Supabase database

3. Check if backend is running:
   ```bash
   docker compose ps
   ```

   You should see:
   ```
   interview_backend    Up (healthy)
   ```

### Step 3: Verify Backend is Working

1. Check backend logs:
   ```bash
   docker compose logs interview_backend
   ```

   Look for:
   ```
   ✅ Connected to database: PostgreSQL ...
   Interview-AI on :8089
   ```

2. Test health endpoint:
   ```bash
   # In browser or using curl
   curl http://localhost:8089/v1/health
   ```

   Should return: `{"ok":true}`

### Step 4: Find Your Computer's IP Address

**For Physical Device Testing:**

You need your computer's local IP address so your phone can connect.

**Windows:**
```powershell
ipconfig
```
Look for "IPv4 Address" under your active network adapter (usually WiFi or Ethernet).

**Example:** `192.168.1.100`

**Linux/Mac:**
```bash
hostname -I
# OR
ip route get 8.8.8.8 | awk '{print $7}'
```

---

## 📱 Part 2: Running the Android App

### Step 1: Configure Network Settings

**For Physical Device:**

1. Open Android Studio
2. Navigate to: `Androidfyp-main/android/app/src/main/java/com/example/finalyearproject/app/repository/network/AiRetrofitClient.kt`

3. Find the `Urls` object and update:
   ```kotlin
   object Urls {
       const val EMULATOR = "http://10.0.2.2:8089/"      // For emulator
       const val PHYSICAL_DEVICE = "http://YOUR_COMPUTER_IP:8089/"  // Replace YOUR_COMPUTER_IP
   }
   ```

4. Also update `InterviewWebSocketClient.kt`:
   - Find the `computerIP` variable
   - Set it to your computer's IP address

**For Emulator:**
- Use `10.0.2.2:8089` (already configured by default)

### Step 2: Build the Android App

**Option A: Build APK (Recommended for Testing)**

1. Open terminal in `Androidfyp-main/android` folder:
   ```bash
   cd Androidfyp-main/android
   ```

2. Build debug APK:
   ```bash
   # Windows
   .\gradlew.bat assembleDebug
   
   # Linux/Mac
   ./gradlew assembleDebug
   ```

3. APK location:
   ```
   Androidfyp-main/android/app/build/outputs/apk/debug/app-debug.apk
   ```

**Option B: Run from Android Studio**

1. Open Android Studio
2. Open project: `Androidfyp-main/android`
3. Wait for Gradle sync
4. Connect device or start emulator
5. Click **Run** button (green play icon)

### Step 3: Install APK on Device

**Method 1: Using ADB (Recommended)**

1. Enable USB Debugging on your phone:
   - Settings → About Phone → Tap "Build Number" 7 times
   - Settings → Developer Options → Enable "USB Debugging"

2. Connect phone via USB

3. Install APK:
   ```bash
   adb install app/build/outputs/apk/debug/app-debug.apk
   ```

**Method 2: Transfer and Install Manually**

1. Copy `app-debug.apk` to your phone
2. On phone: Settings → Security → Enable "Install from Unknown Sources"
3. Open the APK file on your phone and install

### Step 4: Configure App for Your Network

**Important:** Make sure your phone and computer are on the **same WiFi network**.

1. Open the app on your phone
2. The app should automatically detect if it's on emulator or physical device
3. If connection fails, check:
   - Phone and computer on same WiFi
   - Backend is running (`docker compose ps`)
   - Firewall allows port 8089
   - IP address is correct in the app code

---

## ✅ Verification Checklist

### Backend:
- [ ] Docker Desktop is running
- [ ] Backend container is up: `docker compose ps`
- [ ] Backend logs show: `✅ Connected to database`
- [ ] Health check works: `curl http://localhost:8089/v1/health`
- [ ] Computer IP address found

### Android App:
- [ ] Network settings configured (IP address set)
- [ ] APK built successfully
- [ ] APK installed on device/emulator
- [ ] Phone and computer on same WiFi (for physical device)
- [ ] App can connect to backend

---

## 🔄 Daily Workflow

### Starting Everything:

1. **Start Backend:**
   ```bash
   cd wazzafak-ai-main
   docker compose up -d
   ```

2. **Verify Backend:**
   ```bash
   docker compose logs interview_backend
   curl http://localhost:8089/v1/health
   ```

3. **Run/Install Android App:**
   - Build APK: `cd Androidfyp-main/android && .\gradlew.bat assembleDebug`
   - Or run from Android Studio

### Stopping Backend:

```bash
cd wazzafak-ai-main
docker compose down
```

---

## 🐛 Troubleshooting

### Backend Won't Start

**Error: "Cannot connect to database"**
- Check your Supabase password is correct
- Verify connection string format
- Check Supabase dashboard - is database accessible?

**Error: "Port 8089 already in use"**
- Stop other services using port 8089
- Or change port in `docker-compose.yml`

### App Can't Connect to Backend

**"Connection Failed" on Physical Device:**
- Verify phone and computer on same WiFi
- Check computer's IP address is correct
- Test: Open `http://YOUR_IP:8089/v1/health` in phone's browser
- Check Windows Firewall allows port 8089

**"Connection Failed" on Emulator:**
- Use `10.0.2.2:8089` (not `localhost`)
- Verify backend is running

### Backend Logs

View logs in real-time:
```bash
docker compose logs -f interview_backend
```

Stop logs: Press `Ctrl+C`

---

## 📚 Quick Reference Commands

```bash
# Backend
cd wazzafak-ai-main
docker compose up -d              # Start backend
docker compose down               # Stop backend
docker compose logs interview_backend  # View logs
docker compose ps                 # Check status

# Android
cd Androidfyp-main/android
.\gradlew.bat assembleDebug       # Build APK (Windows)
./gradlew assembleDebug            # Build APK (Linux/Mac)
adb install app/build/outputs/apk/debug/app-debug.apk  # Install APK
```

---

## 🎯 Next Steps

Once everything is running:
1. Test CV upload feature
2. Test interview session
3. Check data is stored in Supabase
4. Monitor backend logs for any errors

---

## 💡 Tips

- **Keep backend running** while testing the app
- **Check logs** if something doesn't work
- **Use emulator first** to test, then move to physical device
- **Same WiFi network** is required for physical device
- **Backend must be running** before starting the app






