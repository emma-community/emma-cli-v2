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

func (c *CLI) newK8sCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "k8s",
		Aliases: []string{"kubernetes"},
		Short:   "Manage Kubernetes clusters",
	}
	cmd.AddCommand(c.newK8sListCmd())
	cmd.AddCommand(c.newK8sGetCmd())
	cmd.AddCommand(c.newK8sCreateCmd())
	cmd.AddCommand(c.newK8sDeleteCmd())
	return cmd
}

func k8sToRow(k emma.Kubernetes) []string {
	name := derefStr(k.Name)
	id := ""
	if k.Id != nil {
		id = strconv.Itoa(int(*k.Id))
	}
	status := derefStr(k.Status)
	version := derefStr(k.Version)
	location := derefStr(k.DeploymentLocation)
	connType := derefStr(k.K8sConnectionType)
	return []string{name, id, status, version, location, connType}
}

func (c *CLI) newK8sListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kubernetes clusters",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			req := c.Client.KubernetesClustersAPI.GetKubernetesClusters(ctx)
			if c.ProjectID != nil {
				req = req.ProjectId(*c.ProjectID)
			}
			clusters, _, err := req.Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			rows := make([][]string, 0, len(clusters))
			for _, k := range clusters {
				rows = append(rows, k8sToRow(k))
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, clusters, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "VERSION", "LOCATION", "CONNECTION-TYPE"},
				Rows:    rows,
			})
		},
	}
	return cmd
}

func (c *CLI) newK8sGetCmd() *cobra.Command {
	var id int32

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a Kubernetes cluster by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			cluster, _, err := c.Client.KubernetesClustersAPI.GetKubernetesCluster(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, cluster, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "VERSION", "LOCATION", "CONNECTION-TYPE"},
				Rows:    [][]string{k8sToRow(*cluster)},
			})
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "Cluster ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (c *CLI) newK8sCreateCmd() *cobra.Command {
	var name, deploymentLocation, connectionType string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			createReq := emma.KubernetesCreate{
				Name:               name,
				DeploymentLocation: deploymentLocation,
				K8sConnectionType:  connectionType,
				WorkerNodes:        []emma.KubernetesCreateWorkerNodesInner{},
			}

			cluster, _, err := c.Client.KubernetesClustersAPI.CreateKubernetesCluster(ctx).KubernetesCreate(createReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, cluster, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "VERSION", "LOCATION", "CONNECTION-TYPE"},
				Rows:    [][]string{k8sToRow(*cluster)},
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Cluster name (required)")
	cmd.Flags().StringVar(&deploymentLocation, "deployment-location", "", "Deployment location (required)")
	cmd.Flags().StringVar(&connectionType, "connection-type", "", "K8s connection type (required)")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("deployment-location")
	_ = cmd.MarkFlagRequired("connection-type")

	return cmd
}

func (c *CLI) newK8sDeleteCmd() *cobra.Command {
	var id int32
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				fmt.Fprintf(c.Err, "Delete Kubernetes cluster %d? [y/N] ", id)
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
			_, _, err := c.Client.KubernetesClustersAPI.DeleteKubernetesCluster(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			fmt.Fprintf(c.Out, "Deleted Kubernetes cluster %d.\n", id)
			return nil
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "Cluster ID")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}
