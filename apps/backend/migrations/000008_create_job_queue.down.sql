-- Drop RLS policies
DROP POLICY IF EXISTS job_events_service_all ON job_events;
DROP POLICY IF EXISTS jobs_dead_letter_service_all ON jobs_dead_letter;
DROP POLICY IF EXISTS jobs_service_all ON jobs;

-- Drop job events table
DROP TABLE IF EXISTS job_events;

-- Drop dead letter queue table
DROP TABLE IF EXISTS jobs_dead_letter;

-- Drop jobs table
DROP TABLE IF EXISTS jobs;

-- Drop enums
DROP TYPE IF EXISTS job_status;
DROP TYPE IF EXISTS job_type;
