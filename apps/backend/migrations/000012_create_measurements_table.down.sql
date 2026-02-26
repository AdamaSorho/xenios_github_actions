DROP TABLE IF EXISTS measurements;
DROP TYPE IF EXISTS measurement_type;
DROP TYPE IF EXISTS wearable_source;

-- Restore the original generic measurements table from migration 000004
CREATE TABLE IF NOT EXISTS measurements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recorded_by UUID NOT NULL REFERENCES users(id),
    measurement_type TEXT NOT NULL,
    value NUMERIC(10,3) NOT NULL,
    unit TEXT NOT NULL,
    measured_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_measurements_client_id ON measurements(client_id);
CREATE INDEX IF NOT EXISTS idx_measurements_recorded_by ON measurements(recorded_by);
CREATE INDEX IF NOT EXISTS idx_measurements_type ON measurements(measurement_type);
CREATE INDEX IF NOT EXISTS idx_measurements_measured_at ON measurements(measured_at);

ALTER TABLE measurements ENABLE ROW LEVEL SECURITY;
