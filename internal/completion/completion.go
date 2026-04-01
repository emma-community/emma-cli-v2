// Package completion provides dynamic shell completion helpers for the emma CLI.
//
// Completions query the API for valid values and cache results in ~/.emma/cache/
// with a configurable TTL to avoid hitting the API on every tab press.
package completion

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	emma "github.com/emma-community/emma-go-sdk"
	"github.com/spf13/cobra"
)

const cacheTTL = 1 * time.Hour

var cacheDir = filepath.Join(os.Getenv("HOME"), ".emma", "cache")

type cachedEntry struct {
	Entries   []string  `json:"entries"`
	CachedAt time.Time `json:"cached_at"`
}

func readCache(key string) ([]string, bool) {
	data, err := os.ReadFile(filepath.Join(cacheDir, key+".json"))
	if err != nil {
		return nil, false
	}
	var entry cachedEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}
	if time.Since(entry.CachedAt) > cacheTTL {
		return nil, false
	}
	return entry.Entries, true
}

func writeCache(key string, entries []string) {
	_ = os.MkdirAll(cacheDir, 0700)
	entry := cachedEntry{Entries: entries, CachedAt: time.Now()}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(cacheDir, key+".json"), data, 0600)
}

// DatacenterIDs returns a completion function for datacenter IDs.
func DatacenterIDs(client *emma.APIClient) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if entries, ok := readCache("datacenters"); ok {
			return entries, cobra.ShellCompDirectiveNoFileComp
		}

		dcs, _, err := client.DataCentersAPI.GetDataCenters(context.Background()).Execute()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		entries := make([]string, 0, len(dcs))
		for _, dc := range dcs {
			if dc.Id != nil {
				name := ""
				if dc.Name != nil {
					name = *dc.Name
				}
				entries = append(entries, fmt.Sprintf("%s\t%s", *dc.Id, name))
			}
		}
		writeCache("datacenters", entries)
		return entries, cobra.ShellCompDirectiveNoFileComp
	}
}

// OSIDs returns a completion function for operating system IDs.
func OSIDs(client *emma.APIClient) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if entries, ok := readCache("os"); ok {
			return entries, cobra.ShellCompDirectiveNoFileComp
		}

		oses, _, err := client.OperatingSystemsAPI.GetOperatingSystems(context.Background()).Execute()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		entries := make([]string, 0, len(oses))
		for _, o := range oses {
			if o.Id != nil {
				desc := ""
				if o.Type != nil {
					desc = *o.Type
				}
				if o.Version != nil {
					desc += " " + *o.Version
				}
				entries = append(entries, fmt.Sprintf("%d\t%s", *o.Id, desc))
			}
		}
		writeCache("os", entries)
		return entries, cobra.ShellCompDirectiveNoFileComp
	}
}

// ProviderNames returns a completion function for provider names.
func ProviderNames(client *emma.APIClient) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if entries, ok := readCache("providers"); ok {
			return entries, cobra.ShellCompDirectiveNoFileComp
		}

		providers, _, err := client.ProvidersAPI.GetProviders(context.Background()).Execute()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		entries := make([]string, 0, len(providers))
		for _, p := range providers {
			if p.Name != nil {
				id := ""
				if p.Id != nil {
					id = strconv.Itoa(int(*p.Id))
				}
				entries = append(entries, fmt.Sprintf("%s\t(ID: %s)", *p.Name, id))
			}
		}
		writeCache("providers", entries)
		return entries, cobra.ShellCompDirectiveNoFileComp
	}
}
