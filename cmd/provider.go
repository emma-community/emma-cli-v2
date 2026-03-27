package cmd

import (
	"context"
	"strconv"

	emma "github.com/emma-community/emma-go-sdk"
	apierrors "github.com/emma-community/emma-cli/internal/apierrors"
	"github.com/emma-community/emma-cli/internal/output"
	"github.com/spf13/cobra"
)

// newProviderCmd covers provider, location, datacenter, and os list commands.
func (c *CLI) newProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Query providers, locations, datacenters, and operating systems",
	}
	cmd.AddCommand(c.newProviderListCmd())
	cmd.AddCommand(c.newLocationListCmd())
	cmd.AddCommand(c.newDatacenterListCmd())
	cmd.AddCommand(c.newOSListCmd())
	return cmd
}

func (c *CLI) newProviderListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cloud providers",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			req := c.Client.ProvidersAPI.GetProviders(ctx)
			if c.ProjectID != nil {
				req = req.ProjectId(*c.ProjectID)
			}
			providers, _, err := req.Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			rows := make([][]string, 0, len(providers))
			for _, p := range providers {
				id := ""
				if p.Id != nil {
					id = strconv.Itoa(int(*p.Id))
				}
				rows = append(rows, []string{derefStr(p.Name), id})
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, providers, output.TableView{
				Headers: []string{"NAME", "ID"},
				Rows:    rows,
			})
		},
	}
	return cmd
}

func (c *CLI) newLocationListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "location-list",
		Short: "List available locations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			req := c.Client.LocationsAPI.GetLocations(ctx)
			if c.ProjectID != nil {
				req = req.ProjectId(*c.ProjectID)
			}
			locs, _, err := req.Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			rows := make([][]string, 0, len(locs))
			for _, l := range locs {
				id := ""
				if l.Id != nil {
					id = strconv.Itoa(int(*l.Id))
				}
				rows = append(rows, []string{derefStr(l.Name), id})
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, locs, output.TableView{
				Headers: []string{"NAME", "ID"},
				Rows:    rows,
			})
		},
	}
	return cmd
}

func (c *CLI) newDatacenterListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datacenter-list",
		Short: "List available datacenters",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			req := c.Client.DataCentersAPI.GetDataCenters(ctx)
			if c.ProjectID != nil {
				req = req.ProjectId(*c.ProjectID)
			}
			dcs, _, err := req.Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			rows := make([][]string, 0, len(dcs))
			for _, dc := range dcs {
				loc := derefStr(dc.LocationName)
				provider := derefStr(dc.ProviderName)
				rows = append(rows, []string{derefStr(dc.Name), derefStr(dc.Id), loc, provider})
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, dcs, output.TableView{
				Headers: []string{"NAME", "ID", "LOCATION", "PROVIDER"},
				Rows:    rows,
			})
		},
	}
	return cmd
}

func (c *CLI) newOSListCmd() *cobra.Command {
	var osType, arch string

	cmd := &cobra.Command{
		Use:   "os-list",
		Short: "List available operating systems",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			req := c.Client.OperatingSystemsAPI.GetOperatingSystems(ctx)
			if osType != "" {
				req = req.Type_(osType)
			}
			if arch != "" {
				req = req.Architecture(arch)
			}
			if c.ProjectID != nil {
				req = req.ProjectId(*c.ProjectID)
			}
			oses, _, err := req.Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			rows := make([][]string, 0, len(oses))
			for _, os := range oses {
				id := ""
				if os.Id != nil {
					id = strconv.Itoa(int(*os.Id))
				}
				rows = append(rows, []string{
					derefStr(os.Type),
					id,
					derefStr(os.Version),
					derefStr(os.Architecture),
				})
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, oses, output.TableView{
				Headers: []string{"OS", "ID", "VERSION", "ARCH"},
				Rows:    rows,
			})
		},
	}

	cmd.Flags().StringVar(&osType, "type", "", "Filter by OS type")
	cmd.Flags().StringVar(&arch, "arch", "", "Filter by architecture")
	return cmd
}

// Standalone top-level commands for location, datacenter, os
func newLocationCmd(c *CLI) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "location",
		Short: "Manage locations",
	}
	cmd.AddCommand(c.newLocationListTopCmd())
	return cmd
}

func (c *CLI) newLocationListTopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available locations",
		RunE:  c.newLocationListCmd().RunE,
	}
}

func providerToRow(p emma.Provider) []string {
	id := ""
	if p.Id != nil {
		id = strconv.Itoa(int(*p.Id))
	}
	return []string{derefStr(p.Name), id}
}
