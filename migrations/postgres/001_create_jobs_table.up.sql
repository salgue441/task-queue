-- Create enum types
CREATE TYPE job_status AS ENUM (
    'pending',
    'running',
    'completed',
    'failed',
    'retrying',
    'dead'
);

CREATE TYPE job_priority AS ENUM (
    'low',
    'normal',
    'high',
    'critical'
);

-- Create jobs table
CREATE TABLE IF NOT EXISTS jobs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  type VARCHAR(100) NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}',
  status job_status NOT NULL DEFAULT 'pending',
  priority job_priority NOT NULL DEFAULT 'normal',
  max_retries INTEGER NOT NULL DEFAULT 3,
  retry_count INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  scheduled_at TIMESTAMP WITH TIME ZONE,
  started_at TIMESTAMP WITH TIME ZONE,
  completed_at TIMESTAMP WITH TIME ZONE,
  error TEXT,
  result JSONB,
  worker_id VARCHAR(100),
  metadata JSONB DEFAULT '{}'
);

-- Create indexes for query performance
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_priority ON jobs(priority DESC);
CREATE INDEX idx_jobs_scheduled_at ON jobs(scheduled_at) WHERE scheduled_at IS NOT NULL;
CREATE INDEX idx_jobs_created_at ON jobs(created_at);
CREATE INDEX idx_jobs_type ON jobs(type);
CREATE INDEX idx_jobs_worker_id ON jobs(worker_id) WHERE worker_id IS NOT NULL;
CREATE INDEX idx_jobs_status_priority ON jobs(status, priority DESC) WHERE status = 'pending';

-- Create job_events table for audit trail
CREATE TABLE IF NOT EXISTS job_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(100)
);

CREATE INDEX idx_job_events_job_id ON job_events(job_id);
CREATE INDEX idx_job_events_created_at ON job_events(created_at);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURN TRIGGER AS $$
BEGIN 
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_jobs_updated_at
BEFORE UPDATE ON jobs
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Create function to log job events
CREATE OR REPLACE FUNCTION log_job_event()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO job_events (job_id, event_type, event_data)
        VALUES (NEW.id, 'created', to_jsonb(NEW));
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO job_events (job_id, event_type, event_data)
        VALUES (
            NEW.id, 
            'status_changed', 
            jsonb_build_object(
                'old_status', OLD.status,
                'new_status', NEW.status,
                'old_worker_id', OLD.worker_id,
                'new_worker_id', NEW.worker_id
            )
        );
        
        -- Log additional events based on status changes
        IF NEW.status = 'running' AND OLD.status != 'running' THEN
            INSERT INTO job_events (job_id, event_type, event_data)
            VALUES (NEW.id, 'started', jsonb_build_object('started_at', NEW.started_at));
        ELSIF NEW.status IN ('completed', 'failed', 'dead') AND OLD.status NOT IN ('completed', 'failed', 'dead') THEN
            INSERT INTO job_events (job_id, event_type, event_data)
            VALUES (
                NEW.id, 
                'finished', 
                jsonb_build_object(
                    'status', NEW.status,
                    'completed_at', NEW.completed_at,
                    'error', NEW.error,
                    'result', NEW.result
                )
            );
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to log job events
CREATE TRIGGER log_job_changes
AFTER INSERT OR UPDATE ON jobs
FOR EACH ROW
EXECUTE FUNCTION log_job_event();

-- Create function for getting next available job
CREATE OR REPLACE FUNCTION get_next_job(worker_id VARCHAR, max_jobs INTEGER DEFAULT 1)
RETURNS SETOF jobs AS $$
BEGIN
    RETURN QUERY
    UPDATE jobs
    SET 
        status = 'running',
        worker_id = worker_id,
        started_at = NOW(),
        updated_at = NOW()
    WHERE id IN (
        SELECT id
        FROM jobs
        WHERE status = 'pending'
        AND (scheduled_at IS NULL OR scheduled_at <= NOW())
        ORDER BY priority DESC, created_at
        LIMIT max_jobs
        FOR UPDATE SKIP LOCKED
    )
    RETURNING *;
END;
$$ LANGUAGE plpgsql;

-- Create function for retrying failed jobs
CREATE OR REPLACE FUNCTION retry_failed_jobs(max_retries INTEGER DEFAULT 3)
RETURNS INTEGER AS $$
DECLARE
    rows_affected INTEGER;
BEGIN
    WITH jobs_to_retry AS (
        SELECT id
        FROM jobs
        WHERE status = 'failed'
        AND retry_count < max_retries
        FOR UPDATE SKIP LOCKED
    )
    UPDATE jobs
    SET 
        status = 'retrying',
        retry_count = retry_count + 1,
        error = NULL,
        worker_id = NULL,
        updated_at = NOW()
    WHERE id IN (SELECT id FROM jobs_to_retry)
    RETURNING id INTO rows_affected;
    
    RETURN rows_affected;
END;
$$ LANGUAGE plpgsql;