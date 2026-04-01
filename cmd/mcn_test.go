package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	emma "github.com/emma-community/emma-go-sdk"
)

func TestMCNList(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/multi-cloud-networks" && r.Method == "GET" {
			json.NewEncoder(w).Encode(emma.MultiCloudNetwork{
				DirectConnect: &emma.MultiCloudNetworkDirectConnect{
					Networks: []emma.MultiCloudNetworkDirectConnectNetworksInner{
						{MacroRegion: strPtr("EMEA"), Status: strPtr("ACTIVE")},
						{MacroRegion: strPtr("AMER"), Status: strPtr("DISABLED")},
					},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newMCNListCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("mcn list: %v", err)
	}
	out := cliOut(c)
	if !strings.Contains(out, "EMEA") {
		t.Errorf("expected 'EMEA' in output, got: %s", out)
	}
	if !strings.Contains(out, "AMER") {
		t.Errorf("expected 'AMER' in output, got: %s", out)
	}
}

func TestMCNListEmpty(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(emma.MultiCloudNetwork{})
	}))

	cmd := c.newMCNListCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("mcn list empty: %v", err)
	}
	if !strings.Contains(cliOut(c), "No items found") {
		t.Errorf("expected 'No items found' for empty MCN")
	}
}

func TestMCNEnableNetwork(t *testing.T) {
	actionCalled := false
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/multi-cloud-networks/actions" {
			actionCalled = true
			json.NewEncoder(w).Encode([]emma.MultiCloudNetwork{{}})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newMCNEnableNetworkCmd()
	cmd.Flags().Set("macro-region", "EMEA")
	cmd.Flags().Set("connectivity-center-id", "1")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("mcn enable-network: %v", err)
	}
	if !actionCalled {
		t.Error("expected POST to be called")
	}
}
