package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	emma "github.com/emma-community/emma-go-sdk"
)

func TestK8sList(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/kubernetes" {
			json.NewEncoder(w).Encode([]emma.KubernetesListResponseInner{
				{
					Id:                 int32Ptr(1),
					Name:               strPtr("test-cluster"),
					Status:             strPtr("ACTIVE"),
					Version:            strPtr("1.28"),
					DeploymentLocation: strPtr("EU"),
					K8sConnectionType:  strPtr("InternetConnect"),
				},
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newK8sListCmd()
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("k8s list: %v", err)
	}
	out := cliOut(c)
	if !strings.Contains(out, "test-cluster") {
		t.Errorf("expected 'test-cluster', got: %s", out)
	}
	if !strings.Contains(out, "ACTIVE") {
		t.Errorf("expected 'ACTIVE' status, got: %s", out)
	}
}

func TestK8sGet(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/kubernetes/5" {
			json.NewEncoder(w).Encode(emma.KubernetesGetResponse{
				Id:                 int32Ptr(5),
				Name:               strPtr("my-cluster"),
				Status:             strPtr("ACTIVE"),
				Version:            strPtr("1.29"),
				DeploymentLocation: strPtr("US"),
				K8sConnectionType:  strPtr("DirectConnect"),
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newK8sGetCmd()
	cmd.Flags().Set("id", "5")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("k8s get: %v", err)
	}
	out := cliOut(c)
	if !strings.Contains(out, "my-cluster") {
		t.Errorf("expected 'my-cluster', got: %s", out)
	}
}

func TestK8sCreate(t *testing.T) {
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/v1/kubernetes" {
			var req emma.KubernetesCreateRequest
			json.NewDecoder(r.Body).Decode(&req)

			if req.Name != "new-cluster" {
				t.Errorf("expected name 'new-cluster', got %q", req.Name)
			}
			if req.DeploymentLocation != "eu" {
				t.Errorf("expected deployment location 'eu', got %q", req.DeploymentLocation)
			}
			if req.K8sConnectionType != "internet_connect" {
				t.Errorf("expected connection type 'internet_connect', got %q", req.K8sConnectionType)
			}

			json.NewEncoder(w).Encode(emma.KubernetesCreateResponse{
				Id:                 int32Ptr(10),
				Name:               strPtr("new-cluster"),
				Status:             strPtr("CREATING"),
				K8sConnectionType:  strPtr("internet_connect"),
				DeploymentLocation: strPtr("eu"),
			})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newK8sCreateCmd()
	cmd.Flags().Set("name", "new-cluster")
	cmd.Flags().Set("deployment-location", "EU")
	cmd.Flags().Set("connection-type", "InternetConnect")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("k8s create: %v", err)
	}
	if !strings.Contains(cliOut(c), "new-cluster") {
		t.Errorf("expected 'new-cluster' in output")
	}
}

func TestNormalizeDeploymentLocation(t *testing.T) {
	cases := []struct{ in, want string }{
		{"EU", "eu"}, {"eu", "eu"}, {"Us", "us"}, {"APAC", "apac"},
	}
	for _, tc := range cases {
		if got := normalizeDeploymentLocation(tc.in); got != tc.want {
			t.Errorf("normalizeDeploymentLocation(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeConnectionType(t *testing.T) {
	cases := []struct{ in, want string }{
		{"InternetConnect", "internet_connect"},
		{"DirectConnect", "direct_connect"},
		{"internet_connect", "internet_connect"},
		{"direct_connect", "direct_connect"},
		{"INTERNETCONNECT", "internet_connect"},
		{"DIRECTCONNECT", "direct_connect"},
	}
	for _, tc := range cases {
		if got := normalizeConnectionType(tc.in); got != tc.want {
			t.Errorf("normalizeConnectionType(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestK8sDelete_WithYes(t *testing.T) {
	deleted := false
	c, _ := newTestCLI(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/v1/kubernetes/3" {
			deleted = true
			json.NewEncoder(w).Encode(emma.KubernetesDelete{})
			return
		}
		http.NotFound(w, r)
	}))

	cmd := c.newK8sDeleteCmd()
	cmd.Flags().Set("id", "3")
	cmd.Flags().Set("yes", "true")
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("k8s delete: %v", err)
	}
	if !deleted {
		t.Error("expected DELETE to be called")
	}
}
