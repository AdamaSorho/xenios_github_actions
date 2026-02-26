package repository

import (
	"testing"
)

func TestQueryBuilder_Empty(t *testing.T) {
	qb := newQueryBuilder()
	if clause := qb.whereClause(); clause != "" {
		t.Errorf("expected empty string, got %q", clause)
	}
	if idx := qb.nextParamIdx(); idx != 1 {
		t.Errorf("expected paramIdx 1, got %d", idx)
	}
}

func TestQueryBuilder_AddCondition(t *testing.T) {
	qb := newQueryBuilder()
	qb.addCondition("actor_id", "user-1")

	expected := " WHERE actor_id = $1"
	if clause := qb.whereClause(); clause != expected {
		t.Errorf("expected %q, got %q", expected, clause)
	}
	if len(qb.args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(qb.args))
	}
	if qb.args[0] != "user-1" {
		t.Errorf("expected arg 'user-1', got %v", qb.args[0])
	}
	if idx := qb.nextParamIdx(); idx != 2 {
		t.Errorf("expected paramIdx 2, got %d", idx)
	}
}

func TestQueryBuilder_MultipleConditions(t *testing.T) {
	qb := newQueryBuilder()
	qb.addCondition("actor_id", "user-1")
	qb.addCondition("action", "login")

	expected := " WHERE actor_id = $1 AND action = $2"
	if clause := qb.whereClause(); clause != expected {
		t.Errorf("expected %q, got %q", expected, clause)
	}
	if len(qb.args) != 2 {
		t.Errorf("expected 2 args, got %d", len(qb.args))
	}
	if idx := qb.nextParamIdx(); idx != 3 {
		t.Errorf("expected paramIdx 3, got %d", idx)
	}
}

func TestQueryBuilder_AddTimeCondition(t *testing.T) {
	qb := newQueryBuilder()
	qb.addTimeCondition("created_at", ">=", "2026-01-01")

	expected := " WHERE created_at >= $1"
	if clause := qb.whereClause(); clause != expected {
		t.Errorf("expected %q, got %q", expected, clause)
	}
	if len(qb.args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(qb.args))
	}
}

func TestQueryBuilder_MixedConditions(t *testing.T) {
	qb := newQueryBuilder()
	qb.addCondition("client_id", "client-1")
	qb.addCondition("measurement_type", "weight")
	qb.addTimeCondition("measured_at", ">=", "2026-01-01")
	qb.addTimeCondition("measured_at", "<=", "2026-02-01")

	expected := " WHERE client_id = $1 AND measurement_type = $2 AND measured_at >= $3 AND measured_at <= $4"
	if clause := qb.whereClause(); clause != expected {
		t.Errorf("expected %q, got %q", expected, clause)
	}
	if len(qb.args) != 4 {
		t.Errorf("expected 4 args, got %d", len(qb.args))
	}
	if idx := qb.nextParamIdx(); idx != 5 {
		t.Errorf("expected paramIdx 5, got %d", idx)
	}
}

func TestQueryBuilder_NextParamIdx(t *testing.T) {
	qb := newQueryBuilder()
	if idx := qb.nextParamIdx(); idx != 1 {
		t.Errorf("expected 1, got %d", idx)
	}

	qb.addCondition("a", "b")
	if idx := qb.nextParamIdx(); idx != 2 {
		t.Errorf("expected 2, got %d", idx)
	}

	qb.addTimeCondition("c", ">=", "d")
	if idx := qb.nextParamIdx(); idx != 3 {
		t.Errorf("expected 3, got %d", idx)
	}
}
