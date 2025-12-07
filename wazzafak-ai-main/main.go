package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
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
		log.Printf("⚠️ Detected Supabase pooler (port 6543) - switching to direct connection (port 5432)")
		log.Printf("⚠️ Pooler doesn't support prepared statements, which causes 'prepared statement already exists' errors")
		
		
		
		
		
		queryParams := ""
		if strings.Contains(dsn, "?") {
			parts := strings.Split(dsn, "?")
			if len(parts) > 1 {
				queryParams = parts[1]
				dsn = parts[0] 
			}
		}
		
		
		projectRef := ""
		password := ""
		if strings.Contains(dsn, "postgres.") && strings.Contains(dsn, "@") {
			
			userPart := strings.Split(dsn, "@")[0]
			
			if strings.Contains(userPart, "://") {
				userPart = strings.Split(userPart, "://")[1]
			}
			
			userParts := strings.Split(userPart, ":")
			if len(userParts) >= 2 {
				
				if strings.Contains(userParts[0], ".") {
					projectRef = strings.Split(userParts[0], ".")[1]
				}
				
				password = userParts[1]
			}
		}
		
		
		if projectRef != "" && password != "" {
			
			dsn = fmt.Sprintf("postgresql://postgres:%s@db.%s.supabase.co:5432/postgres", password, projectRef)
			
			if queryParams != "" {
				dsn = dsn + "?" + queryParams
			}
			log.Printf("✅ Switched to direct connection: db.%s.supabase.co:5432", projectRef)
		} else {
			
			log.Printf("⚠️ Could not parse pooler URL, using simple replacement")
			dsn = strings.ReplaceAll(dsn, ".pooler.supabase.com:6543", ".supabase.co:5432")
			dsn = strings.ReplaceAll(dsn, "pooler.supabase.com:6543", "supabase.co:5432")
			
			if strings.Contains(dsn, "postgres.") {
				
				re := regexp.MustCompile(`postgres\.([^:]+):`)
				dsn = re.ReplaceAllString(dsn, "postgres:")
			}
			
			if queryParams != "" {
				if !strings.Contains(dsn, "?") {
					dsn = dsn + "?" + queryParams
				}
			}
		}
		log.Printf("✅ Switched to direct connection to avoid prepared statement cache issues")
	}
	
	
	
	
	if !strings.Contains(dsn, "prepareThreshold") {
		if strings.Contains(dsn, "?") {
			dsn += "&prepareThreshold=0"
		} else {
			dsn += "?prepareThreshold=0"
		}
	}
	
	if !strings.Contains(dsn, "sslmode") {
		if strings.Contains(dsn, "?") {
			dsn += "&sslmode=require"
		} else {
			dsn += "?sslmode=require"
		}
	}
	
	if !strings.Contains(dsn, "statement_cache_mode") {
		dsn += "&statement_cache_mode=describe"
	}
	log.Printf("🔗 Database connection string configured (prepared statements DISABLED via prepareThreshold=0, sslmode=require)")
	
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute) 
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	
	
	var version string
	if err := db.QueryRow("SELECT version()").Scan(&version); err != nil {
		
		log.Printf("⚠️ Database test query failed (non-fatal): %v", err)
		log.Printf("ℹ️ Continuing anyway - connection may still be functional")
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

	
	q := "next interview question for " + in.Domain
	qEmb, err := a.embed(r.Context(), []string{q})
	if err != nil || len(qEmb) == 0 {
		http.Error(w, "embedding error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	queryEmb := qEmb[0]

	
	cvRows, err := a.DB.Query(`
	  SELECT c.chunk_text, c.embedding_json
	  FROM ai.cv_chunks c
	  JOIN ai.cv_documents d ON d.id = c.cv_id
	  WHERE d.user_id = $1
	  ORDER BY c.ord
	  LIMIT 200`, in.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	kbRows, err := a.DB.Query(`
	  SELECT text, embedding_json
	  FROM ai.kb_articles
	  WHERE domain = $1
	  LIMIT 200`, in.Domain)
	if err != nil {
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
		q := "Based on your background, describe how you would design a simple REST API in " + in.Domain + "."
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(nextQResp{Question: q})
		return
	}

	
	sys := fmt.Sprintf(`You are a strict but fair interviewer in the %s domain.
Ask ONE question (<=2 sentences). Use CONTEXT if needed. Do not reveal answers.`, in.Domain)
	user := "CONTEXT:\n" + contextBlock + "\n---\nDifficulty: " + in.Difficulty + "\nReturn exactly one interview question now."

	resp, err := a.LLM.CreateChatCompletion(r.Context(), openai.ChatCompletionRequest{
		Model:       a.ChatModel,
		Temperature: 0.3,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sys},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	question := strings.TrimSpace(resp.Choices[0].Message.Content)

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
		kbRows, err := a.DB.Query(`
			SELECT text FROM ai.kb_articles WHERE domain = $1 LIMIT 6
		`, domain)
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
		log.Printf("⚠️ Evaluation error: %v, using fallback", err)
		
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
		log.Printf("⚠️ Failed to parse evaluation JSON: %v", err)
		
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

	kbRows, err := a.DB.Query(`
	  SELECT text, embedding_json
	  FROM ai.kb_articles
	  WHERE domain = $1
	  LIMIT 200`, in.Domain)
	if err != nil {
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

	
	
	var cvText string
	if userID > 0 {
		
		query := fmt.Sprintf(`SELECT text FROM ai.cv_documents WHERE user_id = %d ORDER BY id DESC LIMIT 1`, userID)
		err := a.DB.QueryRow(query).Scan(&cvText)
		if err != nil && err != sql.ErrNoRows {
			log.Println("Error loading CV:", err)
			
			if strings.Contains(err.Error(), "prepared statement") {
				ctx := context.Background()
				err = a.DB.QueryRowContext(ctx, query).Scan(&cvText)
				if err != nil && err != sql.ErrNoRows {
					log.Println("Error loading CV (retry):", err)
				}
			}
		}
	}

	
	cvContext := ""
	if cvText != "" {
		
		cvPreview := cvText
		if len(cvPreview) > 3000 {
			cvPreview = cvPreview[:3000] + "..."
		}
		cvContext = fmt.Sprintf(`
CANDIDATE'S CV CONTENT (READ THIS CAREFULLY):
%s

IMPORTANT: You MUST ask questions about specific items from this CV. Reference their:
- Skills and technologies mentioned
- Work experience and projects
- Education and certifications
- Achievements and accomplishments
- Any specific tools, frameworks, or methodologies listed`, cvPreview)
	}
	
	systemPrompt := fmt.Sprintf(`You are a senior professional interviewer conducting a comprehensive mock interview. Your role is to assess the candidate thoroughly across multiple dimensions, with HEAVY EMPHASIS on their CV content.

%s

INTERVIEW STRUCTURE & APPROACH:
1. **Opening (First Message)**: Greet professionally: "Hello! Thank you for taking the time today. I'm [Your Name], and I'll be conducting your interview. Let's begin with a brief introduction - could you tell me a bit about yourself?"

2. **Question Types to Cover** (mix throughout the interview, PRIORITIZE CV-BASED QUESTIONS):
   a) **CV-Specific Technical Questions** (MOST IMPORTANT - 50%% of questions):
      - Ask about specific technologies, tools, or skills mentioned in their CV
      - "I see you worked with [Technology from CV]. Can you tell me about your experience with that?"
      - "Your CV mentions [Project/Skill]. Walk me through how you implemented/used that."
      - "You listed [Technology] in your skills. Can you explain a project where you used it?"
      - Deep dive into their work experience: "Tell me more about your role at [Company from CV]."
      - Ask about specific projects: "I noticed you worked on [Project from CV]. What was your contribution?"
      - For Software Engineering: Ask about specific languages, frameworks, databases mentioned in CV
      - For other fields: Ask about domain-specific tools, methodologies, or experiences from CV
   
   b) **Technical Questions** (based on CV field):
      - Algorithms, data structures, system design (if software engineering)
      - Domain-specific technical knowledge related to their CV field
   
   c) **Behavioral Questions** (STAR method) - reference CV when possible:
      - "Tell me about a time when you..." (prefer experiences from their CV)
      - "I see you worked at [Company from CV]. Describe a challenging situation there..."
      - "Give me an example of how you handled..." (relate to their CV experience)
   
   d) **Scenario-Based Questions** (related to CV):
      - "Based on your experience with [CV item], how would you approach..."
      - "Given your background in [CV field], what would you do if..."
      - "Imagine you're faced with [scenario related to their CV experience]..."

3. **Interview Flow** (CV-FOCUSED):
   - Start with introduction/background (1 question) - ask them to elaborate on their CV summary
   - **CV-Specific Technical Deep Dive (5-6 questions)** - This is the MAIN FOCUS:
     * Ask about each major technology/skill from their CV
     * Ask about their work experience in detail
     * Ask about specific projects they listed
     * Ask about their education/certifications if relevant
   - Include behavioral questions (2-3 questions) - reference CV experiences
   - Add scenario-based questions (2-3 questions) - relate to CV background
   - Conclude with any final questions or wrap-up (1 question)
   - Total: 10-12 questions, with 50%%+ being CV-specific

4. **Question Quality**:
   - Ask ONE question at a time
   - **ALWAYS reference specific items from their CV when possible**
   - Be specific: "I see you worked with React. Can you explain how you used it in your [Project Name]?"
   - Follow up on answers with deeper CV-based questions
   - Vary difficulty (mix easy, medium, challenging)
   - Test both knowledge and thinking process
   - If CV is available, make 60-70%% of questions CV-specific

5. **Professional Tone**:
   - Maintain a professional, respectful, and encouraging tone
   - Provide brief acknowledgments of good answers
   - If answer is weak, ask clarifying questions rather than being critical
   - Keep questions clear and concise

6. **CV Integration** (CRITICAL):
   - If CV is provided, you MUST ask questions about it
   - Reference specific companies, projects, technologies, skills from the CV
   - Ask for details about experiences listed in the CV
   - Verify their claimed skills by asking technical questions about them
   - Don't just ask generic questions - make them CV-relevant

7. **Conclusion**:
   - After 10-12 exchanges, thank them professionally
   - "Thank you for your time today. We'll be in touch soon. Do you have any questions for us?"

IMPORTANT: 
- Only ask questions or provide brief acknowledgments
- Do NOT provide answers or hints
- Keep responses conversational and natural
- Focus on assessment, not teaching
- **PRIORITIZE CV-BASED QUESTIONS - Make most questions reference their CV**`, cvContext)

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
	}

	
	conn.WriteMessage(websocket.TextMessage, []byte("Connected. Starting interview..."))
	log.Println("✅ WebSocket connection established, sending greeting...")

	
	if a.LLM == nil && !a.Mock {
		log.Println("❌ LLM client is not initialized. Check USE_MOCK and LLM_API_KEY settings.")
		conn.WriteMessage(websocket.TextMessage, []byte("Error: AI service not configured. Please check backend configuration."))
		return
	}
	
	
	if a.Mock {
		mockGreeting := "Hello! Thank you for taking the time today. I'll be conducting your interview. Let's begin with a brief introduction - could you tell me a bit about yourself?"
		log.Println("🤖 Interviewer (MOCK):", mockGreeting)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: mockGreeting,
		})
		conn.WriteMessage(websocket.TextMessage, []byte(mockGreeting))
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		resp, err := a.LLM.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       a.ChatModel,
			Temperature: 0.7,
			Messages:    messages,
		})
		cancel()

		if err != nil {
			log.Printf("❌ Error starting interview: %v", err)
			conn.WriteMessage(websocket.TextMessage, []byte("Error starting interview: "+err.Error()))
			return
		}

		firstReply := strings.TrimSpace(resp.Choices[0].Message.Content)
		log.Println("🤖 Interviewer:", firstReply)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: firstReply,
		})
		conn.WriteMessage(websocket.TextMessage, []byte(firstReply))
	}

	
	
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	
	for {
		
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		
		_, msg, err := conn.ReadMessage()
		if err != nil {
			
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read error: %v", err)
			} else {
				log.Printf("WebSocket closed: %v", err)
			}
			break
		}

		userMsg := string(msg)
		log.Println("👤 Candidate:", userMsg)

		
		
		if userID > 0 {
			
			lastQuestion := "Interview Question"
			for i := len(messages) - 1; i >= 0; i-- {
				if messages[i].Role == openai.ChatMessageRoleAssistant {
					lastQuestion = messages[i].Content
					break
				}
			}
			
			
			evalResult := a.evaluateAnswerInternal(sessionID, userID, lastQuestion, userMsg, "")
			if evalResult != nil {
				log.Printf("📊 Evaluation: Overall=%.1f, Technical=%.1f, Communication=%.1f, Confidence=%.1f", 
					evalResult["overall"], evalResult["technical_score"], 
					evalResult["communication_score"], evalResult["confidence_score"])
			}
		}

		
		if a.Mock {
			reply := "Interesting — could you tell me a bit more about that?"
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: userMsg,
			})
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: reply,
			})
			conn.WriteMessage(websocket.TextMessage, []byte(reply))
			continue
		}

		if a.LLM == nil {
			log.Println("❌ LLM client is not initialized")
			conn.WriteMessage(websocket.TextMessage, []byte("Error: AI service not available"))
			continue
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: userMsg,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		resp, err := a.LLM.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       a.ChatModel,
			Temperature: 0.7,
			Messages:    messages,
		})
		cancel()

		if err != nil {
			log.Printf("❌ AI error: %v", err)
			conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
			continue
		}

		reply := strings.TrimSpace(resp.Choices[0].Message.Content)
		log.Println("🤖 AI:", reply)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: reply,
		})
		conn.WriteMessage(websocket.TextMessage, []byte(reply))
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

	
	_, err := a.DB.Exec(`
		INSERT INTO ai.practice_sessions (user_id, major, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
	`, in.UserID, in.Major)

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

	
	rows, err := a.DB.Query(`
		SELECT result_json FROM ai.evaluations WHERE session_id = $1
	`, in.SessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	var technicalScore, behavioralScore, cvAnalysisScore sql.NullFloat64
	
	
	uniqueUserScoresID := time.Now().UnixNano()
	query := fmt.Sprintf(`
		SELECT 
			COALESCE(technical_score, 0) as technical_score,
			COALESCE(behavioral_score, 0) as behavioral_score,
			COALESCE(cv_analysis_score, 0) as cv_analysis_score
		FROM public.users
		WHERE id = %d
		
	`, userID, uniqueUserScoresID)
	err := a.DB.QueryRow(query).Scan(&technicalScore, &behavioralScore, &cvAnalysisScore)

	if err != nil {
		
		if strings.Contains(err.Error(), "prepared statement") {
			ctx := context.Background()
			err = a.DB.QueryRowContext(ctx, query).Scan(&technicalScore, &behavioralScore, &cvAnalysisScore)
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

	response := map[string]interface{}{
		"user_id":           userID,
		"technical_score":   technicalScore.Float64,
		"behavioral_score":  behavioralScore.Float64,
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

	
	var cvScore float64
	
	uniqueScoreID := time.Now().UnixNano()
	scoreQuery := fmt.Sprintf(`SELECT COALESCE(cv_analysis_score, 0) FROM public.users WHERE id = %d `, userID, uniqueScoreID)
	ctxScore := context.Background()
	err := a.DB.QueryRowContext(ctxScore, scoreQuery).Scan(&cvScore)
	if err != nil {
		log.Printf("⚠️ Could not fetch CV score: %v", err)
		cvScore = 0
	}

	
	var field, analysis string
	var createdAt time.Time
	
	uniqueAnalysisID := time.Now().UnixNano()
	analysisQuery := fmt.Sprintf(`
		SELECT ai_suggestion, ai_response, created_at
		FROM public.cv_analysis
		WHERE user_id = %d
		ORDER BY created_at DESC
		LIMIT 1
		
	`, userID, uniqueAnalysisID)
	
	log.Printf("🔍 Querying CV analysis for user %d", userID)
	log.Printf("🔍 SQL Query: %s", analysisQuery)
	
	ctx := context.Background()
	err = a.DB.QueryRowContext(ctx, analysisQuery).Scan(&field, &analysis, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("❌ No CV analysis found for user %d in cv_analysis table", userID)
			
			checkQuery := fmt.Sprintf(`SELECT COUNT(*) FROM public.cv_analysis`)
			var totalCount int
			ctxCheck := context.Background()
			if checkErr := a.DB.QueryRowContext(ctxCheck, checkQuery).Scan(&totalCount); checkErr == nil {
				log.Printf("ℹ️ Total CV analyses in table: %d", totalCount)
			} else {
				log.Printf("⚠️ Could not check total count: %v", checkErr)
			}
			
			userCheckQuery := fmt.Sprintf(`SELECT COUNT(*) FROM public.cv_analysis WHERE user_id = %d`, userID)
			var userCount int
			ctxUserCheck := context.Background()
			if userCheckErr := a.DB.QueryRowContext(ctxUserCheck, userCheckQuery).Scan(&userCount); userCheckErr == nil {
				log.Printf("ℹ️ CV analyses for user %d: %d", userID, userCount)
				if userCount > 0 {
					log.Printf("⚠️ WARNING: User has %d CV analyses but query returned no rows - possible data issue", userCount)
					
					simpleQuery := fmt.Sprintf(`SELECT ai_suggestion, ai_response, created_at FROM public.cv_analysis WHERE user_id = %d ORDER BY created_at DESC LIMIT 1`, userID)
					if simpleErr := a.DB.QueryRowContext(ctxUserCheck, simpleQuery).Scan(&field, &analysis, &createdAt); simpleErr == nil {
						log.Printf("✅ Found CV analysis with simpler query!")
						
					} else {
						log.Printf("❌ Simpler query also failed: %v", simpleErr)
						http.Error(w, "No CV analysis found", http.StatusNotFound)
						return
					}
				} else {
					http.Error(w, "No CV analysis found", http.StatusNotFound)
					return
				}
			} else {
				log.Printf("⚠️ Could not check user count: %v", userCheckErr)
				http.Error(w, "No CV analysis found", http.StatusNotFound)
				return
			}
		}
		
		if strings.Contains(err.Error(), "prepared statement") {
			log.Printf("🔄 Prepared statement error, retrying with fresh context...")
			ctxRetry := context.Background()
			err = a.DB.QueryRowContext(ctxRetry, analysisQuery).Scan(&field, &analysis, &createdAt)
			if err != nil {
				if err == sql.ErrNoRows {
					log.Printf("❌ No CV analysis found for user %d (retry)", userID)
					http.Error(w, "No CV analysis found", http.StatusNotFound)
					return
				}
				log.Printf("❌ Error retrieving CV analysis (retry): %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			log.Printf("❌ Error retrieving CV analysis: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	
	log.Printf("✅ Found CV analysis for user %d - Field: %s, Analysis length: %d, Created: %v", 
		userID, field, len(analysis), createdAt)

	
	var cvText string
	cvTextQuery := fmt.Sprintf(`
		SELECT text
		FROM ai.cv_documents
		WHERE user_id = %d
		ORDER BY created_at DESC
		LIMIT 1
	`, userID)
	
	err = a.DB.QueryRow(cvTextQuery).Scan(&cvText)
	if err != nil {
		
		cvText = ""
		log.Printf("ℹ️ CV text not available: %v", err)
	}

	
	cvTextPreview := cvText
	if len(cvTextPreview) > 500 {
		cvTextPreview = cvTextPreview[:500]
	}

	response := map[string]interface{}{
		"user_id":     userID,
		"score":       cvScore,
		"field":       field,
		"analysis":    analysis,
		"cv_text":     cvTextPreview,
		"created_at":  createdAt.Format(time.RFC3339),
		"message":     "Latest CV analysis feedback",
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
