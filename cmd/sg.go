package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"

	emma "github.com/emma-community/emma-go-sdk"
	apierrors "github.com/emma-community/emma-cli/internal/apierrors"
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
	cmd.AddCommand(c.newSGDeleteCmd())
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
		Short: "Get a security group by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			sg, _, err := c.Client.SecurityGroupsAPI.GetSecurityGroup(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, sg, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "RULES"},
				Rows:    [][]string{sgToRow(*sg)},
			})
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

func (c *CLI) newSGDeleteCmd() *cobra.Command {
	var id int32
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a security group",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				fmt.Fprintf(c.Err, "Delete security group %d? [y/N] ", id)
				reader := bufio.NewReader(strings.NewReader(""))
				_ = reader
				var input string
				fmt.Fscan(cmd.InOrStdin(), &input)
				if strings.ToLower(strings.TrimSpace(input)) != "y" {
					fmt.Fprintln(c.Out, "Aborted.")
					return nil
				}
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
