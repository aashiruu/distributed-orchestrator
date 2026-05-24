-- Ensure enum type is created for type-safe status transitions
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_status') THEN
        CREATE TYPE job_status AS ENUM ('PENDING', 'RUNNING', 'SUCCESS', 'FAILED', 'DEAD');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    status job_status NOT NULL DEFAULT 'PENDING',
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    error_log TEXT,
    locked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index critical state columns to keep query scanning complexity at O(1) or O(log N)
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at);
