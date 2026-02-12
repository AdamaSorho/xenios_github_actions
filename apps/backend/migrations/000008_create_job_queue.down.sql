-- Drop RLS policies
DROP POLICY IF EXISTS jobs_dead_letter_service_all ON jobs_dead_letter;
DROP POLICY IF EXISTS jobs_service_all ON jobs;

-- Drop dead letter queue table
DROP TABLE IF EXISTS jobs_dead_letter;

-- Drop jobs table
DROP TABLE IF EXISTS jobs;

-- Drop enums
DROP TYPE IF EXISTS job_status;
DROP TYPE IF EXISTS job_type;
