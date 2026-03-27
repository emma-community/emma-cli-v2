package cmd

import (
	"io"
	"os"

	emma "github.com/emma-community/emma-go-sdk"
	"github.com/emma-community/emma-cli/internal/config"
)

// CLI holds all shared dependencies for the emma CLI.
// It is the central dependency injection container passed to all commands.
type CLI struct {
	Client    *emma.APIClient
	Cfg       *config.Config
	CfgPath   string
	Out       io.Writer // os.Stdout in prod, bytes.Buffer in tests
	Err       io.Writer // os.Stderr in prod
	OutputFmt string    // global --output flag value
	ProjectID *int32    // global --project-id flag value
	Debug     bool      // global --debug flag
	NoColor   bool      // global --no-color flag

	// Build info set at startup
	version string
	commit  string
	date    string
}

// NewCLI creates a CLI with default I/O writers and loaded config.
func NewCLI(version, commit, date string) *CLI {
	cfgPath := config.DefaultConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		// Non-fatal: use empty config and warn on first use
		cfg = &config.Config{}
	}
	return &CLI{
		Out:     os.Stdout,
		Err:     os.Stderr,
		Cfg:     cfg,
		CfgPath: cfgPath,
		version: version,
		commit:  commit,
		date:    date,
	}
}
