package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	
	"github.com/gorilla/websocket"

	_ "github.com/jackc/pgx/v5/stdlib" 
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)



type App struct {
	DB        *sql.DB
	LLM       *openai.Client 
	EmbModel  string
	ChatModel string
	EvalModel string
	Port      string
	Mock      bool
}

// InterviewState tracks the live interview state for one candidate
type InterviewState struct {
	Turn              int
	Responses         []string
	TechAsked         int
	Field             string
	Subfield          string
	Skills            []string
	Projects          []string
	Experiences       []string
	LastUser          string
	LastQuestion      string
	Asked             []string
	InCandidateQPhase bool
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s is required", k)
	}
	return v
}
func getEnv(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}



func main() {
	_ = godotenv.Load()
	
	dsn := mustEnv("DATABASE_URL")

	
	mock := os.Getenv("USE_MOCK") == "1"
	key := os.Getenv("LLM_API_KEY")
	if !mock && key == "" {
		log.Fatal("LLM_API_KEY is required unless USE_MOCK=1")
	}

	
	port := getEnv("PORT", "8089")
	emb := getEnv("EMBEDDING_MODEL", "text-embedding-3-small")
	chat := getEnv("CHAT_MODEL", "gpt-4o-mini")
	eval := getEnv("EVAL_MODEL", "gpt-4o")

	
	
	
	if strings.Contains(dsn, ":6543") {
		log.Printf("ℹ️ Using Supabase pooler connection (port 6543)")
		log.Printf("ℹ️ Pooler connection is being used to avoid IPv6 issues")
		// Keep using pooler connection - don't switch to direct connection
		// The pooler works fine with prepareThreshold=0
		
		
		
		// Keep using pooler connection - it works fine with prepareThreshold=0
		// This avoids IPv6 connection issues on Windows Docker
	}
	
	
	
	
	// Build connection string parameters in correct order
	// Check if connection string already has query parameters
	hasParams := strings.Contains(dsn, "?")
	
	if !strings.Contains(dsn, "prepareThreshold") {
		if hasParams {
			dsn += "&prepareThreshold=0"
		} else {
			dsn += "?prepareThreshold=0"
			hasParams = true
		}
	}
	
	if !strings.Contains(dsn, "sslmode") {
		if hasParams {
			dsn += "&sslmode=require"
		} else {
			dsn += "?sslmode=require"
			hasParams = true
		}
	}
	
	if !strings.Contains(dsn, "statement_cache_mode") {
		if hasParams {
			dsn += "&statement_cache_mode=describe"
		} else {
			dsn += "?statement_cache_mode=describe"
		}
	}
	log.Printf("🔗 Database connection string configured (prepared statements DISABLED via prepareThreshold=0, sslmode=require)")
	
	// Log connection details (without password)
	dsnForLog := dsn
	if strings.Contains(dsnForLog, "@") {
		parts := strings.Split(dsnForLog, "@")
		if len(parts) == 2 {
			// Hide password in logs
			if strings.Contains(parts[0], ":") {
				userPass := strings.Split(parts[0], ":")
				if len(userPass) >= 2 {
					dsnForLog = userPass[0] + ":***@" + parts[1]
				}
			}
		}
	}
	log.Printf("🔗 Connecting to: %s", dsnForLog)
	
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to open database connection: %v\n💡 Check: Connection string format, network connectivity, Supabase project status", err)
	}
	
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)
	
	// Set connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	log.Printf("🔄 Testing database connection (timeout: 30s)...")
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf(`❌ Database connection failed: %v

🔍 TROUBLESHOOTING STEPS:
1. Check Supabase Dashboard:
   - Go to: https://app.supabase.com
   - Verify project "npeusanizvcyjwsgbhfn" is active (not paused)
   - Check Settings > Database > Connection string
   
2. Verify Connection String:
   - Password should be URL-encoded (? → %%3F, @ → %%40)
   - Format: postgresql://postgres.[project-ref]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres
   - Current format: postgresql://postgres.npeusanizvcyjwsgbhfn:***@aws-0-eu-north-1.pooler.supabase.com:6543/postgres
   
3. Check Network/Firewall:
   - Ensure port 6543 is not blocked
   - Try direct connection (port 5432) if pooler fails
   
4. Verify Password:
   - Get password from: Supabase Dashboard > Settings > Database
   - Reset password if needed
   - URL-encode special characters

5. Test Connection:
   - Try connecting via Supabase SQL Editor
   - Check if project is paused or deleted`, err)
	}
	
	log.Printf("✅ Database ping successful")
	
	// Test query with timeout
	testCtx, testCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer testCancel()
	
	var version string
	if err := db.QueryRowContext(testCtx, "SELECT version()").Scan(&version); err != nil {
		log.Printf("⚠️ Database test query failed (non-fatal): %v", err)
		log.Printf("ℹ️ Continuing anyway - connection may still be functional")
		log.Printf("💡 This might indicate prepared statement cache issues - queries may still work")
	} else {
		log.Printf("✅ Connected to database: %s", version)
	}

	
	var llm *openai.Client
	if !mock {
		cfg := openai.DefaultConfig(key)
		if base := os.Getenv("LLM_BASE_URL"); base != "" {
			cfg.BaseURL = base 
		}
		llm = openai.NewClientWithConfig(cfg)
	} else {
		log.Println("** MOCK MODE enabled: no calls to OpenAI **")
	}

	app := &App{
		DB:        db,
		LLM:       llm,
		EmbModel:  emb,
		ChatModel: chat,
		EvalModel: eval,
		Port:      port,
		Mock:      mock,
	}

	
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	})
	mux.HandleFunc("/v1/kb/seed", app.handleSeedKB)
	mux.HandleFunc("/v1/cv/ingest", app.handleIngestCV)
	mux.HandleFunc("/v1/next-question", app.handleNextQ)
	mux.HandleFunc("/v1/evaluate", app.handleEvaluate)
	mux.HandleFunc("/v1/upload-cv", app.handleUploadCV)
	mux.HandleFunc("/v1/cv/upload", app.handleUploadCV)

	mux.HandleFunc("/v1/live-interview", app.handleLiveInterview)
	mux.HandleFunc("/v1/feedback", app.handleFeedback)
	mux.HandleFunc("/v1/session/start", app.handleStartSession)
	mux.HandleFunc("/v1/session/end", app.handleEndSession)
	mux.HandleFunc("/v1/user/scores", app.handleGetUserScores)
	mux.HandleFunc("/v1/user/feedback/cv", app.handleGetCvFeedback)
	mux.HandleFunc("/v1/user/feedback/technical", app.handleGetTechnicalFeedback)
	mux.HandleFunc("/v1/user/feedback/behavioral", app.handleGetBehavioralFeedback)

	
	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		mux.ServeHTTP(w, r)
	})

	
	s := &http.Server{
		Addr:              ":" + port,
		Handler:           logReq(corsHandler),
		ReadHeaderTimeout: 60 * time.Second,  
		ReadTimeout:       15 * time.Minute,  
		WriteTimeout:      15 * time.Minute,  
		IdleTimeout:       300 * time.Second,  
		MaxHeaderBytes:    1 << 20,           
	}

	log.Printf("Interview-AI on :%s", port)
	log.Fatal(s.ListenAndServe())
}

func logReq(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}



func (a *App) embed(ctx context.Context, texts []string) ([][]float32, error) {
	
	if a.LLM == nil {
		out := make([][]float32, len(texts))
		for i, t := range texts {
			out[i] = str2vec(t)
		}
		return out, nil
	}

	
	resp, err := a.LLM.CreateEmbeddings(ctx, openai.EmbeddingRequestStrings{
		Model: openai.EmbeddingModel(a.EmbModel),
		Input: texts,
	})
	if err != nil {
		return nil, err
	}
	out := make([][]float32, len(resp.Data))
	for i, d := range resp.Data {
		out[i] = d.Embedding
	}
	return out, nil
}

func str2vec(s string) []float32 {
	d := 256
	v := make([]float32, d)
	var h uint64 = 1469598103934665603
	for _, r := range s {
		h ^= uint64(r)
		h *= 1099511628211
		i := int(h % uint64(d))
		v[i] += 1
	}
	
	var sum float64
	for _, x := range v {
		sum += float64(x * x)
	}
	if sum > 0 {
		n := float32(1.0 / math.Sqrt(sum))
		for i := range v {
			v[i] *= n
		}
	}
	return v
}

func cosSim(a, b []float32) float64 {
	var dot, na, nb float64
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	den := math.Sqrt(na) * math.Sqrt(nb)
	if den == 0 {
		return 0
	}
	return dot / den
}

type Item struct {
	Text string
	Emb  []float32
}

func topKByCos(items []Item, q []float32, k int) []Item {
	type scored struct {
		it Item
		s  float64
	}
	scores := make([]scored, 0, len(items))
	for _, it := range items {
		scores = append(scores, scored{it: it, s: cosSim(it.Emb, q)})
	}
	
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].s > scores[i].s {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
	if k > len(scores) {
		k = len(scores)
	}
	out := make([]Item, k)
	for i := 0; i < k; i++ {
		out[i] = scores[i].it
	}
	return out
}

func toJSON(v []float32) []byte {
	b, _ := json.Marshal(v)
	return b
}
func parseEmb(b []byte) []float32 {
	if len(b) == 0 {
		return nil
	}
	var f []float32
	if err := json.Unmarshal(b, &f); err != nil {
		return nil
	}
	return f
}

