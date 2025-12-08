# GitHub Push Guide - Backend Only

## ✅ What to Push: **Just the Backend** (`wazzafak-ai-main` folder)

You only need to push the backend server code for cloud deployment.

---

## 🚀 Quick Steps

### Option 1: Create New Repo for Backend Only (Recommended)

```bash
# Navigate to backend folder
cd wazzafak-ai-main

# Initialize git (if not already)
git init

# Add all files
git add .

# Commit
git commit -m "Backend server ready for deployment"

# Create new repo on GitHub, then:
git remote add origin https://github.com/YOUR_USERNAME/interview-backend.git
git branch -M main
git push -u origin main
```

### Option 2: Push Backend to Existing Repo

If you already have a repo with everything:

```bash
# Navigate to root
cd ..

# Initialize git (if not already)
git init

# Create .gitignore to exclude Android build files
# (See below)

# Add only backend
git add wazzafak-ai-main/

# Commit
git commit -m "Add backend server"

# Push
git remote add origin https://github.com/YOUR_USERNAME/YOUR_REPO.git
git push -u origin main
```

---

## 📁 What Gets Pushed

### ✅ Include (Backend):
- `main.go`
- `cv_parser.go`
- `Dockerfile`
- `go.mod`, `go.sum`
- `render.yaml`, `railway.json`
- All `.md` documentation files
- `env.example` (NOT `.env` with secrets!)

### ❌ Exclude:
- `.env` files (contain secrets!)
- `interview-ai.exe` (build artifacts)
- Android app code (`Androidfyp-main/`)
- Python files (`interview_assistant.py`)
- Build artifacts

---

## 🔒 Security: Before Pushing

**IMPORTANT**: Check these files don't contain secrets:

1. **docker-compose.yml** - Remove API keys before pushing:
   ```yaml
   # Change this:
   - LLM_API_KEY=sk-proj-...
   
   # To this:
   - LLM_API_KEY=${LLM_API_KEY}
   ```

2. **Never push**:
   - `.env` files
   - Real API keys
   - Database passwords (use environment variables)

---

## 📝 Recommended .gitignore (Root Level)

If pushing everything to one repo:

```gitignore
# Backend build artifacts
wazzafak-ai-main/interview-ai
wazzafak-ai-main/interview-ai.exe
wazzafak-ai-main/.env

# Android build artifacts
Androidfyp-main/android/app/build/
Androidfyp-main/android/.gradle/
Androidfyp-main/android/local.properties

# Python
venv/
__pycache__/
*.pyc

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store
Thumbs.db
```

---

## 🎯 Best Practice

**Recommended Structure:**
```
GitHub Repo: interview-backend
├── main.go
├── cv_parser.go
├── Dockerfile
├── go.mod
└── ... (all backend files)
```

**Keep separate:**
- Android app → Separate repo (optional)
- Python scripts → Keep local or separate repo

---

## ✅ After Pushing

1. Go to Render.com
2. Connect your GitHub repo
3. Select the `wazzafak-ai-main` directory (or root if backend-only repo)
4. Deploy!

---

## 🔄 Update Process

After pushing:
```bash
cd wazzafak-ai-main
git add .
git commit -m "Update: description"
git push
```

Render will auto-deploy on push! 🚀

