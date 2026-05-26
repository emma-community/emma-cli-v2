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

func TestSpotCreate_WithSSHKeyID(t *testing.T) {
	var receivedBody emma.SpotCreate
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/spot-instances" {
			json.NewDecoder(r.Body).Decode(&receivedBody)
			json.NewEncoder(w).Encode(emma.SpotVm{
				Id: int32Ptr(10), Name: strPtr("test-spot"), Status: strPtr("POWERED_ON"),
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newSpotCreateCmd()
	cmd.Flags().Set("name", "test-spot")
	cmd.Flags().Set("datacenter-id", "aws-us-west-1")
	cmd.Flags().Set("os-id", "1")
	cmd.Flags().Set("vcpu", "2")
	cmd.Flags().Set("ram", "4")
	cmd.Flags().Set("volume-size", "20")
	cmd.Flags().Set("volume-type", "ssd")
	cmd.Flags().Set("cloud-network-type", "multi-cloud")
	cmd.Flags().Set("price", "0.5")
	cmd.Flags().Set("ssh-key-id", "42")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("spot create: %v", err)
	}
	if receivedBody.SshKeyId == nil || *receivedBody.SshKeyId != 42 {
		t.Errorf("expected SshKeyId=42, got %v", receivedBody.SshKeyId)
	}
	if !strings.Contains(cliOut(c), "test-spot") {
		t.Errorf("expected 'test-spot' in output")
	}
}

func TestSpotCreate_WithoutSSHKeyID(t *testing.T) {
	var receivedBody emma.SpotCreate
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/spot-instances" {
			json.NewDecoder(r.Body).Decode(&receivedBody)
			json.NewEncoder(w).Encode(emma.SpotVm{
				Id: int32Ptr(11), Name: strPtr("no-key-spot"), Status: strPtr("POWERED_ON"),
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newSpotCreateCmd()
	cmd.Flags().Set("name", "no-key-spot")
	cmd.Flags().Set("datacenter-id", "aws-us-west-1")
	cmd.Flags().Set("os-id", "1")
	cmd.Flags().Set("vcpu", "2")
	cmd.Flags().Set("ram", "4")
	cmd.Flags().Set("volume-size", "20")
	cmd.Flags().Set("volume-type", "ssd")
	cmd.Flags().Set("cloud-network-type", "multi-cloud")
	cmd.Flags().Set("price", "0.5")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("spot create without ssh-key-id: %v", err)
	}
	if receivedBody.SshKeyId != nil {
		t.Errorf("expected SshKeyId to be nil, got %v", *receivedBody.SshKeyId)
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