func chunkText(s string, size, overlap int) []string {
	if size <= overlap {
		overlap = 0
	}
	var out []string
	for i := 0; i < len(s); i += size - overlap {
		end := i + size
		if end > len(s) {
			end = len(s)
		}
		out = append(out, s[i:end])
		if end == len(s) {
			break
		}
	}
	return out
}
func join(snips []string, max int) string {
	if len(snips) > max {
		snips = snips[:max]
	}
	return strings.Join(snips, "\n---\n")
}
func extractJSON(s string) string {
	i := strings.Index(s, "{")
	j := strings.LastIndex(s, "}")
	if i >= 0 && j > i {
		return s[i : j+1]
	}
	b, _ := json.Marshal(s)
	return `{"error":"model did not return JSON","raw":` + string(b) + `}`
}



type seedReq struct {
	Domain string   `json:"domain"`
	Items  []string `json:"items"`
}

func (a *App) handleSeedKB(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var in seedReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(in.Domain) == "" || len(in.Items) == 0 {
		http.Error(w, "domain+items required", http.StatusBadRequest)
		return
	}

	embs, err := a.embed(r.Context(), in.Items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tx, err := a.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO ai.kb_articles(domain,title,text,embedding_json) VALUES ($1,$2,$3,$4::jsonb)`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	for i, t := range in.Items {
		if _, err := stmt.Exec(in.Domain, fmt.Sprintf("seed-%d", i+1), t, string(toJSON(embs[i]))); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"ok":true,"inserted":%d}`, len(in.Items))))
}

type ingestReq struct {
	UserID int64  `json:"user_id"`
	Text   string `json:"text"`
}

