package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"code.sajari.com/docconv"
	openai "github.com/sashabaranov/go-openai"
)






func (a *App) handleUploadCV(w http.ResponseWriter, r *http.Request) {
	log.Println("📄 /v1/cv/upload hit")
	
	
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Keep-Alive", "timeout=600")
	
	
	log.Printf("📡 Client: %s, User-Agent: %s", r.RemoteAddr, r.UserAgent())

	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	
	
	
	err := r.ParseMultipartForm(20 << 20) 
	if err != nil {
		log.Printf("❌ ParseMultipartForm error: %v (Remote: %s)", err, r.RemoteAddr)
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
			http.Error(w, "Upload timeout: connection too slow. Please try again or use a smaller file.", http.StatusRequestTimeout)
		} else if strings.Contains(err.Error(), "connection reset") || strings.Contains(err.Error(), "broken pipe") {
			http.Error(w, "Connection lost during upload. Please check your network and try again.", http.StatusBadGateway)
		} else {
			http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		}
		return
	}

	
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Printf("❌ FormFile error: %v", err)
		http.Error(w, fmt.Sprintf("No file provided or file error: %v", err), http.StatusBadRequest)
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("⚠️ Error closing file: %v", closeErr)
		}
	}()

	log.Printf("📦 Uploaded file: %s (Size: %d bytes)", handler.Filename, handler.Size)

	
	if handler.Size > 20<<20 {
		log.Printf("❌ File too large: %d bytes", handler.Size)
		http.Error(w, "File size exceeds 20MB limit", http.StatusBadRequest)
		return
	}

	if handler.Size == 0 {
		log.Println("❌ Empty file uploaded")
		http.Error(w, "File is empty", http.StatusBadRequest)
		return
	}

	
	userIDStr := strings.TrimSpace(r.FormValue("user_id"))
	if userIDStr == "" {
		log.Println("❌ Missing user_id")
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	log.Printf("🧠 Received user_id (raw): %s", userIDStr)

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		log.Printf("❌ Invalid user_id (not an integer): %s, error: %v", userIDStr, err)
		http.Error(w, fmt.Sprintf("Invalid user_id: %s", userIDStr), http.StatusBadRequest)
		return
	}

	if userID <= 0 {
		log.Printf("❌ Invalid user_id (must be positive): %d", userID)
		http.Error(w, "user_id must be a positive number", http.StatusBadRequest)
		return
	}

	log.Printf("➡️ Parsed userID (int64): %d", userID)

	
	data := make([]byte, handler.Size)
	n, err := io.ReadFull(file, data)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		log.Printf("❌ File read error: %v (read %d bytes)", err, n)
		http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
		return
	}
	if n == 0 {
		log.Println("❌ No data read from file")
		http.Error(w, "File appears to be empty", http.StatusBadRequest)
		return
	}
	data = data[:n] 
	log.Printf("📊 Read %d bytes from file", len(data))

	
	text, err := extractText(data, handler.Filename)
	if err != nil {
		log.Printf("❌ Text extraction error: %v", err)
		
		http.Error(w, fmt.Sprintf("Failed to extract text from file. Supported formats: PDF, DOCX, DOC, TXT. Error: %v", err), http.StatusBadRequest)
		return
	}

	
	text = normalizeCVText(text)
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		log.Println("❌ No text extracted from file")
		http.Error(w, "No readable text found in the file. Please ensure the file contains text and is not corrupted.", http.StatusBadRequest)
		return
	}

	if len(text) < 50 {
		log.Printf("⚠️ Very short text extracted: %d characters", len(text))
		
	}

	log.Printf("✅ Extracted %d characters from file", len(text))

	
	
	
	ctx := context.Background()
	var userExists bool
	
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM public.users WHERE id = %d)`, userID)
	err = a.DB.QueryRowContext(ctx, query).Scan(&userExists)
	if err != nil {
		log.Printf("⚠️ Error checking user existence: %v", err)
		
		userExists = false
	}

	if !userExists {
		log.Printf("📝 User %d does not exist, creating...", userID)
		
		
		var seqName string
		
		err = a.DB.QueryRowContext(ctx, `SELECT pg_get_serial_sequence('public.users', 'id')`).Scan(&seqName)
		if err == nil && seqName != "" {
			
			seqQuery := fmt.Sprintf(`SELECT setval('%s', GREATEST(COALESCE((SELECT MAX(id) FROM public.users), 0), %d), true)`, seqName, userID)
			_, err = a.DB.ExecContext(ctx, seqQuery)
			if err != nil {
				log.Printf("⚠️ Failed to set sequence: %v", err)
			}
		}
		
		
		
		
		username := fmt.Sprintf("user%d", userID)
		email := fmt.Sprintf("user%d@test.com", userID)
		name := fmt.Sprintf("User %d", userID)
		
		
		
		
		emailEscaped := strings.ReplaceAll(email, "'", "''")
		nameEscaped := strings.ReplaceAll(name, "'", "''")
		usernameEscaped := strings.ReplaceAll(username, "'", "''")
		
		insertQuery := fmt.Sprintf(`
			INSERT INTO public.users (id, email, name, username, created_at, updated_at)
			VALUES (%d, '%s', '%s', '%s', NOW(), NOW())
			ON CONFLICT (id) DO UPDATE SET updated_at = NOW()
		`, userID, emailEscaped, nameEscaped, usernameEscaped)
		
		_, err = a.DB.ExecContext(ctx, insertQuery)
		if err != nil {
			
			if strings.Contains(err.Error(), "column \"username\"") {
				log.Printf("⚠️ Username column not found, trying without it")
				insertQuery = fmt.Sprintf(`
					INSERT INTO public.users (id, email, name, created_at, updated_at)
					VALUES (%d, '%s', '%s', NOW(), NOW())
					ON CONFLICT (id) DO UPDATE SET updated_at = NOW()
				`, userID, emailEscaped, nameEscaped)
				_, err = a.DB.ExecContext(ctx, insertQuery)
			}
			if err != nil {
				log.Printf("⚠️ Failed to create user: %v", err)
				
			} else {
				log.Printf("✅ Created user %d", userID)
			}
		} else {
			log.Printf("✅ Created user %d (with username)", userID)
		}
	}

	
	
	var cvID int64
	
	textEscaped := strings.ReplaceAll(text, "\\", "\\\\")
	textEscaped = strings.ReplaceAll(textEscaped, "'", "''")
	
	
	maxTextLength := 10 * 1024 * 1024 
	if len(textEscaped) > maxTextLength {
		log.Printf("⚠️ CV text is very long (%d chars), truncating to %d chars", len(textEscaped), maxTextLength)
		textEscaped = textEscaped[:maxTextLength]
	}
	
	
	
	uniqueDocID := time.Now().UnixNano()
	insertQuery := fmt.Sprintf(`
		INSERT INTO ai.cv_documents (user_id, text, created_at)
		VALUES (%d, '%s', NOW())
		RETURNING id
		
	`, userID, textEscaped, uniqueDocID)
	
	log.Printf("📝 Inserting CV document for user %d (text length: %d chars)", userID, len(textEscaped))
	err = a.DB.QueryRowContext(ctx, insertQuery).Scan(&cvID)

	if err != nil {
		log.Printf("❌ Failed to insert CV document: %v", err)
		log.Printf("❌ Error type: %T", err)
		log.Printf("❌ Error details: %+v", err)
		
		
		if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "user_id") {
			log.Printf("❌ Foreign key constraint error")
			http.Error(w, fmt.Sprintf("Invalid user_id: user %d does not exist and could not be created", userID), http.StatusBadRequest)
			return
		} else if strings.Contains(err.Error(), "prepared statement") {
			
			log.Printf("⚠️ Prepared statement cache issue, retrying with NEW query string...")
			ctx2 := context.Background()
			uniqueDocID2 := time.Now().UnixNano()
			insertQuery2 := fmt.Sprintf(`
				INSERT INTO ai.cv_documents (user_id, text, created_at)
				VALUES (%d, '%s', NOW())
				RETURNING id
				
			`, userID, textEscaped, uniqueDocID2)
			err = a.DB.QueryRowContext(ctx2, insertQuery2).Scan(&cvID)
			if err != nil {
				log.Printf("❌ Retry also failed: %v", err)
				
				log.Printf("⚠️ Trying alternative approach...")
				uniqueDocID3 := time.Now().UnixNano()
				execQuery := fmt.Sprintf(`
					INSERT INTO ai.cv_documents (user_id, text, created_at)
					VALUES (%d, '%s', NOW())
					
				`, userID, textEscaped, uniqueDocID3)
				_, execErr := a.DB.ExecContext(ctx2, execQuery)
				if execErr != nil {
					log.Printf("❌ Alternative approach also failed: %v", execErr)
					http.Error(w, fmt.Sprintf("Database error inserting CV: %v. Please check backend logs.", execErr), http.StatusInternalServerError)
					return
				}
				
				uniqueSelectID := time.Now().UnixNano()
				selectQuery := fmt.Sprintf(`SELECT id FROM ai.cv_documents WHERE user_id = %d ORDER BY created_at DESC LIMIT 1 `, userID, uniqueSelectID)
				err = a.DB.QueryRowContext(ctx2, selectQuery).Scan(&cvID)
				if err != nil {
					log.Printf("❌ Failed to get CV ID: %v", err)
					http.Error(w, fmt.Sprintf("Database error retrieving CV ID: %v. Please check backend logs.", err), http.StatusInternalServerError)
					return
				}
			}
		} else {
			
			log.Printf("❌ Database error details: %+v", err)
			http.Error(w, fmt.Sprintf("Database error: %v. Please check backend logs for details.", err), http.StatusInternalServerError)
			return
		}
	}
	log.Printf("✅ CV document stored with ID: %d", cvID)

	
	chunks := chunkCVText(text, 4000, 400)
	log.Printf("📊 Split CV into %d chunks", len(chunks))

	
	embedCtx, embedCancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer embedCancel()
	embeddings, err := a.embed(embedCtx, chunks)
	if err != nil {
		log.Println(" Warning: Could not generate embeddings:", err)
		
		embeddings = make([][]float32, len(chunks))
	}

	
	
	ctx = context.Background()
	for i, chunk := range chunks {
		var embJSON string
		if embeddings[i] != nil {
			embBytes, _ := json.Marshal(embeddings[i])
			embJSON = string(embBytes)
			
			embJSON = strings.ReplaceAll(embJSON, "\\", "\\\\")
			embJSON = strings.ReplaceAll(embJSON, "'", "''")
		} else {
			embJSON = "null"
		}
		
		
		chunkEscaped := strings.ReplaceAll(chunk, "'", "''")
		chunkEscaped = strings.ReplaceAll(chunkEscaped, "\\", "\\\\")
		
		
		
		uniqueChunkID := time.Now().UnixNano() + int64(i)
		insertQuery := fmt.Sprintf(`
			INSERT INTO ai.cv_chunks (cv_id, chunk_text, embedding_json, ord, created_at)
			VALUES (%d, '%s', '%s'::jsonb, %d, NOW())
			
		`, cvID, chunkEscaped, embJSON, i, uniqueChunkID)
		
		if _, err := a.DB.ExecContext(ctx, insertQuery); err != nil {
			log.Printf(" Failed to insert chunk %d: %v", i, err)
			
			if strings.Contains(err.Error(), "prepared statement") {
				ctx2 := context.Background()
				uniqueChunkID2 := time.Now().UnixNano() + int64(i) + 1000000
				insertQuery2 := fmt.Sprintf(`
					INSERT INTO ai.cv_chunks (cv_id, chunk_text, embedding_json, ord, created_at)
					VALUES (%d, '%s', '%s'::jsonb, %d, NOW())
					
				`, cvID, chunkEscaped, embJSON, i, uniqueChunkID2)
				if _, retryErr := a.DB.ExecContext(ctx2, insertQuery2); retryErr != nil {
					log.Printf("❌ Retry also failed for chunk %d: %v", i, retryErr)
					http.Error(w, fmt.Sprintf("Database error inserting chunk %d: %v", i, retryErr), http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, fmt.Sprintf("Database error inserting chunk %d: %v", i, err), http.StatusInternalServerError)
				return
			}
		}
	}
	log.Printf("✅ Stored %d chunks for CV ID %d", len(chunks), cvID)

	
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var result struct {
		Field    string `json:"field"`
		Analysis string `json:"analysis"`
	}

	
	if a.LLM != nil {
		sysPrompt := `You are an expert AI career analyst and CV reviewer. Analyze the candidate's CV comprehensively and provide:

1. **Field Identification**: The candidate's likely professional field (e.g., Software Engineering, Marketing, Business, Mechanical Engineering, Medicine, etc.)

2. **Comprehensive Analysis**: Provide a detailed analysis that includes:
   - **Strengths**: Key achievements, skills, and positive aspects (2-3 points)
   - **Areas for Improvement**: Specific aspects the candidate lacks or can improve (2-3 points)
   - **CV Structure Recommendations**: How the CV can be structured better (formatting, organization, sections to add/improve)
   - **Overall Assessment**: Summary of the candidate's profile

Format your response as a well-structured analysis that helps the candidate understand:
- What they're doing well
- What they're missing
- How to improve their CV structure and content
- Specific skills or experiences they should add

IMPORTANT: Return ONLY valid JSON with this exact structure:
{
  "field": "Field Name",
  "analysis": "Comprehensive analysis (5-8 sentences) covering strengths, areas for improvement, CV structure recommendations, and overall assessment. Be specific and actionable."
}`

		resp, err := a.LLM.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       a.ChatModel,
			Temperature: 0.3,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: sysPrompt},
				{Role: openai.ChatMessageRoleUser, Content: "Analyze this CV:\n\n" + text},
			},
		})
		if err != nil {
			log.Printf("⚠️ AI analysis failed: %v", err)
			result.Field = "Unknown"
			result.Analysis = "Analysis unavailable due to API error: " + err.Error()
		} else {
			raw := strings.TrimSpace(resp.Choices[0].Message.Content)
			log.Printf("📝 Raw AI response: %s", raw)
			
			jsonStr := extractJSON(raw)
			log.Printf("📝 Extracted JSON: %s", jsonStr)
			
			if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
				log.Printf("⚠️ Failed to parse JSON: %v. Raw response: %s", err, raw)
				
				result.Field = "Unknown"
				if strings.Contains(strings.ToLower(raw), "software") || strings.Contains(strings.ToLower(raw), "engineering") {
					result.Field = "Software Engineering"
				}
				result.Analysis = "Analysis received but could not be parsed. Raw response: " + raw[:min(200, len(raw))]
			}
		}
	} else {
		
		result.Field = "General"
		result.Analysis = "CV uploaded successfully. Analysis available in real mode."
	}

	if result.Field == "" {
		result.Field = "Unknown"
	}
	if result.Analysis == "" {
		result.Analysis = "Analysis is being processed. Please try again in a moment."
	}

	log.Printf("🧾 DEBUG — userID=%d | field='%s' | analysis='%s'\n", userID, result.Field, result.Analysis)

	
	
	
	
	cvGrade := 75
	if len(result.Analysis) > 200 {
		cvGrade = 80
	}
	if strings.Contains(strings.ToLower(result.Analysis), "strength") && 
	   strings.Contains(strings.ToLower(result.Analysis), "improvement") {
		cvGrade = 85
	}
	
	
	analysisText := result.Analysis
	fieldText := result.Field
	
	if len(analysisText) > 5*1024*1024 {
		analysisText = analysisText[:5*1024*1024]
	}
	if len(fieldText) > 1000 {
		fieldText = fieldText[:1000]
	}
	
	
	log.Printf("🔍 Verifying cv_analysis table exists...")
	tableCheckQuery := fmt.Sprintf(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'cv_analysis'
		)
	`)
	var tableExists bool
	ctxCheck := context.Background()
	if err := a.DB.QueryRowContext(ctxCheck, tableCheckQuery).Scan(&tableExists); err != nil {
		log.Printf("⚠️ Could not verify table existence: %v", err)
		tableExists = true
	} else if !tableExists {
		log.Printf("❌ Table public.cv_analysis does not exist! Creating it...")
		createTableQuery := `
			CREATE TABLE IF NOT EXISTS public.cv_analysis (
				id BIGSERIAL PRIMARY KEY,
				user_id BIGINT NOT NULL,
				grade INTEGER CHECK (grade >= 0 AND grade <= 100),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				ai_suggestion TEXT,
				ai_response TEXT,
				CONSTRAINT cv_analysis_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id)
			)
		`
		if _, err := a.DB.ExecContext(ctxCheck, createTableQuery); err != nil {
			log.Printf("❌ Failed to create cv_analysis table: %v", err)
		} else {
			log.Printf("✅ Created cv_analysis table")
			tableExists = true
		}
	} else {
		log.Printf("✅ Table public.cv_analysis exists")
	}
	
	if tableExists {
		deleteQuery := fmt.Sprintf(`DELETE FROM public.cv_analysis WHERE user_id = %d`, userID)
		_, _ = a.DB.ExecContext(ctxCheck, deleteQuery)
		log.Printf("🗑️ Deleted old CV analysis for user %d (if any)", userID)
	} 
	
	
	analysisEscaped := strings.ReplaceAll(analysisText, "'", "''")
	fieldEscaped := strings.ReplaceAll(fieldText, "'", "''")
	
	suggestionText := "Based on your CV analysis, here are key recommendations:\n\n"
	if strings.Contains(strings.ToLower(result.Analysis), "strength") {
		suggestionText += "• Continue highlighting your strengths and achievements\n"
	}
	if strings.Contains(strings.ToLower(result.Analysis), "improvement") || strings.Contains(strings.ToLower(result.Analysis), "lack") {
		suggestionText += "• Focus on areas identified for improvement in your analysis\n"
	}
	if strings.Contains(strings.ToLower(result.Analysis), "structure") || strings.Contains(strings.ToLower(result.Analysis), "format") {
		suggestionText += "• Consider restructuring your CV for better readability\n"
	}
	if strings.Contains(strings.ToLower(result.Analysis), "skill") {
		suggestionText += "• Ensure all relevant technical skills are clearly listed\n"
	}
	if len(suggestionText) == len("Based on your CV analysis, here are key recommendations:\n\n") {
		suggestionText += "• Add quantifiable achievements with specific metrics\n• Strengthen your technical skills section\n• Ensure clear formatting and organization\n• Highlight relevant projects and experiences"
	}
	suggestionEscaped := strings.ReplaceAll(suggestionText, "'", "''")
	
	log.Printf("📝 Storing CV analysis for user %d", userID)
	log.Printf("📊 Analysis length: %d chars, Suggestion length: %d chars, Grade: %d", len(analysisEscaped), len(suggestionEscaped), cvGrade)
	
	ctxInsert := context.Background()
	var insertSuccess bool
	var execErr error
	
	uniqueInsertID := time.Now().UnixNano()
	cvAnalysisInsert := fmt.Sprintf(`
		INSERT INTO public.cv_analysis (user_id, ai_response, ai_suggestion, grade, created_at, updated_at)
		VALUES (%d, '%s', '%s', %d, NOW(), NOW())
		
	`, userID, analysisEscaped, suggestionEscaped, cvGrade, uniqueInsertID)
	
	log.Printf("🔍 Executing INSERT query for user %d", userID)
	result, execErr := a.DB.ExecContext(ctxInsert, cvAnalysisInsert)
	insertSuccess = execErr == nil
	
	if insertSuccess {
		rowsAffected, _ := result.RowsAffected()
		log.Printf("✅ CV analysis stored successfully (rows affected: %d)", rowsAffected)
		
		verifyQuery := fmt.Sprintf(`SELECT id, user_id, grade FROM public.cv_analysis WHERE user_id = %d ORDER BY created_at DESC LIMIT 1 `, userID, time.Now().UnixNano())
		var verifyID, verifyUserID int64
		var verifyGrade int
		verifyErr := a.DB.QueryRowContext(ctxInsert, verifyQuery).Scan(&verifyID, &verifyUserID, &verifyGrade)
		if verifyErr == nil {
			log.Printf("✅ Verification: CV analysis confirmed in database (ID: %d, User: %d, Grade: %d)", verifyID, verifyUserID, verifyGrade)
		} else {
			log.Printf("⚠️ Verification query failed (non-fatal): %v", verifyErr)
		}
	} else {
		log.Printf("❌ Failed to store CV analysis: %v", execErr)
		if execErr != nil {
			log.Printf("❌ Error details: %T - %s", execErr, execErr.Error())
			
			if strings.Contains(execErr.Error(), "prepared statement") {
				log.Printf("🔄 Prepared statement error detected, retrying with fresh context...")
				ctxRetry := context.Background()
				uniqueRetryID := time.Now().UnixNano()
				retryQuery := fmt.Sprintf(`
					INSERT INTO public.cv_analysis (user_id, ai_response, ai_suggestion, grade, created_at, updated_at)
					VALUES (%d, '%s', '%s', %d, NOW(), NOW())
					
				`, userID, analysisEscaped, suggestionEscaped, cvGrade, uniqueRetryID)
				retryResult, retryErr := a.DB.ExecContext(ctxRetry, retryQuery)
				if retryErr == nil {
					rowsAffected, _ := retryResult.RowsAffected()
					log.Printf("✅ CV analysis stored successfully on retry (rows affected: %d)", rowsAffected)
					insertSuccess = true
					execErr = nil
				} else {
					log.Printf("❌ Retry also failed: %v", retryErr)
				}
			}
		}
	}

	
	
	if insertSuccess {
		log.Printf("📊 Updating user profile score (CV analysis was stored successfully)")
		
		
		
		cvScore := 75.0 
		if len(result.Analysis) > 200 {
			cvScore = 80.0 
		}
		if strings.Contains(strings.ToLower(result.Analysis), "strength") && 
		   strings.Contains(strings.ToLower(result.Analysis), "improvement") {
			cvScore = 85.0 
		}
		
		
		
		ensureColumnQuery := fmt.Sprintf(`
			DO $$ 
			BEGIN
				IF NOT EXISTS (
					SELECT 1 FROM information_schema.columns 
					WHERE table_schema = 'public' 
					AND table_name = 'users' 
					AND column_name = 'cv_analysis_score'
				) THEN
					ALTER TABLE public.users 
					ADD COLUMN cv_analysis_score NUMERIC DEFAULT 0 
					CHECK (cv_analysis_score >= 0 AND cv_analysis_score <= 100);
				END IF;
			END $$;
		`)
		
		ctxScore := context.Background()
		_, err = a.DB.ExecContext(ctxScore, ensureColumnQuery)
		if err != nil {
			log.Printf("⚠️ Could not ensure cv_analysis_score column exists: %v", err)
			
		} else {
			log.Printf("✅ Verified cv_analysis_score column exists")
		}
		
		
		
		
		uniqueScoreID := time.Now().UnixNano()
		updateQuery := fmt.Sprintf(`
			UPDATE public.users 
			SET cv_analysis_score = %.2f,
			    updated_at = NOW()
			WHERE id = %d
			
		`, cvScore, userID, uniqueScoreID)
		
		log.Printf("📝 Updating cv_analysis_score to %.2f for user %d", cvScore, userID)
		result, err := a.DB.ExecContext(ctxScore, updateQuery)
		if err != nil {
			
			if strings.Contains(err.Error(), "column \"cv_analysis_score\"") || strings.Contains(err.Error(), "does not exist") {
				log.Printf("❌ CV analysis score column not found in database. Score will not be updated.")
				log.Printf("ℹ️ To enable CV scores, add column: ALTER TABLE public.users ADD COLUMN cv_analysis_score NUMERIC DEFAULT 0 CHECK (cv_analysis_score >= 0 AND cv_analysis_score <= 100);")
			} else if strings.Contains(err.Error(), "prepared statement") {
				
				log.Printf("🔄 Prepared statement error, retrying score update...")
				ctx2 := context.Background()
				result, err = a.DB.ExecContext(ctx2, updateQuery)
				if err != nil {
					if strings.Contains(err.Error(), "column \"cv_analysis_score\"") {
						log.Printf("❌ CV analysis score column not found. Score will not be updated.")
					} else {
						log.Printf("❌ Retry also failed: %v", err)
					}
				} else {
					rowsAffected, _ := result.RowsAffected()
					log.Printf("✅ Updated user CV analysis score: %.1f%% (rows affected: %d)", cvScore, rowsAffected)
				}
			} else {
				log.Printf("❌ Failed to update CV analysis score: %v", err)
			}
		} else {
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				log.Printf("✅ Updated user CV analysis score: %.1f%% (rows affected: %d)", cvScore, rowsAffected)
			} else {
				log.Printf("⚠️ Score update returned 0 rows affected - user %d might not exist", userID)
			}
		}
	} else {
		log.Printf("⚠️ Skipping score update - CV analysis was not stored successfully")
	}

	
	
	
	if !insertSuccess {
		log.Printf("⚠️ WARNING: CV analysis generated but NOT stored in database")
		log.Printf("⚠️ User will see analysis in response, but it won't be available on profile page")
		log.Printf("⚠️ This may be due to:")
		log.Printf("   - Database connection issue")
		log.Printf("   - Table structure mismatch")
		log.Printf("   - Permission issue")
		log.Printf("   - Data too large for column")
		log.Printf("⚠️ Analysis will still be returned in API response")
	} else {
		log.Printf("✅ CV analysis successfully stored and will be available on profile page")
	}
	
	
	cvScore := 75.0
	if len(result.Analysis) > 200 {
		cvScore = 80.0
	}
	if strings.Contains(strings.ToLower(result.Analysis), "strength") && 
	   strings.Contains(strings.ToLower(result.Analysis), "improvement") {
		cvScore = 85.0
	}
	
	respData := map[string]interface{}{
		"message":         "CV uploaded and analyzed successfully",
		"cv_id":           cvID,
		"chunks_count":    len(chunks),
		"field":           result.Field,
		"analysis":        result.Analysis,
		"analysis_result": result.Analysis,
		"ai_suggestion":   suggestionText,
		"text_length":     len(text),
		"cv_score":        cvScore,
		"stored":          insertSuccess, 
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(respData); err != nil {
		log.Printf("❌ Failed to encode response: %v", err)
		
		return
	}

	log.Printf("✅ CV processed successfully for user %d (CV ID: %d, Chunks: %d, Text: %d chars, Stored: %v)", 
		userID, cvID, len(chunks), len(text), insertSuccess)
}


