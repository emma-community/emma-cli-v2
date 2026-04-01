package cmd

import (
	"context"
	"fmt"

	emma "github.com/emma-community/emma-go-sdk"
	apierrors "github.com/emma-community/emma-cli/internal/apierrors"
	"github.com/emma-community/emma-cli/internal/cmdutil"
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
	cmd.AddCommand(c.newSubnetCreateCmd())
	cmd.AddCommand(c.newSubnetEditCmd())
	cmd.AddCommand(c.newSubnetDeleteCmd())
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
			req := c.Client.SubnetworksAPI.GetSubnetworks(ctx)
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
			subnet, _, err := c.Client.SubnetworksAPI.GetSubnetwork(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, subnet, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "DATACENTER", "CIDR"},
				Rows:    [][]string{subnetToRow(*subnet)},
			})
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Subnetwork ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (c *CLI) newSubnetCreateCmd() *cobra.Command {
	var name, datacenterID, prefix string
	var size int32

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a subnetwork",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			createReq := emma.SubnetworkCreate{
				DataCenterId:    datacenterID,
				SubnetworkSize:  size,
			}
			if name != "" {
				createReq.Name = &name
			}
			if prefix != "" {
				createReq.SubnetworkPrefix = &prefix
			}

			subnet, _, err := c.Client.SubnetworksAPI.SubnetworkCreate(ctx).SubnetworkCreate(createReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, subnet, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "DATACENTER", "CIDR"},
				Rows:    [][]string{subnetToRow(*subnet)},
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Subnetwork name")
	cmd.Flags().StringVar(&datacenterID, "datacenter-id", "", "Datacenter ID (required)")
	cmd.Flags().StringVar(&prefix, "prefix", "", "Subnetwork prefix (CIDR)")
	cmd.Flags().Int32Var(&size, "size", 0, "Subnetwork size (required)")
	_ = cmd.MarkFlagRequired("datacenter-id")
	_ = cmd.MarkFlagRequired("size")
	return cmd
}

func (c *CLI) newSubnetEditCmd() *cobra.Command {
	var id, name string

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a subnetwork",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			editReq := emma.SubnetworkEdit{}
			if name != "" {
				editReq.Name = &name
			}

			subnet, _, err := c.Client.SubnetworksAPI.SubnetworkUpdate(ctx, id).SubnetworkEdit(editReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, subnet, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "DATACENTER", "CIDR"},
				Rows:    [][]string{subnetToRow(*subnet)},
			})
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Subnetwork ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "New name")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (c *CLI) newSubnetDeleteCmd() *cobra.Command {
	var id string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a subnetwork",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmdutil.ConfirmDelete(c.Err, cmd.InOrStdin(), "subnetwork", id, yes) {
				fmt.Fprintln(c.Out, "Aborted.")
				return nil
			}

			ctx := context.Background()
			_, _, err := c.Client.SubnetworksAPI.SubnetworkDelete(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			fmt.Fprintf(c.Out, "Deleted subnetwork %s.\n", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Subnetwork ID")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}
