-- Migration script to add technical_score and behavioral_score columns to public.users table
-- Run this if the automatic migration doesn't work

-- Add technical_score column if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_schema = 'public' 
        AND table_name = 'users' 
        AND column_name = 'technical_score'
    ) THEN
        ALTER TABLE public.users 
        ADD COLUMN technical_score NUMERIC DEFAULT 0 
        CHECK (technical_score >= 0 AND technical_score <= 100);
        RAISE NOTICE 'Added technical_score column';
    ELSE
        RAISE NOTICE 'technical_score column already exists';
    END IF;
END $$;

-- Add behavioral_score column if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_schema = 'public' 
        AND table_name = 'users' 
        AND column_name = 'behavioral_score'
    ) THEN
        ALTER TABLE public.users 
        ADD COLUMN behavioral_score NUMERIC DEFAULT 0 
        CHECK (behavioral_score >= 0 AND behavioral_score <= 100);
        RAISE NOTICE 'Added behavioral_score column';
    ELSE
        RAISE NOTICE 'behavioral_score column already exists';
    END IF;
END $$;

-- Verify columns were added
SELECT 
    column_name, 
    data_type, 
    column_default,
    is_nullable
FROM information_schema.columns 
WHERE table_schema = 'public' 
AND table_name = 'users' 
AND column_name IN ('technical_score', 'behavioral_score')
ORDER BY column_name;

