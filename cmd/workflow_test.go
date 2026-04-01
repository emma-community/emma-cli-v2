package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	emma "github.com/emma-community/emma-go-sdk"
)

func TestWorkflowList(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/workflow-templates" && r.Method == "GET" {
			lastPage := true
			json.NewEncoder(w).Encode(emma.GetWorkflowTemplates200Response{
				Content: []emma.WorkflowTemplate{
					{Id: 1, Name: "deploy-template", Status: "ACTIVE", ResourceType: "vm", ContentType: "script", CreatedByName: "admin"},
				},
				Last: &lastPage,
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newWorkflowListCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("workflow list: %v", err)
	}
	out := cliOut(c)
	if !strings.Contains(out, "deploy-template") {
		t.Errorf("expected 'deploy-template' in output, got: %s", out)
	}
}

func TestWorkflowCreate(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/workflow-templates" {
			json.NewEncoder(w).Encode(emma.WorkflowTemplate{
				Id: 5, Name: "new-wf", Status: "ACTIVE", ResourceType: "vm", ContentType: "script", CreatedByName: "test",
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newWorkflowCreateCmd()
	cmd.Flags().Set("name", "new-wf")
	cmd.Flags().Set("content-type", "script")
	cmd.Flags().Set("content", "echo hello")
	cmd.Flags().Set("status", "ACTIVE")
	cmd.Flags().Set("resource-type", "vm")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("workflow create: %v", err)
	}
	if !strings.Contains(cliOut(c), "new-wf") {
		t.Errorf("expected 'new-wf' in output")
	}
}

func TestWorkflowCreate_InvalidJSON(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	cmd := c.newWorkflowCreateCmd()
	cmd.Flags().Set("name", "test")
	cmd.Flags().Set("content-type", "script")
	cmd.Flags().Set("content", "x")
	cmd.Flags().Set("status", "ACTIVE")
	cmd.Flags().Set("resource-type", "vm")
	cmd.Flags().Set("content-params", "not-json")
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid --content-params JSON") {
		t.Errorf("expected JSON parse error, got: %v", err)
	}
}

func TestWorkflowDelete_WithYes(t *testing.T) {
	deleted := false
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/v1/workflow-templates/2" {
			deleted = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newWorkflowDeleteCmd()
	cmd.Flags().Set("id", "2")
	cmd.Flags().Set("yes", "true")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("workflow delete: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE to be called")
	}
}