func (a *App) handleIngestCV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var in ingestReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if in.UserID <= 0 || strings.TrimSpace(in.Text) == "" {
		http.Error(w, "user_id+text required", http.StatusBadRequest)
		return
	}

	tx, err := a.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var cvID int64
	if err := tx.QueryRow(`INSERT INTO ai.cv_documents(user_id,text) VALUES ($1,$2) RETURNING id`, in.UserID, in.Text).Scan(&cvID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	chunks := chunkText(in.Text, 4000, 400)
	embs, err := a.embed(r.Context(), chunks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stmt, err := tx.Prepare(`INSERT INTO ai.cv_chunks(cv_id,chunk_text,embedding_json,ord) VALUES ($1,$2,$3::jsonb,$4)`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	for i := range chunks {
		if _, err := stmt.Exec(cvID, chunks[i], string(toJSON(embs[i])), i); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"ok":true,"cv_id":%d,"chunks":%d}`, cvID, len(chunks))))
}

type nextQReq struct {
	UserID     int64  `json:"user_id"`
	Domain     string `json:"domain"`
	Difficulty string `json:"difficulty"`
}
type nextQResp struct {
	Question string `json:"question"`
}

func (a *App) handleNextQ(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var in nextQReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if in.UserID <= 0 || in.Domain == "" {
		http.Error(w, "user_id+domain required", http.StatusBadRequest)
		return
	}

	// Create context with timeout for the entire request (60 seconds)
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	log.Printf("🔍 Getting next question for user %d, domain: %s, difficulty: %s (timeout: 60s)", in.UserID, in.Domain, in.Difficulty)

	
	q := "next interview question for " + in.Domain
	qEmb, err := a.embed(ctx, []string{q})
	if err != nil || len(qEmb) == 0 {
		log.Printf("❌ Embedding error: %v", err)
		http.Error(w, "embedding error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	queryEmb := qEmb[0]

	// Use unique query ID to avoid prepared statement cache issues
	uniqueCVQueryID := time.Now().UnixNano()
	cvQuery := fmt.Sprintf(`
	  SELECT c.chunk_text, c.embedding_json
	  FROM ai.cv_chunks c
	  JOIN ai.cv_documents d ON d.id = c.cv_id
	  WHERE d.user_id = $1
	  ORDER BY c.ord
	  LIMIT 200
	  -- Unique CV query ID: %d`, uniqueCVQueryID)
	
	cvRows, err := a.DB.QueryContext(ctx, cvQuery, in.UserID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("❌ Timeout querying CV chunks for user %d", in.UserID)
			http.Error(w, "Request timeout - database query took too long", http.StatusRequestTimeout)
			return
		}
		log.Printf("❌ Database error querying CV chunks: %v", err)
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer cvRows.Close()

	var cvItems []Item
	for cvRows.Next() {
		var txt string
		var embBytes []byte
		if err := cvRows.Scan(&txt, &embBytes); err == nil {
			cvItems = append(cvItems, Item{Text: txt, Emb: parseEmb(embBytes)})
		}
	}

	// Use unique query ID to avoid prepared statement cache issues
	uniqueKBQueryID := time.Now().UnixNano()
	kbQuery := fmt.Sprintf(`
	  SELECT text, embedding_json
	  FROM ai.kb_articles
	  WHERE domain = $1
	  LIMIT 200
	  -- Unique KB query ID: %d`, uniqueKBQueryID)
	
	kbRows, err := a.DB.QueryContext(ctx, kbQuery, in.Domain)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("❌ Timeout querying KB articles for domain %s", in.Domain)
			http.Error(w, "Request timeout - database query took too long", http.StatusRequestTimeout)
			return
		}
		log.Printf("❌ Database error querying KB articles: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer kbRows.Close()

	var kbItems []Item
	for kbRows.Next() {
		var txt string
		var embBytes []byte
		if err := kbRows.Scan(&txt, &embBytes); err == nil {
			kbItems = append(kbItems, Item{Text: txt, Emb: parseEmb(embBytes)})
		}
	}

	ctxItems := append(topKByCos(cvItems, queryEmb, 3), topKByCos(kbItems, queryEmb, 3)...)
	var snips []string
	for _, it := range ctxItems {
		snips = append(snips, it.Text)
	}
	contextBlock := join(snips, 6)

	
	if a.LLM == nil {
		q := "Based on your background, describe how you would manage the department if you were a team lead" + in.Domain + "."
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(nextQResp{Question: q})
		return
	}

	
	sys := fmt.Sprintf(`You are a strict but fair interviewer in the %s domain.
Ask ONE question (<=2 sentences). Use CONTEXT if needed. Do not reveal answers.`, in.Domain)
	user := "CONTEXT:\n" + contextBlock + "\n---\nDifficulty: " + in.Difficulty + "\nReturn exactly one interview question now."

	resp, err := a.LLM.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       a.ChatModel,
		Temperature: 0.3,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sys},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
	})
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("❌ Timeout calling LLM for question generation")
			http.Error(w, "Request timeout - LLM call took too long", http.StatusRequestTimeout)
			return
		}
		log.Printf("❌ LLM error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(resp.Choices) == 0 {
		log.Printf("❌ LLM returned no choices")
		http.Error(w, "LLM returned no response", http.StatusInternalServerError)
		return
	}
	question := strings.TrimSpace(resp.Choices[0].Message.Content)
	if question == "" {
		log.Printf("⚠️ LLM returned empty question, using fallback")
		question = fmt.Sprintf("Tell me about your experience with %s.", in.Domain)
	}

	log.Printf("✅ Generated question for user %d: %s", in.UserID, question[:min(50, len(question))])
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(nextQResp{Question: question})
}

type evalReq struct {
	SessionID string `json:"session_id"`
	UserID    int64  `json:"user_id"`
	Domain    string `json:"domain"`
	Question  string `json:"question"`
	Answer    string `json:"answer"`
}


func (a *App) evaluateAnswerInternal(sessionID string, userID int64, question, answer, domain string) map[string]interface{} {
	
	
	
	if a.Mock || a.LLM == nil {
		overall := 70.0
		strengths := []string{"Good response"}
		weaknesses := []string{}
		
		answerLen := len(answer)
		if answerLen > 50 {
			overall += 10
			strengths = append(strengths, "Detailed response")
		} else if answerLen < 20 {
			overall -= 10
			weaknesses = append(weaknesses, "Answer too brief")
		}
		
		technicalScore := overall * 0.4
		communicationScore := overall * 0.3
		confidenceScore := overall * 0.3
		
		result := map[string]interface{}{
			"overall":            overall,
			"technical_score":   technicalScore,
			"communication_score": communicationScore,
			"confidence_score":  confidenceScore,
			"strengths":          strengths,
			"weaknesses":         weaknesses,
		}
		
		
		resultJSON, _ := json.Marshal(result)
		insertQuery := fmt.Sprintf(`
			INSERT INTO ai.evaluations(session_id, question, answer, result_json, created_at)
			VALUES ('%s', '%s', '%s', '%s'::jsonb, NOW())
		`, sessionID, strings.ReplaceAll(question, "'", "''"), strings.ReplaceAll(answer, "'", "''"), 
			strings.ReplaceAll(string(resultJSON), "'", "''"))
		_, _ = a.DB.Exec(insertQuery)
		return result
	}
	
	
	ctx := context.Background()
	
	
	ctxBlock := ""
	if domain != "" {
		uniqueEvalKBQueryID := time.Now().UnixNano()
		kbQuery := fmt.Sprintf(`
			SELECT text FROM ai.kb_articles WHERE domain = $1 LIMIT 6
			-- Unique eval KB query ID: %d`, uniqueEvalKBQueryID)
		kbRows, err := a.DB.QueryContext(ctx, kbQuery, domain)
		if err == nil {
			defer kbRows.Close()
			var snippets []string
			for kbRows.Next() {
				var txt string
				if err := kbRows.Scan(&txt); err == nil {
					snippets = append(snippets, txt)
				}
			}
			if len(snippets) > 0 {
				ctxBlock = strings.Join(snippets, "\n\n")
			}
		}
	}
	
	
	sys := "You are an interview evaluator. Output STRICT JSON only, no prose."
	user := fmt.Sprintf(`QUESTION: %s
ANSWER: %s
CONTEXT:
%s

RUBRIC:
{
  "criteria": [
    {"name":"Relevance","weight":0.25},
    {"name":"Structure/STAR","weight":0.20},
    {"name":"Technical Correctness","weight":0.30},
    {"name":"Clarity & Conciseness","weight":0.15},
    {"name":"Confidence & Tone","weight":0.10}
  ],
  "scale":{"min":1,"max":5}
}

Return JSON:
{
  "scores":[{"name":"Relevance","score":1-5},{"name":"Structure/STAR","score":1-5},{"name":"Technical Correctness","score":1-5},{"name":"Clarity & Conciseness","score":1-5},{"name":"Confidence & Tone","score":1-5}],
  "overall":1-100,
  "strengths":["..."],
  "weaknesses":["..."],
  "action_items":["..."],
  "suggested_model_answer":"..."
}`, question, answer, ctxBlock)
	
	resp, err := a.LLM.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       a.EvalModel,
		Temperature: 0.0,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sys},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
	})
	
	if err != nil {
		log.Printf(" Evaluation error: %v, using fallback", err)
		
		overall := 70.0
		technicalScore := overall * 0.4
		communicationScore := overall * 0.3
		confidenceScore := overall * 0.3
		result := map[string]interface{}{
			"overall":            overall,
			"technical_score":   technicalScore,
			"communication_score": communicationScore,
			"confidence_score":  confidenceScore,
			"strengths":          []string{},
			"weaknesses":         []string{},
		}
		resultJSON, _ := json.Marshal(result)
		insertQuery := fmt.Sprintf(`
			INSERT INTO ai.evaluations(session_id, question, answer, result_json, created_at)
			VALUES ('%s', '%s', '%s', '%s'::jsonb, NOW())
		`, sessionID, strings.ReplaceAll(question, "'", "''"), strings.ReplaceAll(answer, "'", "''"), 
			strings.ReplaceAll(string(resultJSON), "'", "''"))
		_, _ = a.DB.Exec(insertQuery)
		return result
	}
	
	raw := strings.TrimSpace(resp.Choices[0].Message.Content)
	jsonText := extractJSON(raw)
	
	
	var evalData struct {
		Overall   float64 `json:"overall"`
		Scores    []struct {
			Name  string  `json:"name"`
			Score float64 `json:"score"`
		} `json:"scores"`
		Strengths  []string `json:"strengths"`
		Weaknesses []string `json:"weaknesses"`
	}
	
	if err := json.Unmarshal([]byte(jsonText), &evalData); err != nil {
		log.Printf(" Failed to parse evaluation JSON: %v", err)
		
		overall := 70.0
		result := map[string]interface{}{
			"overall":            overall,
			"technical_score":   overall * 0.4,
			"communication_score": overall * 0.3,
			"confidence_score":  overall * 0.3,
			"strengths":          []string{},
			"weaknesses":         []string{},
		}
		resultJSON, _ := json.Marshal(result)
		insertQuery := fmt.Sprintf(`
			INSERT INTO ai.evaluations(session_id, question, answer, result_json, created_at)
			VALUES ('%s', '%s', '%s', '%s'::jsonb, NOW())
		`, sessionID, strings.ReplaceAll(question, "'", "''"), strings.ReplaceAll(answer, "'", "''"), 
			strings.ReplaceAll(string(resultJSON), "'", "''"))
		_, _ = a.DB.Exec(insertQuery)
		return result
	}
	
	
	var technicalSum, technicalCount float64
	var communicationScore, confidenceScore float64
	
	for _, score := range evalData.Scores {
		nameLower := strings.ToLower(score.Name)
		if strings.Contains(nameLower, "technical") || 
		   strings.Contains(nameLower, "relevance") ||
		   strings.Contains(nameLower, "structure") ||
		   strings.Contains(nameLower, "correctness") {
			technicalSum += score.Score
			technicalCount++
		}
		if strings.Contains(nameLower, "communication") || 
		   strings.Contains(nameLower, "clarity") {
			communicationScore = score.Score * 20 
		}
		if strings.Contains(nameLower, "confidence") || 
		   strings.Contains(nameLower, "tone") {
			confidenceScore = score.Score * 20 
		}
	}
	
	technicalScore := evalData.Overall * 0.4 
	if technicalCount > 0 {
		technicalScore = (technicalSum / technicalCount) * 20 
	}
	if communicationScore == 0 {
		communicationScore = evalData.Overall * 0.3
	}
	if confidenceScore == 0 {
		confidenceScore = evalData.Overall * 0.3
	}
	
	result := map[string]interface{}{
		"overall":            evalData.Overall,
		"technical_score":   technicalScore,
		"communication_score": communicationScore,
		"confidence_score":  confidenceScore,
		"strengths":          evalData.Strengths,
		"weaknesses":         evalData.Weaknesses,
	}
	
	
	insertQuery := fmt.Sprintf(`
		INSERT INTO ai.evaluations(session_id, question, answer, result_json, created_at)
		VALUES ('%s', '%s', '%s', '%s'::jsonb, NOW())
	`, sessionID, strings.ReplaceAll(question, "'", "''"), strings.ReplaceAll(answer, "'", "''"), 
		strings.ReplaceAll(jsonText, "'", "''"))
	_, _ = a.DB.Exec(insertQuery)
	
	return result
}

func (a *App) handleEvaluate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var in evalReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if in.UserID <= 0 || in.Domain == "" || in.Question == "" || in.Answer == "" {
		http.Error(w, "user_id,domain,question,answer required", http.StatusBadRequest)
		return
	}

	qaEmb, err := a.embed(r.Context(), []string{in.Question + " " + in.Answer})
	if err != nil || len(qaEmb) == 0 {
		http.Error(w, "embedding error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	queryEmb := qaEmb[0]

	// Use unique query ID to avoid prepared statement cache issues
	uniqueEvalQueryID := time.Now().UnixNano()
	kbQuery := fmt.Sprintf(`
	  SELECT text, embedding_json
	  FROM ai.kb_articles
	  WHERE domain = $1
	  LIMIT 200
	  -- Unique eval query ID: %d`, uniqueEvalQueryID)
	
	ctx := r.Context()
	kbRows, err := a.DB.QueryContext(ctx, kbQuery, in.Domain)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer kbRows.Close()

	var kbItems []Item
	for kbRows.Next() {
		var txt string
		var embBytes []byte
		if err := kbRows.Scan(&txt, &embBytes); err == nil {
			kbItems = append(kbItems, Item{Text: txt, Emb: parseEmb(embBytes)})
		}
	}
	ctxItems := topKByCos(kbItems, queryEmb, 6)
	var snips []string
	for _, it := range ctxItems {
		snips = append(snips, it.Text)
	}
	ctxBlock := join(snips, 6)

	
	if a.LLM == nil {
		mock := `{
		  "scores":[
		    {"name":"Relevance","score":4},
		    {"name":"Structure/STAR","score":3},
		    {"name":"Technical Correctness","score":4},
		    {"name":"Clarity & Conciseness","score":3},
		    {"name":"Confidence & Tone","score":4}
		  ],
		  "overall":78,
		  "strengths":["Good coverage of key concepts","Clear tech stack"],
		  "weaknesses":["Missing concrete metrics","Shallow on edge cases"],
		  "action_items":["Add latency/error budgets","Describe pagination and auth flows"],
		  "suggested_model_answer":"Define endpoints and resources, use proper HTTP methods/status codes, validate inputs, context timeouts, structured logs, and defensive DB access."
		}`
		_, _ = a.DB.Exec(
			`INSERT INTO ai.evaluations(session_id,question,answer,result_json) VALUES ($1,$2,$3,$4::jsonb)`,
			in.SessionID, in.Question, in.Answer, mock,
		)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mock))
		return
	}

	
	sys := "You are an interview evaluator. Output STRICT JSON only, no prose."
	user := fmt.Sprintf(`QUESTION: %s
ANSWER: %s
CONTEXT:
%s

RUBRIC:
{
  "criteria": [
    {"name":"Relevance","weight":0.25},
    {"name":"Structure/STAR","weight":0.20},
    {"name":"Technical Correctness","weight":0.30},
    {"name":"Clarity & Conciseness","weight":0.15},
    {"name":"Confidence & Tone","weight":0.10}
  ],
  "scale":{"min":1,"max":5}
}

Return JSON:
{
  "scores":[{"name":"Relevance","score":1-5},{"name":"Structure/STAR","score":1-5},{"name":"Technical Correctness","score":1-5},{"name":"Clarity & Conciseness","score":1-5},{"name":"Confidence & Tone","score":1-5}],
  "overall":1-100,
  "strengths":["..."],
  "weaknesses":["..."],
  "action_items":["..."],
  "suggested_model_answer":"..."
}`, in.Question, in.Answer, ctxBlock)

	resp, err := a.LLM.CreateChatCompletion(r.Context(), openai.ChatCompletionRequest{
		Model:       a.EvalModel,
		Temperature: 0.0,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sys},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	raw := strings.TrimSpace(resp.Choices[0].Message.Content)
	jsonText := extractJSON(raw)

	_, _ = a.DB.Exec(
		`INSERT INTO ai.evaluations(session_id,question,answer,result_json) VALUES ($1,$2,$3,$4::jsonb)`,
		in.SessionID, in.Question, in.Answer, jsonText,
	)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonText))
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	
	EnableCompression: true,
}


func isConfused(msg string) bool {
	m := strings.ToLower(msg)
	confusionIndicators := []string{
		"don't understand", "dont understand", "don't understand",
		"i'm confused", "im confused",
		"not sure what you mean",
		"what do you mean",
		"can you rephrase",
		"can you say it differently",
		"you are repeating",
		"this is repeated",
		"same question",
	}
	for _, c := range confusionIndicators {
		if strings.Contains(m, c) {
			return true
		}
	}
	return false
}

func isEmotional(msg string) bool {
	m := strings.ToLower(msg)
	return strings.Contains(m, "anxious") ||
		strings.Contains(m, "anxiety") ||
		strings.Contains(m, "nervous") ||
		strings.Contains(m, "stressed") ||
		strings.Contains(m, "worried") ||
		strings.Contains(m, "sad") ||
		strings.Contains(m, "frustrated") ||
		strings.Contains(m, "overwhelmed") ||
		strings.Contains(m, "not confident") ||
		strings.Contains(m, "i'm scared") ||
		strings.Contains(m, "im scared") ||
		strings.Contains(m, "i feel bad")
}

func isEmotionalEnding(msg string) bool {
	m := strings.ToLower(msg)
	return strings.Contains(m, "hope") ||
		strings.Contains(m, "nervous") ||
		strings.Contains(m, "worried") ||
		strings.Contains(m, "did i do well") ||
		strings.Contains(m, "not confident") ||
		strings.Contains(m, "anxious")
}

func alreadyAsked(st *InterviewState, q string) bool {
	q = strings.TrimSpace(strings.ToLower(q))
	for _, prev := range st.Asked {
		if strings.TrimSpace(strings.ToLower(prev)) == q {
			return true
		}
	}
	return false
}

func tooSimilar(newQ string, prevQs []string) bool {
	newQ = strings.ToLower(newQ)
	for _, old := range prevQs {
		old = strings.ToLower(old)
		if strings.HasPrefix(newQ, strings.Split(old, " ")[0]) {
			return true
		}
		if strings.Contains(newQ, "explain") && strings.Contains(old, "explain") {
			return true
		}
		if strings.Contains(newQ, "describe") && strings.Contains(old, "describe") {
			return true
		}
		if strings.Contains(newQ, "how did you") && strings.Contains(old, "how did you") {
			return true
		}
		if strings.Contains(newQ, "project") && strings.Contains(old, "project") {
			return true
		}
		wordsNew := strings.Fields(newQ)
		wordsOld := strings.Fields(old)
		matches := 0
		for _, wn := range wordsNew {
			if len(wn) <= 4 {
				continue
			}
			for _, wo := range wordsOld {
				if wn == wo {
					matches++
					if matches >= 4 {
						return true
					}
				}
			}
		}
	}
	return false
}

func pickRelevantProject(st *InterviewState, qType string) string {
	if len(st.Projects) == 0 {
		return ""
	}
	if qType == "technical" {
		for _, p := range st.Projects {
			lp := strings.ToLower(p)
			if strings.Contains(lp, "api") ||
				strings.Contains(lp, "backend") ||
				strings.Contains(lp, "integration") ||
				strings.Contains(lp, "real-time") {
				return p
			}
		}
	}
	longest := ""
	for _, p := range st.Projects {
		if len(p) > len(longest) {
			longest = p
		}
	}
	return longest
}

func lastN(arr []string, n int) []string {
	if len(arr) <= n {
		return arr
	}
	return arr[len(arr)-n:]
}

// Extract CV structure using LLM
func (a *App) extractStructure(ctx context.Context, cv string) (field, sub string, skills, projects, exps []string) {
	if a.LLM == nil {
		return "", "", nil, nil, nil
	}
	sys := `You are a precise CV parser. Extract structured information from the CV and return STRICT JSON.
Rules:
- Field must be one of: "Software Engineering", "Business", "Mechanical Engineering", "Graphic Design", "Marketing", "Other"
- Subfield must be a short phrase (e.g., Frontend, Backend, Finance, UI/UX, etc.)
- Skills: list of individual skills (not sentences)
- Projects: 1–2 sentence summaries or titles
- Experiences: short role@company descriptions
No commentary outside JSON.`
	user := fmt.Sprintf(`Extract structured info from this CV.
CV:
---
%s
---
Return JSON:
{
  "field": "Software Engineering | Business | Mechanical Engineering | Graphic Design | Marketing | Other",
  "subfield": "short phrase",
  "skills": ["skill1","skill2",...],
  "projects": ["project summary or title", "..."],
  "experiences": ["role@company with brief context", "..."]
}`, cv)

	ctx2, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	resp, err := a.LLM.CreateChatCompletion(ctx2, openai.ChatCompletionRequest{
		Model:       a.ChatModel,
		Temperature: 0.1,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sys},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		MaxTokens: 500,
	})
	if err != nil || len(resp.Choices) == 0 {
		return "", "", nil, nil, nil
	}
	js := extractJSON(strings.TrimSpace(resp.Choices[0].Message.Content))
	var out struct {
		Field       string   `json:"field"`
		Subfield    string   `json:"subfield"`
		Skills      []string `json:"skills"`
		Projects    []string `json:"projects"`
		Experiences []string `json:"experiences"`
	}
	_ = json.Unmarshal([]byte(js), &out)
	return strings.TrimSpace(out.Field), strings.TrimSpace(out.Subfield), out.Skills, out.Projects, out.Experiences
}

// First message: greeting + explain structure + ask how they feel
func (a *App) firstQuestion(ctx context.Context, st *InterviewState, cv string) string {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ Panic in firstQuestion: %v", r)
		}
	}()
	
	if a.LLM == nil {
		return "Hi there, I'm your interviewer today. We'll go through some general questions and then some based on your CV. Before we dive in, how are you feeling today?"
	}
	sys := `You are a professional recruiter conducting a live video interview.
Task: Generate a natural spoken introduction in 2–3 sentences.
Content rules:
- Sentence 1: Introduce yourself and your role.
- Sentence 2: Briefly explain the interview structure (general questions + CV-based questions).
- Sentence 3: Politely ask how they are feeling before starting.
Tone: Warm, human, confident, not robotic.
Do NOT: Ask any interview or CV questions yet. Use emojis or lists.`
	user := "Generate the spoken introduction now."

	ctx2, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	resp, err := a.LLM.CreateChatCompletion(ctx2, openai.ChatCompletionRequest{
		Model:       a.ChatModel,
		Temperature: 0.4,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sys},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		MaxTokens: 200,
	})
	if err != nil {
		log.Printf("❌ Error in firstQuestion LLM call: %v", err)
		return "Hi, I'm your interviewer today. We'll start with a few general questions and then some based on your CV. How are you feeling before we begin?"
	}
	if len(resp.Choices) == 0 {
		log.Println("⚠️ firstQuestion: LLM returned no choices")
		return "Hi, I'm your interviewer today. We'll start with a few general questions and then some based on your CV. How are you feeling before we begin?"
	}
	result := strings.TrimSpace(resp.Choices[0].Message.Content)
	if result == "" {
		log.Println("⚠️ firstQuestion: LLM returned empty content")
		return "Hi, I'm your interviewer today. We'll start with a few general questions and then some based on your CV. How are you feeling before we begin?"
	}
	return result
}

// Decide question type strictly by turn number
func decideType(st *InterviewState) string {
	switch st.Turn {
	case 1:
		return "personal_intro"
	case 2:
		return "motivation"
	case 3, 4, 5:
		return "technical"
	case 6:
		return "scenario"
	case 7:
		return "experience"
	case 8:
		return "behavioral"
	default:
		if st.Turn%2 == 0 {
			return "experience"
		}
		return "scenario"
	}
}

// Main question generator
func (a *App) nextInterviewQuestion(ctx context.Context, st *InterviewState, cv string) string {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ Panic in nextInterviewQuestion: %v", r)
		}
	}()
	
	qType := decideType(st)

	// Fully deterministic for the first two questions
	if qType == "personal_intro" {
		return "To start off, could you tell me a bit about yourself and your background?"
	}
	if qType == "motivation" {
		return "What motivated you to apply for this type of role and work in this kind of environment?"
	}

	// First preference: CV-tailored question
	if qType != "personal_intro" && qType != "motivation" {
		if cvQ := a.generateCVTailoredQuestion(ctx, st, qType); cvQ != "" {
			if qType == "technical" {
				st.TechAsked++
			}
			return cvQ
		}
	}

	// LLM generic interview question
	var prevBlock string
	if len(st.Asked) == 0 {
		prevBlock = "(none)"
	} else {
		var rows []string
		for _, q := range st.Asked {
			q = strings.TrimSpace(q)
			if q != "" {
				rows = append(rows, "- "+q)
			}
		}
		if len(rows) == 0 {
			prevBlock = "(none)"
		} else {
			prevBlock = strings.Join(rows, "\n")
		}
	}

	typeLine := ""
	switch qType {
	case "experience":
		typeLine = "Ask about a specific project or experience from their CV. Prefer concrete projects if available. Ask for impact, decisions, or challenges."
	case "technical":
		typeLine = "Ask a focused technical question grounded in their listed skills or projects. One concept only, 1–2 sentences."
	case "scenario":
		typeLine = "Ask a realistic work scenario question that fits their field and subfield."
	case "behavioral":
		typeLine = "Ask about past behavior, such as conflict, teamwork, ownership, or dealing with pressure."
	default:
		typeLine = "Ask a professional interview question linked to their background."
	}

	sys := `You are a senior recruiter interviewing a candidate in a live conversation.
GOAL: Generate ONE and ONLY ONE interview question.
GENERAL RULES:
- Maximum 2 sentences.
- Use QUESTION_TYPE_INSTRUCTION to choose the topic.
- Prefer using their CV (skills, projects, experiences) when relevant.
- Sound like a human interviewer, not an AI.
- Vary your language naturally.
- Avoid repeating the same question or wording from PREVIOUS_QUESTIONS.
MUST NOT:
- Do NOT paraphrase or repeat the candidate's previous answers.
- Do NOT ask more than one question.
- Do NOT say "Thanks", "Understood", "Noted", or any acknowledgement.
- Do NOT use emojis.
OUTPUT: Return only the question text.`
	user := fmt.Sprintf(`STRUCTURED_CV:
Field=%s | Subfield=%s
Skills=%v
RelevantProject=%s
Experiences=%v

PREVIOUS_QUESTIONS:
%s

QUESTION_TYPE_INSTRUCTION:
%s`, st.Field, st.Subfield, st.Skills, pickRelevantProject(st, qType), st.Experiences, prevBlock, typeLine)

	if a.LLM == nil {
		switch qType {
		case "technical":
			st.TechAsked++
			return "Could you describe a technical challenge from one of your projects and how you solved it?"
		case "scenario":
			return "Imagine you join a project that is behind schedule. What would you do first?"
		case "behavioral":
			return "Tell me about a time you had a conflict with someone at work and how you handled it."
		default:
			return "Could you walk me through one of the main projects on your CV and your contribution?"
		}
	}

	ctx2, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	resp, err := a.LLM.CreateChatCompletion(ctx2, openai.ChatCompletionRequest{
		Model:       a.ChatModel,
		Temperature: 0.5,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sys},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		MaxTokens: 200,
	})
	if err != nil {
		log.Printf("❌ Error in nextInterviewQuestion LLM call: %v", err)
		switch qType {
		case "technical":
			st.TechAsked++
			return "Could you describe a technical challenge from one of your projects and how you solved it?"
		case "scenario":
			return "Imagine you join a project that is behind schedule. What would you do first?"
		case "behavioral":
			return "Tell me about a time you had a conflict with someone at work and how you handled it."
		default:
			return "Could you walk me through one of the main projects on your CV and your contribution?"
		}
	}
	if len(resp.Choices) == 0 {
		log.Println("⚠️ nextInterviewQuestion: LLM returned no choices")
		switch qType {
		case "technical":
			st.TechAsked++
			return "Could you describe a technical challenge from one of your projects and how you solved it?"
		case "scenario":
			return "Imagine you join a project that is behind schedule. What would you do first?"
		case "behavioral":
			return "Tell me about a time you had a conflict with someone at work and how you handled it."
		default:
			return "Could you walk me through one of the main projects on your CV and your contribution?"
		}
	}

	out := strings.TrimSpace(resp.Choices[0].Message.Content)
	if out == "" {
		log.Println("⚠️ nextInterviewQuestion: LLM returned empty content")
		switch qType {
		case "technical":
			st.TechAsked++
			return "Could you describe a technical challenge from one of your projects and how you solved it?"
		case "scenario":
			return "Imagine you join a project that is behind schedule. What would you do first?"
		case "behavioral":
			return "Tell me about a time you had a conflict with someone at work and how you handled it."
		default:
			return "Could you walk me through one of the main projects on your CV and your contribution?"
		}
	}
	if tooSimilar(out, lastN(st.Asked, 4)) {
		if alt := a.generateCVTailoredQuestion(ctx, st, qType); alt != "" && !tooSimilar(alt, lastN(st.Asked, 4)) {
			out = alt
		}
	}
	if qType == "technical" {
		st.TechAsked++
	}
	return out
}

// Generate CV-tailored question
func (a *App) generateCVTailoredQuestion(ctx context.Context, st *InterviewState, qType string) string {
	if a.LLM == nil {
		return ""
	}
	if len(st.Skills) == 0 && len(st.Projects) == 0 && len(st.Experiences) == 0 {
		return ""
	}

	sys := `You are a senior human interviewer generating a CV-based question.
GOALS:
- Ask ONE creative, fresh, and unique question based on the candidate's CV.
- Rotate topics: skills → project → experience → challenge → decision-making → teamwork → impact.
- Explore a NEW angle every time.
STRICT RULES:
- Max 2 sentences.
- Must be different in structure and topic from the last 4 questions.
- Do NOT reuse similar verbs, patterns, or templates.
- Do NOT repeat any concept asked earlier.
- Do NOT ask generic "describe/explain/tell me".
- No filler, no acknowledgment.
STYLE: Human, conversational, professional. Curious but not interrogative. Realistic questions a real interviewer would ask.`
	user := fmt.Sprintf(`CANDIDATE STRUCTURE:
Field: %s
Subfield: %s
Skills: %v
Projects: %v
Experiences: %v

LAST_QUESTIONS (avoid all topics & patterns):
%v

YOUR TASK: Generate ONE fresh, novel CV-based question exploring a different aspect of the candidate's background.`, st.Field, st.Subfield, st.Skills, st.Projects, st.Experiences, lastN(st.Asked, 4))

	ctx2, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	resp, err := a.LLM.CreateChatCompletion(ctx2, openai.ChatCompletionRequest{
		Model:       a.ChatModel,
		Temperature: 0.4,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sys},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		MaxTokens: 200,
	})
	if err != nil || len(resp.Choices) == 0 {
		return ""
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content)
}

// Generate acknowledgement
var neutralAcks = []string{
	"Alright.", "I understand.", "Okay.", "Good to know.", "Got it.", "That makes sense.", "Thanks for sharing that.",
}
var softEmpathyAcks = []string{
	"I understand how that could feel stressful.", "I hear you, and that's completely valid.", "I get that this can feel a bit overwhelming.",
}

func pickNeutralAck() string {
	return neutralAcks[rand.Intn(len(neutralAcks))]
}

func (a *App) generateAcknowledgement(ctx context.Context, st *InterviewState) string {
	ans := strings.TrimSpace(st.LastUser)
	if ans == "" {
		return pickNeutralAck()
	}
	if isEmotional(ans) {
		return softEmpathyAcks[rand.Intn(len(softEmpathyAcks))]
	}
	if len(ans) < 25 {
		return pickNeutralAck()
	}

	sys := `You are a recruiter. Return ONE very short acknowledgement to the candidate's answer.
RULES:
- Max 1 short sentence.
- Do NOT summarize or repeat the candidate's answer.
- Do NOT mention the specific content or topic of the answer.
- Neutral professional tone.
- No follow-up questions.
- No emojis.`
	user := fmt.Sprintf("Candidate answer: %s\nLast question asked: %s", ans, st.LastQuestion)

	ctx2, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	resp, err := a.LLM.CreateChatCompletion(ctx2, openai.ChatCompletionRequest{
		Model:       a.ChatModel,
		Temperature: 0.2,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sys},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		MaxTokens: 50,
	})
	if err != nil || len(resp.Choices) == 0 {
		return pickNeutralAck()
	}
	ack := strings.TrimSpace(resp.Choices[0].Message.Content)
	if ack == "" {
		return pickNeutralAck()
	}
	return ack
}

