# 🚀 Quick Deploy to Cloud (5 Minutes)

## Fastest Option: Render.com

### Step 1: Push to GitHub (2 min)
```bash
cd wazzafak-ai-main
git init
git add .
git commit -m "Ready for deployment"
git remote add origin https://github.com/YOUR_USERNAME/YOUR_REPO.git
git push -u origin main
```

### Step 2: Deploy on Render (3 min)

1. Go to: **https://render.com** → Sign up (free)

2. Click **"New +"** → **"Web Service"**

3. Connect your GitHub repo

4. Settings:
   - **Name**: `interview-backend`
   - **Environment**: `Docker`
   - **Root Directory**: `wazzafak-ai-main` (or leave empty if repo root)

5. **Environment Variables** (click "Advanced" → "Add Environment Variable"):
   ```
   DATABASE_URL = postgresql://postgres.npeusanizvcyjwsgbhfn:YOUR_PASSWORD@aws-0-eu-north-1.pooler.supabase.com:6543/postgres?sslmode=require
   USE_MOCK = 0
   LLM_API_KEY = sk-proj-YOUR_API_KEY_HERE (Get from https://platform.openai.com/api-keys)
   LLM_BASE_URL = (leave empty)
   EMBEDDING_MODEL = text-embedding-3-small
   CHAT_MODEL = gpt-4o-mini
   EVAL_MODEL = gpt-4o
   PORT = 8089
   ```

6. Click **"Create Web Service"**

7. Wait 5-10 minutes for build

8. **Get your URL**: `https://your-service-name.onrender.com`

### Step 3: Update Android App

Edit `Androidfyp-main/android/app/src/main/java/com/example/finalyearproject/app/utils/NetworkUtils.kt`:

```kotlin
fun configureAiServiceUrl() {
    val isPhysicalDevice = Build.FINGERPRINT.contains("generic").not() &&
            Build.FINGERPRINT.contains("unknown").not() &&
            Build.MODEL.contains("google_sdk").not() &&
            Build.MODEL.contains("Emulator").not() &&
            Build.MODEL.contains("Android SDK built for").not() &&
            Build.MANUFACTURER.contains("Genymotion").not() &&
            Build.BRAND.startsWith("generic").not() &&
            Build.DEVICE.startsWith("generic").not() &&
            Build.PRODUCT.contains("google_sdk").not()
    
    // Use cloud backend for all devices
    AiRetrofitClient.setBaseUrl("https://your-service-name.onrender.com")
    Log.d(TAG, "Using cloud backend: https://your-service-name.onrender.com")
    AiRetrofitClient.reset()
}
```

### Step 4: Test

```bash
curl https://your-service-name.onrender.com/v1/health
# Should return: {"ok":true}
```

---

## Alternative: Railway.app

1. Go to: **https://railway.app** → Sign up
2. **New Project** → **Deploy from GitHub**
3. Select your repo
4. Add same environment variables
5. Deploy!

---

## ⚠️ Important Notes:

- **Free tier sleeps** after 15 min inactivity (Render)
- First request after sleep takes ~30 seconds
- For always-on, upgrade to paid ($7/month on Render)

---

## ✅ Done!

Your backend is now in the cloud! Update your Android app URL and test! 🎉

