package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	apierrors "github.com/emma-community/emma-cli/internal/apierrors"
	"github.com/emma-community/emma-cli/internal/cmdutil"
	"github.com/emma-community/emma-cli/internal/output"
	emma "github.com/emma-community/emma-go-sdk"
	"github.com/spf13/cobra"
)

// k8sRowable abstracts over the multiple K8s response types in SDK v0.0.12.
// Each SDK response type (list, get, create, update, delete) has generated
// getter methods that satisfy this interface.
type k8sRowable interface {
	GetId() int32
	GetName() string
	GetStatus() string
	GetVersion() string
	GetDeploymentLocation() string
	GetK8sConnectionType() string
}

func k8sToRow(k k8sRowable) []string {
	id := strconv.Itoa(int(k.GetId()))
	return []string{
		k.GetName(),
		id,
		k.GetStatus(),
		k.GetVersion(),
		k.GetDeploymentLocation(),
		k.GetK8sConnectionType(),
	}
}

func (c *CLI) newK8sCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "k8s",
		Aliases: []string{"kubernetes"},
		Short:   "Manage Kubernetes clusters",
	}
	cmd.AddCommand(c.newK8sListCmd())
	cmd.AddCommand(c.newK8sGetCmd())
	cmd.AddCommand(c.newK8sCreateCmd())
	cmd.AddCommand(c.newK8sEditCmd())
	cmd.AddCommand(c.newK8sDeleteCmd())
	return cmd
}

func (c *CLI) newK8sListCmd() *cobra.Command {
	var statusFilter string

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
				if statusFilter != "" && !strings.EqualFold(derefStr(k.Status), statusFilter) {
					continue
				}
				rows = append(rows, k8sToRow(&k))
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, clusters, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "VERSION", "LOCATION", "CONNECTION-TYPE"},
				Rows:    rows,
			})
		},
	}

	cmd.Flags().StringVar(&statusFilter, "status", "", "Filter by status")
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
				Rows:    [][]string{k8sToRow(cluster)},
			})
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "Cluster ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (c *CLI) newK8sCreateCmd() *cobra.Command {
	var name, deploymentLocation, connectionType, version string
	var workerName, workerDatacenterID, workerVCpuType, workerVolumeType string
	var workerVCpu, workerRam, workerVolumeSize int32
	var workersFile, autoscalingFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Kubernetes cluster",
		Long: `Create a Kubernetes cluster.

For a single worker group, use --worker-* flags.
For multiple worker groups, use --workers-file with a JSON file:
  [{"name":"group1","dataCenterId":"dc-1","vCpuType":"shared","vCpu":2,"ramGb":4,"volumeGb":16,"volumeType":"ssd"}]

For autoscaling, use --autoscaling-file with a JSON file:
  [{"groupName":"group1","dataCenterId":"dc-1","minimumNodes":1,"maximumNodes":5,"targetNodes":2}]`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			var workerNodes []emma.KubernetesCreateRequestWorkerNodesInner
			if workersFile != "" {
				data, err := os.ReadFile(workersFile)
				if err != nil {
					return fmt.Errorf("reading workers file: %w", err)
				}
				if err := json.Unmarshal(data, &workerNodes); err != nil {
					return fmt.Errorf("invalid workers JSON: %w", err)
				}
			} else if workerName != "" {
				workerNodes = []emma.KubernetesCreateRequestWorkerNodesInner{{
					Name:         workerName,
					DataCenterId: workerDatacenterID,
					VCpuType:     workerVCpuType,
					VCpu:         workerVCpu,
					RamGb:        workerRam,
					VolumeGb:     workerVolumeSize,
					VolumeType:   workerVolumeType,
				}}
			}

			createReq := emma.KubernetesCreateRequest{
				Name:               name,
				DeploymentLocation: normalizeDeploymentLocation(deploymentLocation),
				K8sConnectionType:  normalizeConnectionType(connectionType),
				WorkerNodes:        workerNodes,
			}
			if version != "" {
				createReq.K8sVersion = &version
			}

			if autoscalingFile != "" {
				data, err := os.ReadFile(autoscalingFile)
				if err != nil {
					return fmt.Errorf("reading autoscaling file: %w", err)
				}
				var configs []emma.KubernetesCreateRequestAutoscalingConfigsInner
				if err := json.Unmarshal(data, &configs); err != nil {
					return fmt.Errorf("invalid autoscaling JSON: %w", err)
				}
				createReq.AutoscalingConfigs = configs
			}

			cluster, _, err := c.Client.KubernetesClustersAPI.CreateKubernetesCluster(ctx).KubernetesCreateRequest(createReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, cluster, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "VERSION", "LOCATION", "CONNECTION-TYPE"},
				Rows:    [][]string{k8sToRow(cluster)},
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Cluster name (required)")
	cmd.Flags().StringVar(&deploymentLocation, "deployment-location", "", "Deployment location: eu, us, apac (required)")
	cmd.Flags().StringVar(&connectionType, "connection-type", "", "K8s connection type: internet_connect or direct_connect (required)")
	cmd.Flags().StringVar(&version, "version", "", "Kubernetes version")

	cmd.Flags().StringVar(&workerName, "worker-name", "", "Worker node group name")
	cmd.Flags().StringVar(&workerDatacenterID, "worker-datacenter-id", "", "Worker node datacenter ID")
	cmd.Flags().StringVar(&workerVCpuType, "worker-vcpu-type", "shared", "Worker node vCPU type")
	cmd.Flags().Int32Var(&workerVCpu, "worker-vcpu", 0, "Worker node vCPUs")
	cmd.Flags().Int32Var(&workerRam, "worker-ram", 0, "Worker node RAM in GB")
	cmd.Flags().Int32Var(&workerVolumeSize, "worker-volume-size", 0, "Worker node volume size in GB")
	cmd.Flags().StringVar(&workerVolumeType, "worker-volume-type", "", "Worker node volume type")

	cmd.Flags().StringVar(&workersFile, "workers-file", "", "JSON file with multiple worker node groups")
	cmd.Flags().StringVar(&autoscalingFile, "autoscaling-file", "", "JSON file with autoscaling configs")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("deployment-location")
	_ = cmd.MarkFlagRequired("connection-type")

	return cmd
}