// Rephrase question
func (a *App) rephraseQuestion(ctx context.Context, st *InterviewState) string {
	orig := strings.TrimSpace(st.LastQuestion)
	if orig == "" {
		return "Let me phrase that differently: could you walk me through your thought process in that situation?"
	}
	if a.LLM == nil {
		return "Let me say it a different way: " + orig
	}

	sys := `You are a recruiter rephrasing a question because the candidate is confused.
RULES:
- Change the angle completely (not the wording only).
- Make it easier, simpler, more concrete.
- Remove technical wording.
- ONE short sentence only.
- Do NOT repeat the original question structure.
- Do NOT say sorry or apologise.
- Keep it conversational.`
	user := fmt.Sprintf("Original question: %s\nCandidate said: %s", orig, st.LastUser)

	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	resp, err := a.LLM.CreateChatCompletion(ctx2, openai.ChatCompletionRequest{
		Model:       a.ChatModel,
		Temperature: 0.3,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sys},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		MaxTokens: 100,
	})
	if err != nil || len(resp.Choices) == 0 {
		return "Let me rephrase that: " + orig
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content)
}


func (a *App) maybeFollowUp(ctx context.Context, st *InterviewState) string {
	if a.LLM == nil {
		return ""
	}
	if isConfused(st.LastUser) {
		return ""
	}
	if len(st.LastUser) < 40 {
		return ""
	}
	if rand.Float64() > 0.05 {
		return ""
	}

	sys := `You are a professional interviewer. Write ONE follow-up question that stays on the same topic as the original question.
STRICT RULES:
- Max 9 words.
- One sentence.
- Same topic ONLY.
- No generic 'tell me more'.
- No repeating earlier questions.
- No re-asking anything from the last 3 questions.
- No filler, no fluff.`
	user := fmt.Sprintf("LastQuestion: %s\nCandidateAnswer: %s\nPrevious3: %v", st.LastQuestion, st.LastUser, lastN(st.Asked, 3))

	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	resp, err := a.LLM.CreateChatCompletion(ctx2, openai.ChatCompletionRequest{
		Model:       a.ChatModel,
		Temperature: 0.25,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sys},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		MaxTokens: 50,
	})
	if err != nil || len(resp.Choices) == 0 {
		return ""
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content)
}

