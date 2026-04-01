package cmdutil

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfirmDelete_Yes(t *testing.T) {
	w := &bytes.Buffer{}
	r := strings.NewReader("y\n")
	if !ConfirmDelete(w, r, "VM", 42, false) {
		t.Error("expected true for 'y' input")
	}
	if !strings.Contains(w.String(), "Delete VM 42?") {
		t.Errorf("expected prompt, got: %s", w.String())
	}
}

func TestConfirmDelete_No(t *testing.T) {
	w := &bytes.Buffer{}
	r := strings.NewReader("n\n")
	if ConfirmDelete(w, r, "VM", 42, false) {
		t.Error("expected false for 'n' input")
	}
}

func TestConfirmDelete_EmptyInput(t *testing.T) {
	w := &bytes.Buffer{}
	r := strings.NewReader("\n")
	if ConfirmDelete(w, r, "VM", 42, false) {
		t.Error("expected false for empty input (defaults to No)")
	}
}

func TestConfirmDelete_SkipWithYes(t *testing.T) {
	w := &bytes.Buffer{}
	r := strings.NewReader("")
	if !ConfirmDelete(w, r, "VM", 42, true) {
		t.Error("expected true when yes=true")
	}
	if w.Len() > 0 {
		t.Error("expected no prompt when yes=true")
	}
}
