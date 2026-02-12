-- Revert trigger-based enforcement back to rule-based enforcement

DROP TRIGGER IF EXISTS events_audit_no_update ON events_audit;
DROP TRIGGER IF EXISTS events_audit_no_delete ON events_audit;
DROP FUNCTION IF EXISTS prevent_audit_mutation();

-- Restore the original rules
CREATE OR REPLACE RULE events_audit_no_update AS ON UPDATE TO events_audit DO INSTEAD NOTHING;
CREATE OR REPLACE RULE events_audit_no_delete AS ON DELETE TO events_audit DO INSTEAD NOTHING;