// Answer candidate question
func (a *App) answerCandidateQuestion(ctx context.Context, st *InterviewState, question string) string {
	q := strings.TrimSpace(question)
	if q == "" {
		return "Those are great questions, and the hiring manager will share more details with you in the next stage."
	}

	sys := `You are a recruiter. The candidate is asking a question about the role or company.
RULES:
- Answer in 2–3 sentences.
- Be concise and realistic.
- If you don't know a detail, give a professional generic answer.
- No emojis. No bullet points.`
	user := fmt.Sprintf("Candidate question: %q\nField: %s | Subfield: %s", q, st.Field, st.Subfield)

	ctx2, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	resp, err := a.LLM.CreateChatCompletion(ctx2, openai.ChatCompletionRequest{
		Model:       a.ChatModel,
		Temperature: 0.4,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sys},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		MaxTokens: 150,
	})
	if err != nil || len(resp.Choices) == 0 {
		return "Great question — the hiring team will share more specifics with you in the next interview stage."
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content)
}

// Generate closing message
func (a *App) generateClosingMessage(ctx context.Context, st *InterviewState) string {
	if a.LLM == nil {
		return "Thank you for your time today. This concludes the interview and we will follow up with you soon."
	}

	sys := `You are a professional recruiter. Your job is to close the interview politely.
STRICT RULES:
- 2–3 sentences maximum.
- Professional and warm.
- No repeated phrases ("Noted", "Understood", "Thanks for sharing") unless really natural.
- No filler, no clichés.
- If the interview felt emotional, include a gentle reassurance.
- Do NOT ask any new questions.
- No emojis.`
	user := fmt.Sprintf(`Field: %s
Subfield: %s
LastUserMessage: %s
SummaryOfQuestions: %v`, st.Field, st.Subfield, st.LastUser, lastN(st.Asked, 8))

	ctx2, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	resp, err := a.LLM.CreateChatCompletion(ctx2, openai.ChatCompletionRequest{
		Model:       a.ChatModel,
		Temperature: 0.3,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sys},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		MaxTokens: 150,
	})
	if err != nil || len(resp.Choices) == 0 {
		return "Thank you for your time. We will stay in touch with the next steps."
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content)
}

