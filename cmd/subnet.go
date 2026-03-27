package cmd

import (
	"context"
	"fmt"

	emma "github.com/emma-community/emma-go-sdk"
	apierrors "github.com/emma-community/emma-cli/internal/apierrors"
	"github.com/emma-community/emma-cli/internal/output"
	"github.com/spf13/cobra"
)

func (c *CLI) newSubnetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subnet",
		Short: "Manage subnetworks",
	}
	cmd.AddCommand(c.newSubnetListCmd())
	cmd.AddCommand(c.newSubnetGetCmd())
	return cmd
}

func subnetToRow(s emma.Subnetwork) []string {
	id := derefStr(s.Id)
	name := derefStr(s.Name)
	status := derefStr(s.Status)
	dc := derefStr(s.DataCenterId)
	prefix := derefStr(s.SubnetworkPrefix)
	size := ""
	if s.SubnetworkSize != nil {
		size = fmt.Sprintf("/%d", *s.SubnetworkSize)
	}
	return []string{name, id, status, dc, prefix + size}
}

func (c *CLI) newSubnetListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subnetworks",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			req := c.Client.SubnetworksAPI.V1SubnetworksGet(ctx)
			if c.ProjectID != nil {
				req = req.ProjectId(*c.ProjectID)
			}
			subnets, _, err := req.Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			rows := make([][]string, 0, len(subnets))
			for _, s := range subnets {
				rows = append(rows, subnetToRow(s))
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, subnets, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "DATACENTER", "CIDR"},
				Rows:    rows,
			})
		},
	}
	return cmd
}

func (c *CLI) newSubnetGetCmd() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a subnetwork by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			// List all and filter — SDK may not have a get-by-id for subnets
			subnets, _, err := c.Client.SubnetworksAPI.V1SubnetworksGet(ctx).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			for _, s := range subnets {
				if derefStr(s.Id) == id {
					return output.Render(c.Out, c.OutputFmt, c.NoColor, s, output.TableView{
						Headers: []string{"NAME", "ID", "STATUS", "DATACENTER", "CIDR"},
						Rows:    [][]string{subnetToRow(s)},
					})
				}
			}
			return fmt.Errorf("subnetwork %q not found", id)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Subnetwork ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}
