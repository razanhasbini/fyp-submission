# 🔐 Security Guidelines

## ⚠️ Important: Before Committing to GitHub

**NEVER commit sensitive credentials to GitHub!**

### What to Remove/Replace:

1. **In `docker-compose.yml`:**
   - Replace `DATABASE_URL` with placeholder: `postgresql://postgres.YOUR_PROJECT_REF:YOUR_PASSWORD@...`
   - Replace `LLM_API_KEY` with placeholder: `sk-your-openai-api-key-here`

2. **Check for other sensitive files:**
   - `.env` files (should be in `.gitignore`)
   - API keys in code
   - Database passwords
   - Private keys

### Safe Configuration:

Use `docker-compose.example.yml` as a template. Copy it to `docker-compose.yml` and fill in your credentials locally (never commit the filled version).

### Best Practices:

1. **Use Environment Variables:**
   ```yaml
   - DATABASE_URL=${DATABASE_URL}
   - LLM_API_KEY=${LLM_API_KEY}
   ```
   Then create a `.env` file (which is in `.gitignore`)

2. **Use Secrets Management:**
   - Docker secrets
   - AWS Secrets Manager
   - HashiCorp Vault
   - GitHub Secrets (for CI/CD)

3. **Rotate Credentials:**
   - Change passwords regularly
   - Revoke and regenerate API keys
   - Use different keys for dev/prod

### If You Accidentally Committed Secrets:

1. **Immediately rotate all exposed credentials:**
   - Change database password
   - Revoke and regenerate API keys

2. **Remove from Git history:**
   ```bash
   git filter-branch --force --index-filter \
     "git rm --cached --ignore-unmatch wazzafak-ai-main/docker-compose.yml" \
     --prune-empty --tag-name-filter cat -- --all
   ```

3. **Force push (if already pushed):**
   ```bash
   git push origin --force --all
   ```

4. **Consider using git-secrets or similar tools**

---

**Remember:** Security is everyone's responsibility. Always review what you're committing!