func (a *App) handleLiveInterview(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Println("🎙️ New live interview session started")

	sessionID := fmt.Sprintf("live_%d_%d", time.Now().Unix(), time.Now().UnixNano()%1000000)

	query := r.URL.Query()
	userIDStr := query.Get("user_id")
	var userID int64
	if userIDStr != "" {
		fmt.Sscan(userIDStr, &userID)
	}

	// Load latest CV text for this user
	var cvText string
	if userID > 0 {
		log.Printf("🔍 Loading CV for user_id: %d", userID)
		ctx := context.Background()
		query := fmt.Sprintf(`SELECT text FROM ai.cv_documents WHERE user_id = %d ORDER BY created_at DESC LIMIT 1 -- Unique CV query for user %d`, userID, userID)
		err := a.DB.QueryRowContext(ctx, query).Scan(&cvText)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("⚠️ No CV found for user_id: %d", userID)
			} else {
				log.Printf("❌ Error loading CV for user %d: %v", userID, err)
				query2 := fmt.Sprintf(`SELECT text FROM ai.cv_documents WHERE user_id = %d ORDER BY id DESC LIMIT 1 -- Retry CV query for user %d`, userID, userID)
				if err2 := a.DB.QueryRowContext(ctx, query2).Scan(&cvText); err2 != nil {
					log.Printf("❌ Retry also failed for user %d: %v", userID, err2)
				} else {
					log.Printf("✅ CV loaded successfully (retry) for user %d (%d chars)", userID, len(cvText))
				}
			}
		} else {
			log.Printf("✅ CV loaded successfully for user %d (%d characters)", userID, len(cvText))
		}
	}

	// Analyze CV structure once
	st := &InterviewState{Turn: 0, TechAsked: 0}
	if strings.TrimSpace(cvText) != "" && a.LLM != nil {
		if fld, sub, skills, projs, exps := a.extractStructure(r.Context(), cvText); fld != "" {
			st.Field = fld
			st.Subfield = sub
			st.Skills = skills
			st.Projects = projs
			st.Experiences = exps
			log.Printf("📋 CV Structure: Field=%s, Subfield=%s, Skills=%d, Projects=%d, Experiences=%d", 
				fld, sub, len(skills), len(projs), len(exps))
		}
	}

	// 1) Send friendly intro ONLY (no interview question yet)
	opening := a.firstQuestion(r.Context(), st, cvText)
	if opening == "" {
		opening = "Hi there, I'm your interviewer today. We'll go through some general questions and then some based on your CV. Before we dive in, how are you feeling today?"
		log.Println("⚠️ firstQuestion returned empty, using fallback")
	}
	st.LastQuestion = opening
	if err := conn.WriteMessage(websocket.TextMessage, []byte(opening)); err != nil {
		log.Printf("❌ Error sending opening message: %v", err)
		return
	}

	// Conversation loop
	for {
		conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read error: %v", err)
			} else {
				log.Printf("WebSocket closed: %v", err)
			}
			break
		}

		userMsg := strings.TrimSpace(string(msg))
		if userMsg == "" {
			continue
		}
		st.Responses = append(st.Responses, userMsg)
		st.LastUser = userMsg
		log.Println("👤 Candidate:", userMsg)

		// Evaluate answer asynchronously (non-blocking)
		if userID > 0 {
			go func() {
				log.Printf("📊 Evaluating answer asynchronously for user %d", userID)
				evalResult := a.evaluateAnswerInternal(sessionID, userID, st.LastQuestion, userMsg, "")
				if evalResult != nil {
					log.Printf("📊 Evaluation complete: Overall=%.1f, Technical=%.1f, Communication=%.1f, Confidence=%.1f", 
						evalResult["overall"], evalResult["technical_score"], 
						evalResult["communication_score"], evalResult["confidence_score"])
				}
			}()
		}

		// 2) Confusion handling: rephrase the last question
		if isConfused(userMsg) && st.LastQuestion != "" {
			reply := a.rephraseQuestion(r.Context(), st)
			st.LastQuestion = reply
			st.Asked = append(st.Asked, reply)
			_ = conn.WriteMessage(websocket.TextMessage, []byte(reply))
			continue
		}

		// 3) Candidate question phase (end of interview)
		if st.InCandidateQPhase {
			reply := a.answerCandidateQuestion(r.Context(), st, userMsg)
			st.LastQuestion = reply
			st.Asked = append(st.Asked, reply)
			_ = conn.WriteMessage(websocket.TextMessage, []byte(reply))

			// Soft or neutral closing depending on emotion
			var final string
			if isEmotionalEnding(st.LastUser) {
				final = "Thank you for sharing your thoughts today. You expressed yourself well, and we appreciate the effort you put into this interview."
			} else {
				final = a.generateClosingMessage(r.Context(), st)
			}
			_ = conn.WriteMessage(websocket.TextMessage, []byte(final))
			break
		}

		// 4) Optional follow-up (rare, only on long answers)
		if st.Turn > 0 && len(userMsg) > 40 && !isConfused(userMsg) {
			if follow := a.maybeFollowUp(r.Context(), st); follow != "" {
				st.LastQuestion = follow
				st.Asked = append(st.Asked, follow)
				_ = conn.WriteMessage(websocket.TextMessage, []byte(follow))
				continue
			}
		}

		// 5) Advance turn and possibly switch to closing
		st.Turn++

		// After ~9 questions, move to candidate question phase
		if st.Turn >= 9 {
			ack := a.generateAcknowledgement(r.Context(), st)
			closing := strings.TrimSpace(ack + " Before we wrap up, do you have any questions for us about the role or the company?")
			st.LastQuestion = closing
			st.Asked = append(st.Asked, closing)
			st.InCandidateQPhase = true
			_ = conn.WriteMessage(websocket.TextMessage, []byte(closing))
			continue
		}

		// 6) Generate main interview question for this turn
		question := a.nextInterviewQuestion(r.Context(), st, cvText)
		if question == "" {
			question = "Tell me about a project from your CV that you're proud of and why."
		}

		// 7) Generate acknowledgement (only if answer was meaningful)
		ack := ""
		if len(userMsg) > 20 {
			ack = a.generateAcknowledgement(r.Context(), st)
		}

		msgOut := strings.TrimSpace(strings.TrimSpace(ack) + " " + strings.TrimSpace(question))
		if msgOut == "" {
			msgOut = question
		}

		st.LastQuestion = question
		st.Asked = append(st.Asked, question)

		_ = conn.WriteMessage(websocket.TextMessage, []byte(msgOut))
	}

	
	log.Println("📊 Generating interview feedback summary...")
	
	
	if userID > 0 {
		
		
		query := fmt.Sprintf(`
			SELECT result_json 
			FROM ai.evaluations 
			WHERE session_id = '%s'
			ORDER BY created_at
		`, sessionID)
		
		rows, err := a.DB.Query(query)
		
		if err == nil {
			defer rows.Close()
			
			var technicalScores, communicationScores, confidenceScores []float64
			var allStrengths, allWeaknesses []string
			
			for rows.Next() {
				var resultJSON []byte
				if err := rows.Scan(&resultJSON); err == nil {
					var evalResult struct {
						Overall   float64   `json:"overall"`
						TechnicalScore   float64   `json:"technical_score"`
						CommunicationScore float64   `json:"communication_score"`
						ConfidenceScore  float64   `json:"confidence_score"`
						Strengths  []string `json:"strengths"`
						Weaknesses []string `json:"weaknesses"`
					}
					if json.Unmarshal(resultJSON, &evalResult) == nil {
						if evalResult.TechnicalScore > 0 {
							technicalScores = append(technicalScores, evalResult.TechnicalScore)
						}
						if evalResult.CommunicationScore > 0 {
							communicationScores = append(communicationScores, evalResult.CommunicationScore)
						}
						if evalResult.ConfidenceScore > 0 {
							confidenceScores = append(confidenceScores, evalResult.ConfidenceScore)
						}
						allStrengths = append(allStrengths, evalResult.Strengths...)
						allWeaknesses = append(allWeaknesses, evalResult.Weaknesses...)
					}
				}
			}
			
			
			var avgTechnical, avgCommunication, avgConfidence float64
			if len(technicalScores) > 0 {
				for _, s := range technicalScores {
					avgTechnical += s
				}
				avgTechnical /= float64(len(technicalScores))
			}
			if len(communicationScores) > 0 {
				for _, s := range communicationScores {
					avgCommunication += s
				}
				avgCommunication /= float64(len(communicationScores))
			}
			if len(confidenceScores) > 0 {
				for _, s := range confidenceScores {
					avgConfidence += s
				}
				avgConfidence /= float64(len(confidenceScores))
			}
			
			
			behavioralScore := (avgCommunication + avgConfidence) / 2.0
			
			
			feedbackText := fmt.Sprintf("Technical: %.1f%%, Communication: %.1f%%, Confidence: %.1f%%", 
				avgTechnical, avgCommunication, avgConfidence)
			if len(allStrengths) > 0 {
				feedbackText += fmt.Sprintf("\n\nStrengths: %s", strings.Join(allStrengths[:min(5, len(allStrengths))], ", "))
			}
			if len(allWeaknesses) > 0 {
				feedbackText += fmt.Sprintf("\n\nAreas for improvement: %s", strings.Join(allWeaknesses[:min(5, len(allWeaknesses))], ", "))
			}
			
			
			if avgTechnical > 0 || behavioralScore > 0 {
				
				deleteQuery := fmt.Sprintf(`
					DELETE FROM ai.feedback WHERE user_id = %d
				`, userID)
				_, _ = a.DB.Exec(deleteQuery) 
				
				
				feedbackTextEscaped := strings.ReplaceAll(feedbackText, "\\", "\\\\")
				feedbackTextEscaped = strings.ReplaceAll(feedbackTextEscaped, "'", "''")
				insertQuery := fmt.Sprintf(`
					INSERT INTO ai.feedback (session_id, user_id, overall_score, technical_score, communication_score, confidence_score, text_feedback, created_at)
					VALUES ('%s', %d, %.2f, %.2f, %.2f, %.2f, '%s', NOW())
				`, sessionID, userID, (avgTechnical+behavioralScore)/2.0, avgTechnical, avgCommunication, avgConfidence, 
					feedbackTextEscaped)
				
				log.Printf("📝 Storing interview feedback for user %d, session %s", userID, sessionID)
				_, err = a.DB.Exec(insertQuery)
				if err != nil {
					log.Printf("❌ Failed to store feedback: %v", err)
					
					if strings.Contains(err.Error(), "prepared statement") {
						ctx := context.Background()
						_, err = a.DB.ExecContext(ctx, insertQuery)
						if err != nil {
							log.Printf("❌ Retry also failed: %v", err)
						} else {
							log.Printf("✅ Stored feedback (retry succeeded) - Technical: %.1f%%, Behavioral: %.1f%%", avgTechnical, behavioralScore)
						}
					}
				} else {
					log.Printf("✅ Stored feedback - Technical: %.1f%%, Behavioral: %.1f%%", avgTechnical, behavioralScore)
					
					verifyQuery := fmt.Sprintf(`SELECT COUNT(*) FROM ai.feedback WHERE user_id = %d`, userID)
					var count int
					if verifyErr := a.DB.QueryRow(verifyQuery).Scan(&count); verifyErr == nil {
						log.Printf("✅ Verified: %d feedback records for user %d", count, userID)
					}
				}
				
				
				updateQuery := fmt.Sprintf(`
					UPDATE public.users 
					SET technical_score = %.2f, 
					    behavioral_score = %.2f,
					    updated_at = NOW()
					WHERE id = %d
				`, avgTechnical, behavioralScore, userID)
				
				_, err = a.DB.Exec(updateQuery)
				if err != nil {
					log.Printf("⚠️ Failed to update user scores: %v", err)
				} else {
					log.Printf("✅ Updated user profile scores - Technical: %.1f%%, Behavioral: %.1f%%", 
						avgTechnical, behavioralScore)
				}
			}
		} else {
			log.Printf("⚠️ Could not fetch evaluations: %v", err)
		}
	}
	
	log.Println("✅ Live interview ended")
}



