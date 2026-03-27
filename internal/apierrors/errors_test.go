package apierrors

import (
	"errors"
	"fmt"
	"net/url"
	"testing"
)

func TestFormat_NilError(t *testing.T) {
	if err := Format(nil); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestFormat_NetworkError(t *testing.T) {
	urlErr := &url.Error{
		Op:  "Get",
		URL: "https://api.emma.ms/external/v1/vms",
		Err: fmt.Errorf("connection refused"),
	}
	err := Format(urlErr)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !contains(err.Error(), "network error") {
		t.Errorf("expected 'network error' in message, got: %s", err.Error())
	}
}

func TestFormat_GenericError(t *testing.T) {
	wrapped := fmt.Errorf("some other error")
	result := Format(wrapped)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.Error() != wrapped.Error() {
		t.Errorf("expected passthrough, got %q", result.Error())
	}
}

func TestFormat_UnauthorizedError(t *testing.T) {
	err := fmt.Errorf("401 Unauthorized")
	result := Format(err)
	if result == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestFormat_ForbiddenError(t *testing.T) {
	err := fmt.Errorf("403 Forbidden")
	result := Format(err)
	if result == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestFormat_NotFoundError(t *testing.T) {
	err := fmt.Errorf("404 Not Found")
	result := Format(err)
	if result == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestFormat_GenericOpenAPIError(t *testing.T) {
	// Test that a regular error passes through
	err := errors.New("generic API error")
	result := Format(err)
	if result == nil {
		t.Fatal("expected non-nil")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
