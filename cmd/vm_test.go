package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	emma "github.com/emma-community/emma-go-sdk"
	"github.com/emma-community/emma-cli/internal/config"
)

func newTestCLI(t *testing.T, handler http.Handler) (*CLI, *httptest.Server) {
	t.Helper()
	// Wrap handler to set Content-Type header — SDK requires it for response parsing
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler.ServeHTTP(w, r)
	})
	srv := httptest.NewServer(wrapped)
	t.Cleanup(srv.Close)

	cfg := emma.NewConfiguration()
	cfg.Servers = emma.ServerConfigurations{{URL: srv.URL}}
	cfg.AddDefaultHeader("Authorization", "Bearer test-token")
	client := emma.NewAPIClient(cfg)

	cli := &CLI{
		Client:    client,
		Cfg:       &config.Config{},
		Out:       &bytes.Buffer{},
		Err:       &bytes.Buffer{},
		OutputFmt: "table",
	}
	return cli, srv
}

func cliOut(c *CLI) string  { return c.Out.(*bytes.Buffer).String() }
func cliErr(c *CLI) string  { return c.Err.(*bytes.Buffer).String() }
func strPtr(s string) *string  { return &s }
func int32Ptr(i int32) *int32  { return &i }
func float32Ptr(f float32) *float32 { return &f }

func TestVMList(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/vms" {
			json.NewEncoder(w).Encode([]emma.Vm{
				{Id: int32Ptr(1), Name: strPtr("test-vm"), Status: strPtr("RUNNING")},
				{Id: int32Ptr(2), Name: strPtr("other-vm"), Status: strPtr("STOPPED")},
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newVMListCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("vm list: %v", err)
	}
	out := cliOut(c)
	if !strings.Contains(out, "test-vm") {
		t.Errorf("expected 'test-vm', got: %s", out)
	}
	if !strings.Contains(out, "other-vm") {
		t.Errorf("expected 'other-vm', got: %s", out)
	}
}

func TestVMListStatusFilter(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]emma.Vm{
			{Id: int32Ptr(1), Name: strPtr("running-vm"), Status: strPtr("RUNNING")},
			{Id: int32Ptr(2), Name: strPtr("stopped-vm"), Status: strPtr("STOPPED")},
		})
	}))

	cmd := c.newVMListCmd()
	cmd.Flags().Set("status", "RUNNING")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("vm list --status: %v", err)
	}
	out := cliOut(c)
	if !strings.Contains(out, "running-vm") {
		t.Errorf("expected 'running-vm' in output")
	}
	if strings.Contains(out, "stopped-vm") {
		t.Errorf("'stopped-vm' should be filtered out")
	}
}

func TestVMGet(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/vms/42" {
			json.NewEncoder(w).Encode(emma.Vm{
				Id: int32Ptr(42), Name: strPtr("my-vm"), Status: strPtr("RUNNING"),
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newVMGetCmd()
	cmd.Flags().Set("id", "42")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("vm get: %v", err)
	}
	if !strings.Contains(cliOut(c), "my-vm") {
		t.Errorf("expected 'my-vm' in output")
	}
}

func TestVMDelete_WithYes(t *testing.T) {
	deleted := false
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/v1/vms/5" {
			deleted = true
			json.NewEncoder(w).Encode(map[string]any{})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newVMDeleteCmd()
	cmd.Flags().Set("id", "5")
	cmd.Flags().Set("yes", "true")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("vm delete: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE to be called")
	}
	if !strings.Contains(cliOut(c), "Deleted VM 5") {
		t.Errorf("unexpected output: %s", cliOut(c))
	}
}

func TestVMActions_Start(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/vms/10/actions" {
			json.NewEncoder(w).Encode(emma.Vm{
				Id: int32Ptr(10), Name: strPtr("started-vm"), Status: strPtr("RUNNING"),
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newVMActionsCmd()
	cmd.Flags().Set("id", "10")
	cmd.Flags().Set("action", "start")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("vm actions start: %v", err)
	}
	if !strings.Contains(cliOut(c), "started-vm") {
		t.Errorf("expected 'started-vm' in output, got: %s", cliOut(c))
	}
}

func TestVMActions_UnknownAction(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	cmd := c.newVMActionsCmd()
	cmd.Flags().Set("id", "1")
	cmd.Flags().Set("action", "explode")
	err := cmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
	if !strings.Contains(err.Error(), "unknown action") {
		t.Errorf("expected 'unknown action' error, got: %v", err)
	}
}

func TestVMGetJSON(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(emma.Vm{
			Id: int32Ptr(1), Name: strPtr("json-vm"), Status: strPtr("RUNNING"),
		})
	}))
	c.OutputFmt = "json"

	cmd := c.newVMGetCmd()
	cmd.Flags().Set("id", "1")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("vm get --output json: %v", err)
	}
	out := cliOut(c)
	if !strings.Contains(out, `"name":"json-vm"`) {
		t.Errorf("expected JSON output with 'json-vm', got: %s", out)
	}
}
