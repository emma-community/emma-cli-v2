// Package testutil provides shared test helpers for emma CLI command tests.
package testutil

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	emma "github.com/emma-community/emma-go-sdk"
	"github.com/emma-community/emma-cli/cmd"
	"github.com/emma-community/emma-cli/internal/config"
)

// TestCLI holds a configured CLI and mock server for testing.
type TestCLI struct {
	CLI    *cmd.CLI
	Server *httptest.Server
	Out    *bytes.Buffer
	Err    *bytes.Buffer
}

// NewTestCLI creates a CLI backed by a mock HTTP server.
// The handler receives all API requests. Call cleanup when done.
func NewTestCLI(t *testing.T, handler http.Handler) *TestCLI {
	t.Helper()
	srv := httptest.NewServer(handler)

	cfg := emma.NewConfiguration()
	cfg.Servers = emma.ServerConfigurations{
		{URL: srv.URL},
	}
	cfg.AddDefaultHeader("Authorization", "Bearer test-token")
	client := emma.NewAPIClient(cfg)

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	cli := &cmd.CLI{
		Client:    client,
		Cfg:       &config.Config{},
		Out:       outBuf,
		Err:       errBuf,
		OutputFmt: "table",
	}

	t.Cleanup(srv.Close)

	return &TestCLI{
		CLI:    cli,
		Server: srv,
		Out:    outBuf,
		Err:    errBuf,
	}
}