type startSessionReq struct {
	UserID int64  `json:"user_id"`
	Major  string `json:"major,omitempty"`
}

type startSessionResp struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

func (a *App) handleStartSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var in startSessionReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if in.UserID <= 0 {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	
	sessionID := fmt.Sprintf("sess_%d_%d", in.UserID, time.Now().UnixNano())

	
	// Use unique query ID to avoid prepared statement cache issues
	uniqueSessionID := time.Now().UnixNano()
	sessionQuery := fmt.Sprintf(`
		INSERT INTO ai.practice_sessions (user_id, major, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		-- Unique session insert ID: %d`, uniqueSessionID)
	
	ctx := r.Context()
	_, err := a.DB.ExecContext(ctx, sessionQuery, in.UserID, in.Major)

	if err != nil {
		log.Printf("Warning: Could not insert practice session: %v", err)
		
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(startSessionResp{
		SessionID: sessionID,
		Message:   "Interview session started",
	})
}

type endSessionReq struct {
	SessionID string `json:"session_id"`
	UserID    int64  `json:"user_id"`
}

func (a *App) handleEndSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var in endSessionReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if in.SessionID == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}

	
	// Use unique query ID to avoid prepared statement cache issues
	uniqueEndSessionID := time.Now().UnixNano()
	endSessionQuery := fmt.Sprintf(`
		SELECT result_json FROM ai.evaluations WHERE session_id = $1
		-- Unique end session query ID: %d`, uniqueEndSessionID)
	
	ctx := r.Context()
	rows, err := a.DB.QueryContext(ctx, endSessionQuery, in.SessionID)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var totalScore float64
	var count int
	var strengths, weaknesses []string

	for rows.Next() {
		var resultJSON []byte
		if err := rows.Scan(&resultJSON); err != nil {
			continue
		}

		var evalResult struct {
			Overall    float64  `json:"overall"`
			Strengths  []string `json:"strengths"`
			Weaknesses []string `json:"weaknesses"`
		}
		if err := json.Unmarshal(resultJSON, &evalResult); err == nil {
			totalScore += evalResult.Overall
			count++
			strengths = append(strengths, evalResult.Strengths...)
			weaknesses = append(weaknesses, evalResult.Weaknesses...)
		}
	}

	grade := 0
	if count > 0 {
		grade = int(totalScore / float64(count))
	}

	
	behavioralFeedback := "Interview completed."
	technicalFeedback := "Review the evaluation results for detailed feedback."

	if len(strengths) > 0 {
		behavioralFeedback = fmt.Sprintf("Key strengths: %s", strings.Join(strengths[:min(3, len(strengths))], ", "))
	}
	if len(weaknesses) > 0 {
		technicalFeedback = fmt.Sprintf("Areas for improvement: %s", strings.Join(weaknesses[:min(3, len(weaknesses))], ", "))
	}

	
	if in.UserID > 0 {
		_, _ = a.DB.Exec(`
			UPDATE ai.practice_sessions 
			SET grade = $1, behavioral_feedback = $2, technical_feedback = $3, updated_at = NOW()
			WHERE user_id = $4 
			ORDER BY created_at DESC 
			LIMIT 1
		`, grade, behavioralFeedback, technicalFeedback, in.UserID)
	}

	response := map[string]interface{}{
		"session_id":          in.SessionID,
		"grade":               grade,
		"behavioral_feedback": behavioralFeedback,
		"technical_feedback":  technicalFeedback,
		"message":             "Interview session ended",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}