func (c *CLI) newK8sEditCmd() *cobra.Command {
	var id int32
	var workerName, workerDatacenterID, workerVCpuType, workerVolumeType string
	var workerVCpu, workerRam, workerVolumeSize int32
	var workersFile, autoscalingFile string

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			var workerNodes []emma.KubernetesUpdateRequestWorkerNodesInner
			if workersFile != "" {
				data, err := os.ReadFile(workersFile)
				if err != nil {
					return fmt.Errorf("reading workers file: %w", err)
				}
				if err := json.Unmarshal(data, &workerNodes); err != nil {
					return fmt.Errorf("invalid workers JSON: %w", err)
				}
			} else if workerName != "" {
				workerNodes = []emma.KubernetesUpdateRequestWorkerNodesInner{{
					Name:         workerName,
					DataCenterId: workerDatacenterID,
					VCpuType:     workerVCpuType,
					VCpu:         workerVCpu,
					RamGb:        workerRam,
					VolumeGb:     workerVolumeSize,
					VolumeType:   workerVolumeType,
				}}
			}

			updateReq := emma.KubernetesUpdateRequest{
				WorkerNodes: workerNodes,
			}

			if autoscalingFile != "" {
				data, err := os.ReadFile(autoscalingFile)
				if err != nil {
					return fmt.Errorf("reading autoscaling file: %w", err)
				}
				var configs []emma.KubernetesUpdateRequestAutoscalingConfigsInner
				if err := json.Unmarshal(data, &configs); err != nil {
					return fmt.Errorf("invalid autoscaling JSON: %w", err)
				}
				updateReq.AutoscalingConfigs = configs
			}

			cluster, _, err := c.Client.KubernetesClustersAPI.EditKubernetesCluster(ctx, id).KubernetesUpdateRequest(updateReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, cluster, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "VERSION", "LOCATION", "CONNECTION-TYPE"},
				Rows:    [][]string{k8sToRow(cluster)},
			})
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "Cluster ID (required)")
	cmd.Flags().StringVar(&workerName, "worker-name", "", "Worker node group name")
	cmd.Flags().StringVar(&workerDatacenterID, "worker-datacenter-id", "", "Worker node datacenter ID")
	cmd.Flags().StringVar(&workerVCpuType, "worker-vcpu-type", "shared", "Worker node vCPU type")
	cmd.Flags().Int32Var(&workerVCpu, "worker-vcpu", 0, "Worker node vCPUs")
	cmd.Flags().Int32Var(&workerRam, "worker-ram", 0, "Worker node RAM in GB")
	cmd.Flags().Int32Var(&workerVolumeSize, "worker-volume-size", 0, "Worker node volume size in GB")
	cmd.Flags().StringVar(&workerVolumeType, "worker-volume-type", "", "Worker node volume type")
	cmd.Flags().StringVar(&workersFile, "workers-file", "", "JSON file with multiple worker node groups")
	cmd.Flags().StringVar(&autoscalingFile, "autoscaling-file", "", "JSON file with autoscaling configs")

	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func normalizeDeploymentLocation(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func normalizeConnectionType(s string) string {
	v := strings.ToLower(strings.TrimSpace(s))
	switch v {
	case "directconnect":
		return "direct_connect"
	case "internetconnect":
		return "internet_connect"
	default:
		return v
	}
}

func (c *CLI) newK8sDeleteCmd() *cobra.Command {
	var id int32
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Kubernetes cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmdutil.ConfirmDelete(c.Err, cmd.InOrStdin(), "Kubernetes cluster", id, yes) {
				fmt.Fprintln(c.Out, "Aborted.")
				return nil
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
