package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	emma "github.com/emma-community/emma-go-sdk"
)

func TestSubnetList(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/subnetworks" {
			json.NewEncoder(w).Encode([]emma.Subnetwork{
				{Id: strPtr("sub-1"), Name: strPtr("my-subnet"), Status: strPtr("ACTIVE")},
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newSubnetListCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("subnet list: %v", err)
	}
	if !strings.Contains(cliOut(c), "my-subnet") {
		t.Errorf("expected 'my-subnet' in output")
	}
}

func TestSubnetGet_DirectAPI(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/subnetworks/sub-42" && r.Method == "GET" {
			json.NewEncoder(w).Encode(emma.Subnetwork{
				Id: strPtr("sub-42"), Name: strPtr("direct-subnet"), Status: strPtr("ACTIVE"),
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newSubnetGetCmd()
	cmd.Flags().Set("id", "sub-42")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("subnet get: %v", err)
	}
	if !strings.Contains(cliOut(c), "direct-subnet") {
		t.Errorf("expected 'direct-subnet' in output")
	}
}

func TestSubnetCreate(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/subnetworks" {
			json.NewEncoder(w).Encode(emma.Subnetwork{
				Id: strPtr("sub-new"), Name: strPtr("created-subnet"), Status: strPtr("CREATING"),
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newSubnetCreateCmd()
	cmd.Flags().Set("datacenter-id", "dc-1")
	cmd.Flags().Set("size", "24")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("subnet create: %v", err)
	}
	if !strings.Contains(cliOut(c), "created-subnet") {
		t.Errorf("expected 'created-subnet' in output")
	}
}

func TestSubnetDelete_WithYes(t *testing.T) {
	deleted := false
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/v1/subnetworks/sub-1" {
			deleted = true
			json.NewEncoder(w).Encode(emma.Subnetwork{})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newSubnetDeleteCmd()
	cmd.Flags().Set("id", "sub-1")
	cmd.Flags().Set("yes", "true")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("subnet delete: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE to be called")
	}
}
