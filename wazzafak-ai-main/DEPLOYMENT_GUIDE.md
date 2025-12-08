# Quick Backend Deployment Guide

Deploy your backend to the cloud so you can run it without your laptop!

## Option 1: Render.com (Recommended - Easiest & Free)

### Steps:

1. **Sign up at Render.com** (free tier available)
   - Go to: https://render.com
   - Sign up with GitHub (easiest)

2. **Push your code to GitHub** (if not already):
   ```bash
   git init
   git add .
   git commit -m "Initial commit"
   git remote add origin https://github.com/YOUR_USERNAME/YOUR_REPO.git
   git push -u origin main
   ```

3. **Create New Web Service on Render**:
   - Go to: https://dashboard.render.com
   - Click "New +" → "Web Service"
   - Connect your GitHub repository
   - Select the `wazzafak-ai-main` directory

4. **Configure the service**:
   - **Name**: `interview-backend` (or any name)
   - **Environment**: `Docker`
   - **Region**: Choose closest to you
   - **Branch**: `main` (or your branch)
   - **Root Directory**: `wazzafak-ai-main` (if repo root, leave empty)

5. **Set Environment Variables** (in Render dashboard):
   - `DATABASE_URL` = `postgresql://postgres.npeusanizvcyjwsgbhfn:YOUR_PASSWORD@aws-0-eu-north-1.pooler.supabase.com:6543/postgres?sslmode=require` (Get from Supabase Dashboard)
   - `USE_MOCK` = `0`
   - `LLM_API_KEY` = `sk-proj-YOUR_API_KEY_HERE` (Get from https://platform.openai.com/api-keys)
   - `LLM_BASE_URL` = (leave empty)
   - `EMBEDDING_MODEL` = `text-embedding-3-small`
   - `CHAT_MODEL` = `gpt-4o-mini`
   - `EVAL_MODEL` = `gpt-4o`
   - `PORT` = `8089`

6. **Deploy**:
   - Click "Create Web Service"
   - Wait 5-10 minutes for first build
   - Your backend will be live at: `https://your-service-name.onrender.com`

7. **Update Android App**:
   - Update `NetworkUtils.kt` to use your Render URL:
   ```kotlin
   AiRetrofitClient.setBaseUrl("https://your-service-name.onrender.com")
   ```

---

## Option 2: Railway.app (Also Easy & Free)

### Steps:

1. **Sign up at Railway**:
   - Go to: https://railway.app
   - Sign up with GitHub

2. **Create New Project**:
   - Click "New Project"
   - Select "Deploy from GitHub repo"
   - Choose your repository

3. **Configure Service**:
   - Railway auto-detects Dockerfile
   - Add environment variables (same as Render above)

4. **Deploy**:
   - Railway auto-deploys
   - Get your URL from the dashboard

---

## Option 3: Fly.io (Good for Docker)

### Steps:

1. **Install Fly CLI**:
   ```bash
   # Windows (PowerShell)
   powershell -Command "iwr https://fly.io/install.ps1 -useb | iex"
   ```

2. **Sign up**: https://fly.io

3. **Deploy**:
   ```bash
   fly launch
   # Follow prompts, set environment variables
   ```

---

## Quick Comparison

| Platform | Free Tier | Ease | Best For |
|----------|-----------|------|----------|
| **Render.com** | ✅ Yes | ⭐⭐⭐⭐⭐ | Easiest, best UI |
| **Railway** | ✅ Yes | ⭐⭐⭐⭐ | Auto-deploy, simple |
| **Fly.io** | ✅ Yes | ⭐⭐⭐ | More control |

---

## Important Notes:

1. **Free tiers may sleep** after inactivity (Render: 15 min, Railway: varies)
   - First request after sleep takes ~30 seconds to wake up
   - Consider upgrading to paid plan for always-on

2. **Update Android App URL**:
   - Change from `http://10.0.2.2:8089` to your cloud URL
   - Update in `NetworkUtils.kt`

3. **WebSocket Support**:
   - Render.com: ✅ Supports WebSockets
   - Railway: ✅ Supports WebSockets
   - Fly.io: ✅ Supports WebSockets

4. **Health Check**:
   - Your app has `/v1/health` endpoint
   - Render will use this automatically

---

## After Deployment:

1. Test your backend:
   ```bash
   curl https://your-service-name.onrender.com/v1/health
   ```

2. Update Android app:
   - Change base URL in `NetworkUtils.kt`
   - Rebuild app

3. Test CV upload and interview from your phone!

---

## Troubleshooting:

- **Build fails**: Check Dockerfile and ensure all dependencies are included
- **Connection timeout**: Check environment variables are set correctly
- **WebSocket not working**: Ensure platform supports WebSockets (all above do)

---

**Recommended: Start with Render.com - it's the easiest!** 🚀

