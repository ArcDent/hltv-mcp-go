package errors

import (
	stderrors "errors"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("ENTITY_NOT_FOUND", "no team matched", false, nil)
	if err.Code != "ENTITY_NOT_FOUND" {
		t.Errorf("code mismatch: %s", err.Code)
	}
	if err.Retryable {
		t.Error("expected non-retryable")
	}
}

func TestIs(t *testing.T) {
	if !Is(New("UPSTREAM_TIMEOUT", "timeout", true, nil)) {
		t.Error("expected true")
	}
	if Is(stderrors.New("plain")) {
		t.Error("expected false for plain error")
	}
}
