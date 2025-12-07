-- Initialize PostgreSQL with PGVector extension and AI schema
-- This script runs automatically when the database container starts

-- Enable PGVector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create the ai schema
CREATE SCHEMA IF NOT EXISTS ai;

-- Create public.users table if not exists (referenced by foreign keys)
CREATE TABLE IF NOT EXISTS public.users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- AI Interview scores (default to 0 for new users)
    technical_score NUMERIC DEFAULT 0 CHECK (technical_score >= 0 AND technical_score <= 100),
    behavioral_score NUMERIC DEFAULT 0 CHECK (behavioral_score >= 0 AND behavioral_score <= 100),
    cv_analysis_score NUMERIC DEFAULT 0 CHECK (cv_analysis_score >= 0 AND cv_analysis_score <= 100)
);

-- ============================================
-- AI Schema Tables
-- ============================================

-- CV Documents table - stores full CV text
CREATE TABLE IF NOT EXISTS ai.cv_documents (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT cv_documents_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_cv_documents_user_id ON ai.cv_documents(user_id);

-- CV Chunks table - stores CV chunks with embeddings
CREATE TABLE IF NOT EXISTS ai.cv_chunks (
    id BIGSERIAL PRIMARY KEY,
    cv_id BIGINT NOT NULL,
    chunk_text TEXT NOT NULL,
    embedding_json JSONB,
    ord INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT cv_chunks_cv_id_fkey FOREIGN KEY (cv_id) REFERENCES ai.cv_documents(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_cv_chunks_cv_id ON ai.cv_chunks(cv_id);

-- Evaluations table - stores interview evaluation results
CREATE TABLE IF NOT EXISTS ai.evaluations (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    result_json JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_evaluations_session_id ON ai.evaluations(session_id);

-- Feedback table - stores interview feedback scores
CREATE TABLE IF NOT EXISTS ai.feedback (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    session_id VARCHAR(255) NOT NULL,
    overall_score NUMERIC,
    technical_score NUMERIC,
    communication_score NUMERIC,
    confidence_score NUMERIC,
    text_feedback TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT feedback_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_feedback_session_id ON ai.feedback(session_id);
CREATE INDEX IF NOT EXISTS idx_feedback_user_id ON ai.feedback(user_id);

-- Knowledge Base Articles table - stores domain knowledge for interviews
CREATE TABLE IF NOT EXISTS ai.kb_articles (
    id BIGSERIAL PRIMARY KEY,
    domain VARCHAR(255) NOT NULL,
    title TEXT,
    text TEXT,
    embedding_json JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_kb_articles_domain ON ai.kb_articles(domain);

-- Practice Sessions table - stores interview practice session data
CREATE TABLE IF NOT EXISTS ai.practice_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    grade INTEGER CHECK (grade >= 0 AND grade <= 100),
    major VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    behavioral_feedback TEXT,
    technical_feedback TEXT,
    CONSTRAINT practice_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_practice_sessions_user_id ON ai.practice_sessions(user_id);

-- ============================================
-- Public Schema Tables (for backward compatibility)
-- ============================================

-- CV Analyses table (backward compatibility)
CREATE TABLE IF NOT EXISTS public.cv_analyses (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    analysis TEXT,
    comments TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT cv_analyses_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_cv_analyses_user_id ON public.cv_analyses(user_id);

-- ============================================
-- Insert sample user for testing
-- ============================================

INSERT INTO public.users (id, email, name) 
VALUES (1, 'test@example.com', 'Test User')
ON CONFLICT (id) DO NOTHING;

-- Verify setup
DO $$
BEGIN
    RAISE NOTICE 'Database initialization complete!';
    RAISE NOTICE 'PGVector extension: enabled';
    RAISE NOTICE 'AI schema tables: created';
END $$;

