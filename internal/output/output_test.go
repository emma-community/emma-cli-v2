package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRender_TableFormat(t *testing.T) {
	var buf bytes.Buffer
	tv := TableView{
		Headers: []string{"NAME", "ID", "STATUS"},
		Rows: [][]string{
			{"vm-1", "101", "RUNNING"},
			{"vm-2", "102", "STOPPED"},
		},
	}

	err := Render(&buf, "table", true, nil, tv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "NAME") {
		t.Errorf("expected NAME header in output, got:\n%s", out)
	}
	if !strings.Contains(out, "vm-1") {
		t.Errorf("expected vm-1 in output, got:\n%s", out)
	}
	if !strings.Contains(out, "RUNNING") {
		t.Errorf("expected RUNNING in output, got:\n%s", out)
	}
}

type testItem struct {
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
}

func TestRender_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	items := []testItem{
		{Name: "vm-1", Status: "RUNNING"},
		{Name: "vm-2", Status: "STOPPED"},
	}

	tv := TableView{
		Headers: []string{"NAME", "STATUS"},
		Rows:    [][]string{{"vm-1", "RUNNING"}, {"vm-2", "STOPPED"}},
	}

	err := Render(&buf, "json", true, items, tv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []testItem
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, buf.String())
	}

	if len(decoded) != 2 {
		t.Errorf("expected 2 items, got %d", len(decoded))
	}
	if decoded[0].Name != "vm-1" {
		t.Errorf("expected name=vm-1, got %q", decoded[0].Name)
	}
	// Verify typed struct — not string arrays
	if decoded[0].Status != "RUNNING" {
		t.Errorf("expected status=RUNNING, got %q", decoded[0].Status)
	}
}

func TestRender_YAMLFormat(t *testing.T) {
	var buf bytes.Buffer
	items := []testItem{
		{Name: "vm-1", Status: "RUNNING"},
	}
	tv := TableView{}

	err := Render(&buf, "yaml", true, items, tv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []testItem
	if err := yaml.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid YAML: %v\nOutput: %s", err, buf.String())
	}
	if len(decoded) != 1 || decoded[0].Name != "vm-1" {
		t.Errorf("unexpected decoded value: %+v", decoded)
	}
}

func TestRender_EmptyData_Table(t *testing.T) {
	var buf bytes.Buffer
	tv := TableView{
		Headers: []string{"NAME"},
		Rows:    [][]string{},
	}

	err := Render(&buf, "table", true, nil, tv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No items found.") {
		t.Errorf("expected 'No items found.', got:\n%s", out)
	}
}

func TestRender_EmptyData_JSON(t *testing.T) {
	var buf bytes.Buffer
	tv := TableView{}

	err := Render(&buf, "json", true, nil, tv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := strings.TrimSpace(buf.String())
	if out != "[]" {
		t.Errorf("expected '[]', got %q", out)
	}
}

func TestRender_DefaultIsTable(t *testing.T) {
	var buf bytes.Buffer
	tv := TableView{
		Headers: []string{"NAME"},
		Rows:    [][]string{{"vm-1"}},
	}

	err := Render(&buf, "table", true, nil, tv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty table output")
	}
}
