package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	apierrors "github.com/emma-community/emma-cli/internal/apierrors"
	"github.com/emma-community/emma-cli/internal/cmdutil"
	"github.com/emma-community/emma-cli/internal/output"
	emma "github.com/emma-community/emma-go-sdk"
	"github.com/spf13/cobra"
)

func (c *CLI) newWorkflowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Manage workflow templates",
	}
	cmd.AddCommand(c.newWorkflowListCmd())
	cmd.AddCommand(c.newWorkflowCreateCmd())
	cmd.AddCommand(c.newWorkflowDeleteCmd())
	return cmd
}

func workflowToRow(w emma.WorkflowTemplate) []string {
	return []string{
		w.Name,
		strconv.Itoa(int(w.Id)),
		w.Status,
		w.ResourceType,
		w.ContentType,
		w.CreatedByName,
	}
}

func (c *CLI) newWorkflowListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workflow templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Fetch all pages
			var allWorkflows []emma.WorkflowTemplate
			page := int32(0)
			for {
				resp, _, err := c.Client.WorkflowsAPI.GetWorkflowTemplates(ctx).Page(page).Size(100).Execute()
				if err != nil {
					return apierrors.Format(err)
				}
				allWorkflows = append(allWorkflows, resp.Content...)
				if resp.Last != nil && *resp.Last {
					break
				}
				if resp.TotalPages != nil && page+1 >= *resp.TotalPages {
					break
				}
				page++
			}

			rows := make([][]string, 0, len(allWorkflows))
			for _, w := range allWorkflows {
				rows = append(rows, workflowToRow(w))
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, allWorkflows, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "RESOURCE-TYPE", "CONTENT-TYPE", "CREATED-BY"},
				Rows:    rows,
			})
		},
	}
	return cmd
}

func (c *CLI) newWorkflowCreateCmd() *cobra.Command {
	var name, description, contentType, content, status, resourceType string
	var contentParamsJSON, resourceParamsJSON, tagsJSON string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a workflow template",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			createReq := emma.WorkflowTemplateCreate{
				Name:         name,
				ContentType:  contentType,
				Content:      content,
				Status:       status,
				ResourceType: resourceType,
			}
			if description != "" {
				createReq.Description = &description
			}

			if contentParamsJSON != "" {
				var params []emma.WorkflowTemplateCreateContentParamsInner
				if err := json.Unmarshal([]byte(contentParamsJSON), &params); err != nil {
					return fmt.Errorf("invalid --content-params JSON: %w", err)
				}
				createReq.ContentParams = params
			}
			if resourceParamsJSON != "" {
				var params []emma.WorkflowTemplateCreateResourceParamsInner
				if err := json.Unmarshal([]byte(resourceParamsJSON), &params); err != nil {
					return fmt.Errorf("invalid --resource-params JSON: %w", err)
				}
				createReq.ResourceParams = params
			}
			if tagsJSON != "" {
				var tags []emma.WorkflowTemplateCreateTagsInner
				if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
					return fmt.Errorf("invalid --tags JSON: %w", err)
				}
				createReq.Tags = tags
			}

			wf, _, err := c.Client.WorkflowsAPI.CreateWorkflowTemplate(ctx).WorkflowTemplateCreate(createReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, wf, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "RESOURCE-TYPE", "CONTENT-TYPE", "CREATED-BY"},
				Rows:    [][]string{workflowToRow(*wf)},
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Template name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Template description")
	cmd.Flags().StringVar(&contentType, "content-type", "", "Content type (required)")
	cmd.Flags().StringVar(&content, "content", "", "Template content (required)")
	cmd.Flags().StringVar(&status, "status", "", "Status (required)")
	cmd.Flags().StringVar(&resourceType, "resource-type", "", "Resource type (required)")
	cmd.Flags().StringVar(&contentParamsJSON, "content-params", "", "Content params as JSON array")
	cmd.Flags().StringVar(&resourceParamsJSON, "resource-params", "", "Resource params as JSON array")
	cmd.Flags().StringVar(&tagsJSON, "tags", "", "Tags as JSON array")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("content-type")
	_ = cmd.MarkFlagRequired("content")
	_ = cmd.MarkFlagRequired("status")
	_ = cmd.MarkFlagRequired("resource-type")
	return cmd
}

func (c *CLI) newWorkflowDeleteCmd() *cobra.Command {
	var id int32
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a workflow template",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmdutil.ConfirmDelete(c.Err, cmd.InOrStdin(), "workflow template", id, yes) {
				fmt.Fprintln(c.Out, "Aborted.")
				return nil
			}

			ctx := context.Background()
			_, err := c.Client.WorkflowsAPI.DeleteWorkflowTemplate(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			fmt.Fprintf(c.Out, "Deleted workflow template %d.\n", id)
			return nil
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "Workflow template ID")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}
