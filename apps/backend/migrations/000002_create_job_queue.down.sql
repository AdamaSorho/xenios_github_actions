-- Drop audit table
DROP TABLE IF EXISTS events_audit;

-- Drop dead letter queue table
DROP TABLE IF EXISTS jobs_dead_letter;

-- Drop jobs table
DROP TABLE IF EXISTS jobs;

-- Drop enums
DROP TYPE IF EXISTS job_status;
DROP TYPE IF EXISTS job_type;
