package entities

import "testing"

func TestIsValidAuditAction_KnownActions_ReturnsTrue(t *testing.T) {
	knownActions := []string{
		"user.login",
		"user.logout",
		"user.registered",
		"artifact.upload",
		"artifact.extract",
		"insight.generate",
		"insight.approve",
		"insight.reject",
		"summary.send",
		"client.view",
		"auth.login",
		"auth.login_failed",
		"auth.logout",
		"auth.token_refreshed",
		"auth.token_replay_detected",
		"artifact.upload_requested",
		"artifact.upload_confirmed",
		"artifact.classified",
		"artifact.download_requested",
		"artifact.job_enqueued",
	}

	for _, action := range knownActions {
		if !IsValidAuditAction(action) {
			t.Errorf("expected '%s' to be valid", action)
		}
	}
}

func TestIsValidAuditAction_UnknownAction_ReturnsFalse(t *testing.T) {
	unknownActions := []string{
		"",
		"unknown.action",
		"user.delete",
		"admin.override",
	}

	for _, action := range unknownActions {
		if IsValidAuditAction(action) {
			t.Errorf("expected '%s' to be invalid", action)
		}
	}
}

func TestAuditQueryFilter_ZeroValue_IsEmpty(t *testing.T) {
	filter := AuditQueryFilter{}
	if filter.ActorID != "" {
		t.Error("expected empty ActorID")
	}
	if filter.Limit != 0 {
		t.Error("expected zero Limit")
	}
	if filter.From != nil {
		t.Error("expected nil From")
	}
}
