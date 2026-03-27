// Package output provides rendering utilities for CLI output.
//
// Render flow:
//
//	data any ──► format? ──► "json"  ──► json.NewEncoder(w).Encode(data)
//	                     ├─► "yaml"  ──► yaml.Marshal(data) → w
//	                     └─► "table" ──► TableView.Rows → tablewriter → w
//	                                         │
//	                                         └─► colorize STATUS col if TTY && !noColor
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// TableView holds the headers and rows for table rendering.
type TableView struct {
	Headers []string
	Rows    [][]string
}

// statusColorMap maps status strings to color attributes.
var statusColorMap = map[string]*color.Color{
	"RUNNING":  color.New(color.FgGreen),
	"ACTIVE":   color.New(color.FgGreen),
	"STOPPED":  color.New(color.FgYellow),
	"ERROR":    color.New(color.FgRed),
	"CREATING": color.New(color.FgCyan),
}

// statusColIndex finds the index of the STATUS column in headers.
func statusColIndex(headers []string) int {
	for i, h := range headers {
		if strings.ToUpper(h) == "STATUS" {
			return i
		}
	}
	return -1
}

// colorizeStatus applies color to a status string value.
func colorizeStatus(status string) string {
	c, ok := statusColorMap[strings.ToUpper(status)]
	if !ok {
		return status
	}
	return c.Sprint(status)
}

// Render outputs data in the requested format to w.
// - table: uses tablewriter, colorizes STATUS column when !noColor and stdout is a TTY
// - json:  json.NewEncoder(w).Encode(data) — typed struct, not string rows
// - yaml:  yaml.v3 Marshal
// - zero rows/nil data: prints "No items found." for table, `[]` for json
func Render(w io.Writer, format string, noColor bool, data any, table TableView) error {
	switch strings.ToLower(format) {
	case "json":
		if data == nil {
			_, err := fmt.Fprintln(w, "[]")
			return err
		}
		return json.NewEncoder(w).Encode(data)

	case "yaml":
		if data == nil {
			_, err := fmt.Fprintln(w, "[]")
			return err
		}
		out, err := yaml.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshaling yaml: %w", err)
		}
		_, err = w.Write(out)
		return err

	default: // table
		if len(table.Rows) == 0 {
			_, err := fmt.Fprintln(w, "No items found.")
			return err
		}

		isTTY := term.IsTerminal(int(os.Stdout.Fd()))
		useColor := !noColor && isTTY && !color.NoColor
		statusIdx := statusColIndex(table.Headers)

		tw := tablewriter.NewWriter(w)
		tw.SetHeader(table.Headers)
		tw.SetBorder(false)
		tw.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		tw.SetAlignment(tablewriter.ALIGN_LEFT)
		tw.SetHeaderLine(false)
		tw.SetColumnSeparator("  ")
		tw.SetTablePadding("  ")
		tw.SetNoWhiteSpace(true)
		tw.SetAutoFormatHeaders(false)

		for _, row := range table.Rows {
			if useColor && statusIdx >= 0 && statusIdx < len(row) {
				rowCopy := make([]string, len(row))
				copy(rowCopy, row)
				rowCopy[statusIdx] = colorizeStatus(row[statusIdx])
				tw.Append(rowCopy)
			} else {
				tw.Append(row)
			}
		}
		tw.Render()
		return nil
	}
}
