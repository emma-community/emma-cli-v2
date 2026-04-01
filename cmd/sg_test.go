package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	emma "github.com/emma-community/emma-go-sdk"
)

func TestSGList(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/security-groups" {
			json.NewEncoder(w).Encode([]emma.SecurityGroup{
				{Id: int32Ptr(1), Name: strPtr("sg-test"), SynchronizationStatus: strPtr("SYNCHRONIZED")},
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newSGListCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("sg list: %v", err)
	}
	if !strings.Contains(cliOut(c), "sg-test") {
		t.Errorf("expected 'sg-test' in output")
	}
}

func TestSGGetShowsRules(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/security-groups/3" {
			json.NewEncoder(w).Encode(emma.SecurityGroup{
				Id:   int32Ptr(3),
				Name: strPtr("web-sg"),
				Rules: []emma.SecurityGroupRule{
					{Direction: strPtr("inbound"), Protocol: strPtr("tcp"), Ports: strPtr("80"), IpRange: strPtr("0.0.0.0/0")},
					{Direction: strPtr("inbound"), Protocol: strPtr("tcp"), Ports: strPtr("443"), IpRange: strPtr("0.0.0.0/0")},
				},
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newSGGetCmd()
	cmd.Flags().Set("id", "3")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("sg get: %v", err)
	}
	out := cliOut(c)
	if !strings.Contains(out, "web-sg") {
		t.Errorf("expected 'web-sg' in output")
	}
	if !strings.Contains(out, "Rules:") {
		t.Errorf("expected rules section in output, got: %s", out)
	}
	if !strings.Contains(out, "80") {
		t.Errorf("expected port 80 in rules, got: %s", out)
	}
}

func TestSGCreate(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/security-groups" {
			json.NewEncoder(w).Encode(emma.SecurityGroup{
				Id: int32Ptr(10), Name: strPtr("new-sg"),
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newSGCreateCmd()
	cmd.Flags().Set("name", "new-sg")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("sg create: %v", err)
	}
	if !strings.Contains(cliOut(c), "new-sg") {
		t.Errorf("expected 'new-sg' in output")
	}
}
