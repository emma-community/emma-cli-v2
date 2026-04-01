package cmd

import (
	"context"
	"fmt"
	"strconv"

	emma "github.com/emma-community/emma-go-sdk"
	apierrors "github.com/emma-community/emma-cli/internal/apierrors"
	"github.com/emma-community/emma-cli/internal/cmdutil"
	"github.com/emma-community/emma-cli/internal/output"
	"github.com/spf13/cobra"
)

func (c *CLI) newSGCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sg",
		Short: "Manage security groups",
	}
	cmd.AddCommand(c.newSGListCmd())
	cmd.AddCommand(c.newSGGetCmd())
	cmd.AddCommand(c.newSGCreateCmd())
	cmd.AddCommand(c.newSGUpdateCmd())
	cmd.AddCommand(c.newSGDeleteCmd())
	cmd.AddCommand(c.newSGInstanceAddCmd())
	cmd.AddCommand(c.newSGInstanceListCmd())
	return cmd
}

func sgToRow(sg emma.SecurityGroup) []string {
	name := derefStr(sg.Name)
	id := ""
	if sg.Id != nil {
		id = strconv.Itoa(int(*sg.Id))
	}
	syncStatus := derefStr(sg.SynchronizationStatus)
	rules := strconv.Itoa(len(sg.Rules))
	return []string{name, id, syncStatus, rules}
}

func sgRuleToRow(r emma.SecurityGroupRule) []string {
	direction := derefStr(r.Direction)
	protocol := derefStr(r.Protocol)
	ports := derefStr(r.Ports)
	ip := derefStr(r.IpRange)
	policy := derefStr(r.Policy)
	return []string{direction, protocol, ports, ip, policy}
}

func (c *CLI) newSGListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List security groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			req := c.Client.SecurityGroupsAPI.GetSecurityGroups(ctx)
			if c.ProjectID != nil {
				req = req.ProjectId(*c.ProjectID)
			}
			sgs, _, err := req.Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			rows := make([][]string, 0, len(sgs))
			for _, sg := range sgs {
				rows = append(rows, sgToRow(sg))
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, sgs, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "RULES"},
				Rows:    rows,
			})
		},
	}
	return cmd
}

func (c *CLI) newSGGetCmd() *cobra.Command {
	var id int32

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a security group by ID (includes rules)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			sg, _, err := c.Client.SecurityGroupsAPI.GetSecurityGroup(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			if err := output.Render(c.Out, c.OutputFmt, c.NoColor, sg, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "RULES"},
				Rows:    [][]string{sgToRow(*sg)},
			}); err != nil {
				return err
			}

			// Show rules inline for table/wide output
			if c.OutputFmt == "table" || c.OutputFmt == "wide" {
				if len(sg.Rules) > 0 {
					fmt.Fprintln(c.Out)
					fmt.Fprintln(c.Out, "Rules:")
					ruleRows := make([][]string, 0, len(sg.Rules))
					for _, r := range sg.Rules {
						ruleRows = append(ruleRows, sgRuleToRow(r))
					}
					return output.Render(c.Out, "table", c.NoColor, nil, output.TableView{
						Headers: []string{"DIRECTION", "PROTOCOL", "PORTS", "IP-RANGE", "POLICY"},
						Rows:    ruleRows,
					})
				}
			}
			return nil
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "Security group ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (c *CLI) newSGCreateCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a security group",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			createReq := emma.SecurityGroupRequest{
				Name: name,
			}

			sg, _, err := c.Client.SecurityGroupsAPI.SecurityGroupCreate(ctx).SecurityGroupRequest(createReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, sg, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "RULES"},
				Rows:    [][]string{sgToRow(*sg)},
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Security group name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func (c *CLI) newSGUpdateCmd() *cobra.Command {
	var id int32
	var name string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a security group",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			updateReq := emma.SecurityGroupRequest{
				Name: name,
			}

			sg, _, err := c.Client.SecurityGroupsAPI.SecurityGroupUpdate(ctx, id).SecurityGroupRequest(updateReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, sg, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "RULES"},
				Rows:    [][]string{sgToRow(*sg)},
			})
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "Security group ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "New name (required)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func (c *CLI) newSGDeleteCmd() *cobra.Command {
	var id int32
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a security group",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmdutil.ConfirmDelete(c.Err, cmd.InOrStdin(), "security group", id, yes) {
				fmt.Fprintln(c.Out, "Aborted.")
				return nil
			}

			ctx := context.Background()
			_, _, err := c.Client.SecurityGroupsAPI.SecurityGroupDelete(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			fmt.Fprintf(c.Out, "Deleted security group %d.\n", id)
			return nil
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "Security group ID")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (c *CLI) newSGInstanceAddCmd() *cobra.Command {
	var sgID, instanceID int32

	cmd := &cobra.Command{
		Use:   "instance-add",
		Short: "Add a compute instance to a security group",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			addReq := emma.SecurityGroupInstanceAdd{
				InstanceId: &instanceID,
			}

			_, _, err := c.Client.SecurityGroupsAPI.SecurityGroupInstanceAdd(ctx, sgID).SecurityGroupInstanceAdd(addReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			fmt.Fprintf(c.Out, "Added instance %d to security group %d.\n", instanceID, sgID)
			return nil
		},
	}

	cmd.Flags().Int32Var(&sgID, "sg-id", 0, "Security group ID (required)")
	cmd.Flags().Int32Var(&instanceID, "instance-id", 0, "Compute instance ID (required)")
	_ = cmd.MarkFlagRequired("sg-id")
	_ = cmd.MarkFlagRequired("instance-id")
	return cmd
}

func (c *CLI) newSGInstanceListCmd() *cobra.Command {
	var sgID int32

	cmd := &cobra.Command{
		Use:   "instance-list",
		Short: "List compute instances in a security group",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			instances, _, err := c.Client.SecurityGroupsAPI.SecurityGroupInstances(ctx, sgID).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			rows := make([][]string, 0, len(instances))
			for _, inst := range instances {
				rows = append(rows, vmToRow(inst))
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, instances, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "OS", "VCPU", "RAM(GB)", "DATACENTER", "IP"},
				Rows:    rows,
			})
		},
	}

	cmd.Flags().Int32Var(&sgID, "sg-id", 0, "Security group ID (required)")
	_ = cmd.MarkFlagRequired("sg-id")
	return cmd
}

