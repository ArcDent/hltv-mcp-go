package errors

import "testing"

func TestNew(t *testing.T) {
	err := New(CodeEntityNotFound, "no team matched", false, nil)
	if err.Code != CodeEntityNotFound {
		t.Errorf("code mismatch: %s", err.Code)
	}
	if err.Retryable {
		t.Error("expected non-retryable")
	}
}
