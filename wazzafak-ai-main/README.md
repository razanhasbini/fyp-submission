# 🎤 Interview AI Backend

Go backend server for the Interview AI system, providing CV analysis, live interview sessions, and feedback generation.

## 🚀 Quick Start

### Prerequisites

- Docker Desktop installed and running
- Supabase account (for database)
- OpenAI API key (for AI features)

### Setup

1. **Configure Database Connection:**

   Edit `docker-compose.yml` and update:
   ```yaml
   - DATABASE_URL=postgresql://postgres.YOUR_PROJECT_REF:YOUR_PASSWORD@aws-0-YOUR_REGION.pooler.supabase.com:6543/postgres?sslmode=require
   - LLM_API_KEY=sk-your-openai-api-key-here
   ```

   **Important:** URL-encode special characters in password:
   - `?` → `%3F`
   - `@` → `%40`
   - `#` → `%23`

2. **Start Backend:**

   ```bash
   docker compose up --build
   ```

3. **Verify:**

   ```bash
   curl http://localhost:8089/v1/health
   # Should return: {"ok":true}
   ```

## 📖 Full Documentation

See `BACKEND_RUN_GUIDE.md` for detailed setup instructions.

## 🔧 Configuration

All configuration is in `docker-compose.yml`:

- `DATABASE_URL`: Supabase PostgreSQL connection string
- `LLM_API_KEY`: OpenAI API key
- `USE_MOCK`: Set to `1` for testing without OpenAI
- `PORT`: Server port (default: 8089)

## 📡 API Endpoints

- `GET /v1/health` - Health check
- `POST /v1/cv/upload` - Upload and analyze CV
- `WS /v1/live-interview?user_id=X` - Live interview WebSocket
- `GET /v1/user/scores?user_id=X` - Get user scores
- `GET /v1/user/feedback/cv?user_id=X` - Get CV feedback
- `GET /v1/user/feedback/technical?user_id=X` - Get technical feedback
- `GET /v1/user/feedback/behavioral?user_id=X` - Get behavioral feedback

## 🐛 Troubleshooting

```bash
# View logs
docker compose logs -f interview_backend

# Restart
docker compose restart interview_backend

# Check status
docker compose ps
```

## 📝 License

[Add your license]






