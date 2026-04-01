package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	emma "github.com/emma-community/emma-go-sdk"
)

func TestSpotList(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/spot-instances" {
			json.NewEncoder(w).Encode([]emma.SpotVm{
				{Id: int32Ptr(1), Name: strPtr("spot-1"), Status: strPtr("RUNNING")},
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newSpotListCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("spot list: %v", err)
	}
	if !strings.Contains(cliOut(c), "spot-1") {
		t.Errorf("expected 'spot-1' in output")
	}
}

func TestSpotActions_Start(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/spot-instances/5/actions" {
			json.NewEncoder(w).Encode(emma.SpotVm{
				Id: int32Ptr(5), Name: strPtr("started-spot"), Status: strPtr("RUNNING"),
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newSpotActionsCmd()
	cmd.Flags().Set("id", "5")
	cmd.Flags().Set("action", "start")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("spot actions start: %v", err)
	}
	if !strings.Contains(cliOut(c), "started-spot") {
		t.Errorf("expected 'started-spot' in output")
	}
}

func TestSpotActions_UnknownAction(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	cmd := c.newSpotActionsCmd()
	cmd.Flags().Set("id", "1")
	cmd.Flags().Set("action", "fly")
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
	if !strings.Contains(err.Error(), "unknown action") {
		t.Errorf("expected 'unknown action', got: %v", err)
	}
}

func TestSpotDelete_WithYes(t *testing.T) {
	deleted := false
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/v1/spot-instances/7" {
			deleted = true
			json.NewEncoder(w).Encode(map[string]any{})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newSpotDeleteCmd()
	cmd.Flags().Set("id", "7")
	cmd.Flags().Set("yes", "true")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("spot delete: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE to be called")
	}
}