func chunkCVText(s string, size, overlap int) []string {
	if size <= overlap {
		overlap = 0
	}
	var out []string
	for i := 0; i < len(s); i += size - overlap {
		end := i + size
		if end > len(s) {
			end = len(s)
		}
		chunk := strings.TrimSpace(s[i:end])
		if len(chunk) > 0 {
			out = append(out, chunk)
		}
		if end == len(s) {
			break
		}
	}
	return out
}





func extractText(data []byte, filename string) (string, error) {
	
	if len(data) == 0 {
		return "", fmt.Errorf("file is empty")
	}

	
	filenameLower := strings.ToLower(filename)
	var contentType string
	
	switch {
	case strings.HasSuffix(filenameLower, ".pdf"):
		contentType = "application/pdf"
	case strings.HasSuffix(filenameLower, ".docx"):
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case strings.HasSuffix(filenameLower, ".doc"):
		
		contentType = "application/msword"
	case strings.HasSuffix(filenameLower, ".txt"):
		contentType = "text/plain"
	case strings.HasSuffix(filenameLower, ".rtf"):
		contentType = "application/rtf"
	default:
		return "", fmt.Errorf("unsupported file type: %s. Supported formats: PDF, DOCX, DOC, TXT, RTF", filename)
	}

	log.Printf("📄 Extracting text from %s (type: %s, size: %d bytes)", filename, contentType, len(data))

	
	reader := bytes.NewReader(data)
	
	
	res, err := docconv.Convert(reader, contentType, true)
	if err != nil {
		log.Printf("❌ docconv conversion error: %v", err)
		
		if contentType == "application/msword" {
			log.Println("🔄 Retrying DOC file with alternative method...")
			
			reader.Seek(0, io.SeekStart)
			res, err = docconv.Convert(reader, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", true)
			if err != nil {
				return "", fmt.Errorf("failed to extract text from DOC file: %w. Please try converting to DOCX or PDF", err)
			}
		} else {
			return "", fmt.Errorf("failed to extract text: %w. Please ensure the file is not corrupted and is a valid %s file", err, contentType)
		}
	}

	
	extractedText := strings.TrimSpace(res.Body)
	if len(extractedText) == 0 {
		return "", fmt.Errorf("no text could be extracted from the file. The file may be empty, corrupted, or contain only images")
	}

	log.Printf("✅ Successfully extracted %d characters from %s", len(extractedText), filename)
	return extractedText, nil
}





func normalizeCVText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	sp := regexp.MustCompile(`[ \t]+`)
	s = sp.ReplaceAllString(s, " ")
	bl := regexp.MustCompile(`\n{2,}`)
	s = bl.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}
