-- Replace append-only rules with a trigger for stronger enforcement.
-- Rules silently discard UPDATE/DELETE; a trigger raises an error,
-- which is required for HIPAA-compliant audit trail integrity.

-- Drop the existing rules (they silently swallow writes)
DROP RULE IF EXISTS events_audit_no_update ON events_audit;
DROP RULE IF EXISTS events_audit_no_delete ON events_audit;

-- Create a trigger function that raises an exception on UPDATE or DELETE
CREATE OR REPLACE FUNCTION prevent_audit_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'events_audit is append-only: % operations are not allowed', TG_OP;
END;
$$ LANGUAGE plpgsql;

-- Apply the trigger for UPDATE
DROP TRIGGER IF EXISTS events_audit_no_update ON events_audit;
CREATE TRIGGER events_audit_no_update
    BEFORE UPDATE ON events_audit
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_mutation();

-- Apply the trigger for DELETE
DROP TRIGGER IF EXISTS events_audit_no_delete ON events_audit;
CREATE TRIGGER events_audit_no_delete
    BEFORE DELETE ON events_audit
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_mutation();
