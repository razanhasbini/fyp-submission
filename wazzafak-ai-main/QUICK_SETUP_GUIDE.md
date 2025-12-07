# 🚀 Quick Setup Guide - Supabase Connection

## ❌ IMPORTANT: You DON'T Use Google Login Credentials!

When you log into Supabase with Google, that's **only for accessing the dashboard**. You need a **different password** - the database password that Supabase created for your project.

---

## 📍 Step 1: Get Your Database Password from Supabase

### Where to Find It:

1. **Go to**: https://app.supabase.com
2. **Log in** with Google (this is just to access the dashboard)
3. **Click** on your project: `npeusanizvcyjwsgbhfn`
4. **Click** **Settings** (⚙️ gear icon) in left sidebar
5. **Click** **Database** in the settings menu
6. **Scroll down** to **"Connection string"** section
7. **Click** on **"Connection pooling"** tab
8. **Copy** the connection string - it looks like:
   ```
   postgres://postgres.npeusanizvcyjwsgbhfn:YOUR_PASSWORD_HERE@aws-0-us-east-1.pooler.supabase.com:6543/postgres
   ```

### The Password is:
- The part **after** `postgres.npeusanizvcyjwsgbhfn:`
- And **before** `@`

**Example:**
```
postgres://postgres.npeusanizvcyjwsgbhfn:MyPassword123!@aws-0-us-east-1.pooler.supabase.com:6543/postgres
                                    ^^^^^^^^^^^^^^^^
                                    This is your password!
```

---

## 📝 Step 2: Edit docker-compose.yml

### Open the File:
- File: `wazzafak-ai-main/docker-compose.yml`
- Line: **20**

### Find This Line:
```yaml
- DATABASE_URL=${DATABASE_URL:-postgres://postgres.npeusanizvcyjwsgbhfn:[YOUR_SUPABASE_DB_PASSWORD]@aws-0-[REGION].pooler.supabase.com:6543/postgres?sslmode=require}
```

### Replace TWO Things:

1. **Replace `[YOUR_SUPABASE_DB_PASSWORD]`** with your actual password
2. **Replace `[REGION]`** with your region (e.g., `us-east-1`, `eu-west-1`)

### Example:

**BEFORE:**
```yaml
- DATABASE_URL=${DATABASE_URL:-postgres://postgres.npeusanizvcyjwsgbhfn:[YOUR_SUPABASE_DB_PASSWORD]@aws-0-[REGION].pooler.supabase.com:6543/postgres?sslmode=require}
```

**AFTER (with example values):**
```yaml
- DATABASE_URL=${DATABASE_URL:-postgres://postgres.npeusanizvcyjwsgbhfn:MyPassword123!@aws-0-us-east-1.pooler.supabase.com:6543/postgres?sslmode=require}
```

---

## 🎯 Visual Guide: What to Replace

```
postgres://postgres.npeusanizvcyjwsgbhfn:[YOUR_SUPABASE_DB_PASSWORD]@aws-0-[REGION].pooler.supabase.com:6543/postgres?sslmode=require
                                      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^  ^^^^^^^^
                                      Replace with your password      Replace with region
```

**Example:**
```
postgres://postgres.npeusanizvcyjwsgbhfn:MyPassword123!@aws-0-us-east-1.pooler.supabase.com:6543/postgres?sslmode=require
                                      ^^^^^^^^^^^^^^^^  ^^^^^^^^^
                                      Your password     Your region
```

---

## 🔍 How to Find Your Region

The region is usually in the connection string you copied from Supabase. Look for:
- `aws-0-us-east-1` → region is `us-east-1`
- `aws-0-eu-west-1` → region is `eu-west-1`
- `aws-0-ap-southeast-1` → region is `ap-southeast-1`

**Common regions:**
- `us-east-1` (US East - most common)
- `us-west-1` (US West)
- `eu-west-1` (Europe)
- `ap-southeast-1` (Asia Pacific)

---

## ✅ Step 3: Restart Backend

After editing `docker-compose.yml`, restart the backend:

```bash
cd wazzafak-ai-main
docker compose down
docker compose up -d --build
```

---

## 🧪 Step 4: Verify It Works

Check the logs:
```bash
docker compose logs interview_backend
```

You should see:
```
✅ Connected to database: PostgreSQL ...
```

---

## 🆘 Troubleshooting

### "I can't see the password in Supabase"
- The password might be hidden. Look for a **"Reveal"** or **"Show"** button
- Or click **"Reset database password"** to create a new one (copy it immediately!)

### "I don't know my region"
- Check the connection string from Supabase - it shows the region
- Or look at your Supabase project URL

### "Connection failed"
- Make sure you copied the password correctly (no extra spaces)
- Check that your IP is allowed in Supabase (Settings → Database → Connection pooling → Allowed IPs)
- Try using "Allow all IPs" for testing

---

## 📋 Summary Checklist

- [ ] Logged into Supabase dashboard (with Google - just to access)
- [ ] Went to Settings → Database
- [ ] Found Connection string section
- [ ] Copied the connection string
- [ ] Extracted the password (part between `:` and `@`)
- [ ] Found the region (part after `aws-0-` and before `.pooler`)
- [ ] Edited `docker-compose.yml` line 20
- [ ] Replaced `[YOUR_SUPABASE_DB_PASSWORD]` with actual password
- [ ] Replaced `[REGION]` with actual region
- [ ] Restarted backend: `docker compose down && docker compose up -d --build`
- [ ] Verified connection in logs

---

## 💡 Alternative: Use .env File (More Secure)

Instead of editing docker-compose.yml, you can create a `.env` file:

1. Create file: `wazzafak-ai-main/.env`
2. Add this line (replace with your values):
   ```env
   DATABASE_URL=postgres://postgres.npeusanizvcyjwsgbhfn:YOUR_PASSWORD@aws-0-us-east-1.pooler.supabase.com:6543/postgres?sslmode=require
   ```
3. The docker-compose.yml will automatically use this value

---

## 🎯 Still Confused?

If you're stuck:
1. Take a screenshot of your Supabase Database settings page
2. Or tell me what you see in the "Connection string" section
3. I'll help you find the exact values to use!