type feedbackReq struct {
	SessionID          string  `json:"session_id"`
	UserID             int64   `json:"user_id"`
	OverallScore       float64 `json:"overall_score"`
	TechnicalScore     float64 `json:"technical_score"`
	CommunicationScore float64 `json:"communication_score"`
	ConfidenceScore    float64 `json:"confidence_score"`
	TextFeedback       string  `json:"text_feedback"`
}

func (a *App) handleFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var in feedbackReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if in.SessionID == "" {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}

	
	deleteQuery := fmt.Sprintf(`
		DELETE FROM ai.feedback WHERE user_id = %d
	`, in.UserID)
	_, _ = a.DB.Exec(deleteQuery) 
	
	
	_, err := a.DB.Exec(`
		INSERT INTO ai.feedback (session_id, user_id, overall_score, technical_score, communication_score, confidence_score, text_feedback, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`, in.SessionID, in.UserID, in.OverallScore, in.TechnicalScore, in.CommunicationScore, in.ConfidenceScore, in.TextFeedback)

	if err != nil {
		http.Error(w, "Failed to save feedback: "+err.Error(), http.StatusInternalServerError)
		return
	}

	
	if in.UserID > 0 {
		
		behavioralScore := (in.CommunicationScore + in.ConfidenceScore) / 2.0
		
		_, err = a.DB.Exec(`
			UPDATE public.users 
			SET technical_score = $1, 
			    behavioral_score = $2,
			    updated_at = NOW()
			WHERE id = $3
		`, in.TechnicalScore, behavioralScore, in.UserID)
		if err != nil {
			log.Printf("⚠️ Failed to update user profile scores: %v", err)
		} else {
			log.Printf("✅ Updated user profile - Technical: %.1f%%, Behavioral: %.1f%%", 
				in.TechnicalScore, behavioralScore)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true,"message":"Feedback saved successfully"}`))
}



func (a *App) handleGetUserScores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	var userID int64
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	var cvAnalysisScore sql.NullFloat64
	
	uniqueUserScoresID := time.Now().UnixNano()
	userQuery := fmt.Sprintf(`
		SELECT COALESCE(cv_analysis_score, 0) as cv_analysis_score
		FROM public.users
		WHERE id = %d
		-- Unique query ID: %d
	`, userID, uniqueUserScoresID)
	err := a.DB.QueryRow(userQuery).Scan(&cvAnalysisScore)

	if err != nil {
		if strings.Contains(err.Error(), "prepared statement") {
			ctx := context.Background()
			err = a.DB.QueryRowContext(ctx, userQuery).Scan(&cvAnalysisScore)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "User not found", http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Get technical and behavioral scores from ai.feedback table
	var technicalScore, behavioralScore float64
	uniqueFeedbackID := time.Now().UnixNano()
	feedbackQuery := fmt.Sprintf(`
		SELECT 
			COALESCE(technical_score, 0) as technical_score,
			COALESCE((communication_score + confidence_score) / 2.0, 0) as behavioral_score
		FROM ai.feedback
		WHERE user_id = %d
		ORDER BY created_at DESC
		LIMIT 1
		-- Unique query ID: %d
	`, userID, uniqueFeedbackID)
	
	var feedbackTechnical, feedbackBehavioral sql.NullFloat64
	feedbackErr := a.DB.QueryRow(feedbackQuery).Scan(&feedbackTechnical, &feedbackBehavioral)
	if feedbackErr == nil {
		technicalScore = feedbackTechnical.Float64
		behavioralScore = feedbackBehavioral.Float64
	} else if feedbackErr != sql.ErrNoRows {
		log.Printf("⚠️ Could not fetch feedback scores: %v", feedbackErr)
		// Continue with 0 values if there's an error (but not if no rows found)
	}

	response := map[string]interface{}{
		"user_id":           userID,
		"technical_score":   technicalScore,
		"behavioral_score":  behavioralScore,
		"cv_analysis_score": cvAnalysisScore.Float64,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}




func (a *App) handleGetCvFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	var userID int64
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	// Create context with timeout for the entire request (60 seconds)
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	log.Printf("🔍 Querying CV feedback for user %d (timeout: 60s)", userID)

	
	var cvScore float64
	
	uniqueScoreID := time.Now().UnixNano()
	scoreQuery := fmt.Sprintf(`SELECT COALESCE(cv_analysis_score, 0) FROM public.users WHERE id = $1 -- Unique query ID: %d`, uniqueScoreID)
	err := a.DB.QueryRowContext(ctx, scoreQuery, userID).Scan(&cvScore)
	if err != nil {
		log.Printf("⚠️ Could not fetch CV score: %v", err)
		cvScore = 0
	}

	
	var aiSuggestion, aiResponse string
	var grade int
	var createdAt time.Time
	
	uniqueAnalysisID := time.Now().UnixNano()
	analysisQuery := fmt.Sprintf(`
		SELECT ai_suggestion, ai_response, grade, created_at
		FROM public.cv_analysis
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
		-- Unique query ID: %d
	`, uniqueAnalysisID)
	
	log.Printf("🔍 SQL Query: %s", analysisQuery)
	
	err = a.DB.QueryRowContext(ctx, analysisQuery, userID).Scan(&aiSuggestion, &aiResponse, &grade, &createdAt)
	if err != nil {
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("❌ No CV analysis found for user %d in cv_analysis table", userID)
				// Return empty response instead of 404 so app can handle gracefully
				response := map[string]interface{}{
					"user_id":       userID,
					"grade":         nil,
					"ai_response":   nil,
					"ai_suggestion": nil,
					"score":         cvScore,
					"cv_text":       "",
					"message":       "No CV analysis found. Please upload a CV first.",
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				return
			}
			
			// Check for context timeout
			if ctx.Err() == context.DeadlineExceeded {
				log.Printf("❌ Timeout retrieving CV analysis for user %d", userID)
				http.Error(w, "Request timeout - database query took too long", http.StatusRequestTimeout)
				return
			}
			
			log.Printf("❌ Error retrieving CV analysis: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	
	log.Printf("✅ Found CV analysis for user %d - Grade: %d, Analysis length: %d, Suggestions length: %d, Created: %v", 
		userID, grade, len(aiResponse), len(aiSuggestion), createdAt)

	
	var cvText string
	cvTextQuery := `SELECT text FROM ai.cv_documents WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`
	
	err = a.DB.QueryRowContext(ctx, cvTextQuery, userID).Scan(&cvText)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("ℹ️ CV text not available: %v", err)
		}
		cvText = ""
	}

	
	cvTextPreview := cvText
	if len(cvTextPreview) > 500 {
		cvTextPreview = cvTextPreview[:500]
	}

	response := map[string]interface{}{
		"user_id":       userID,
		"grade":         grade,
		"ai_response":   aiResponse,
		"ai_suggestion": aiSuggestion,
		"score":         cvScore,
		"cv_text":       cvTextPreview,
		"created_at":    createdAt.Format(time.RFC3339),
		"message":       "Latest CV analysis feedback",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


func (a *App) handleGetTechnicalFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	var userID int64
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	
	var technicalScore, overallScore float64
	var textFeedback string
	var createdAt time.Time
	
	
	query := fmt.Sprintf(`
		SELECT 
			technical_score,
			overall_score,
			text_feedback,
			created_at
		FROM ai.feedback
		WHERE user_id = %d
		ORDER BY created_at DESC
		LIMIT 1
	`, userID)
	
	err := a.DB.QueryRow(query).Scan(&technicalScore, &overallScore, &textFeedback, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No technical feedback found", http.StatusNotFound)
			return
		}
		
		if strings.Contains(err.Error(), "prepared statement") {
			ctx := context.Background()
			err = a.DB.QueryRowContext(ctx, query).Scan(&technicalScore, &overallScore, &textFeedback, &createdAt)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "No technical feedback found", http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"user_id":         userID,
		"technical_score": technicalScore,
		"overall_score":   overallScore,
		"feedback":        textFeedback,
		"created_at":      createdAt.Format(time.RFC3339),
		"message":         "Latest technical interview feedback",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


func (a *App) handleGetBehavioralFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	var userID int64
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	
	var communicationScore, confidenceScore, overallScore float64
	var textFeedback string
	var createdAt time.Time
	
	
	query := fmt.Sprintf(`
		SELECT 
			communication_score,
			confidence_score,
			overall_score,
			text_feedback,
			created_at
		FROM ai.feedback
		WHERE user_id = %d
		ORDER BY created_at DESC
		LIMIT 1
	`, userID)
	
	err := a.DB.QueryRow(query).Scan(&communicationScore, &confidenceScore, &overallScore, &textFeedback, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No behavioral feedback found", http.StatusNotFound)
			return
		}
		
		if strings.Contains(err.Error(), "prepared statement") {
			ctx := context.Background()
			err = a.DB.QueryRowContext(ctx, query).Scan(&communicationScore, &confidenceScore, &overallScore, &textFeedback, &createdAt)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "No behavioral feedback found", http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	behavioralScore := (communicationScore + confidenceScore) / 2.0
	
	languageScore := communicationScore

	response := map[string]interface{}{
		"user_id":            userID,
		"behavioral_score":   behavioralScore,
		"communication_score": communicationScore,
		"confidence_score":   confidenceScore,
		"language_score":     languageScore,
		"overall_score":      overallScore,
		"feedback":           textFeedback,
		"created_at":         createdAt.Format(time.RFC3339),
		"message":            "Latest behavioral interview feedback",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}



func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
