package cmd

import (
	"context"
	"fmt"
	"strconv"

	apierrors "github.com/emma-community/emma-cli/internal/apierrors"
	"github.com/emma-community/emma-cli/internal/output"
	"github.com/spf13/cobra"
)

// addConfigResourceCmds adds resource configuration listing subcommands to config.
func (c *CLI) addConfigResourceCmds(parent *cobra.Command) {
	parent.AddCommand(c.newVMConfigsCmd())
	parent.AddCommand(c.newSpotConfigsCmd())
	parent.AddCommand(c.newK8sConfigsCmd())
	parent.AddCommand(c.newVolumeConfigsCmd())
}

func (c *CLI) newVMConfigsCmd() *cobra.Command {
	var datacenterID, vcpuType, volumeType string
	var providerId, locationId, vcpu, ram int32

	cmd := &cobra.Command{
		Use:   "vm-configs",
		Short: "List available VM hardware configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			req := c.Client.ComputeInstancesConfigurationsAPI.GetVmConfigs(ctx)
			if datacenterID != "" {
				req = req.DataCenterId(datacenterID)
			}
			if vcpuType != "" {
				req = req.VCpuType(vcpuType)
			}
			if volumeType != "" {
				req = req.VolumeType(volumeType)
			}
			if providerId > 0 {
				req = req.ProviderId(providerId)
			}
			if locationId > 0 {
				req = req.LocationId(locationId)
			}
			if vcpu > 0 {
				req = req.VCpu(vcpu)
			}
			if ram > 0 {
				req = req.RamGb(ram)
			}
			if c.ProjectID != nil {
				req = req.ProjectId(*c.ProjectID)
			}

			// Fetch all pages
			var allRows [][]string
			var allData []interface{}
			page := int32(0)
			for {
				resp, _, err := req.Page(page).Size(100).Execute()
				if err != nil {
					return apierrors.Format(err)
				}
				for _, cfg := range resp.Content {
					cost := ""
					if cfg.Cost != nil && cfg.Cost.PricePerUnit != nil {
						cost = fmt.Sprintf("%.4f %s/%s", *cfg.Cost.PricePerUnit, derefStr(cfg.Cost.Currency), derefStr(cfg.Cost.Unit))
					}
					allRows = append(allRows, []string{
						derefStr(cfg.ProviderName),
						derefStr(cfg.DataCenterName),
						derefStr(cfg.VCpuType),
						i32str(cfg.VCpu),
						i32str(cfg.RamGb),
						i32str(cfg.VolumeGb),
						derefStr(cfg.VolumeType),
						cost,
					})
				}
				allData = append(allData, resp.Content)
				if resp.Last != nil && *resp.Last {
					break
				}
				page++
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, allData, output.TableView{
				Headers: []string{"PROVIDER", "DATACENTER", "VCPU-TYPE", "VCPU", "RAM(GB)", "VOL(GB)", "VOL-TYPE", "COST"},
				Rows:    allRows,
			})
		},
	}

	cmd.Flags().StringVar(&datacenterID, "datacenter-id", "", "Filter by datacenter ID")
	cmd.Flags().StringVar(&vcpuType, "vcpu-type", "", "Filter by vCPU type")
	cmd.Flags().StringVar(&volumeType, "volume-type", "", "Filter by volume type")
	cmd.Flags().Int32Var(&providerId, "provider-id", 0, "Filter by provider ID")
	cmd.Flags().Int32Var(&locationId, "location-id", 0, "Filter by location ID")
	cmd.Flags().Int32Var(&vcpu, "vcpu", 0, "Filter by vCPU count")
	cmd.Flags().Int32Var(&ram, "ram", 0, "Filter by RAM in GB")
	return cmd
}

func (c *CLI) newSpotConfigsCmd() *cobra.Command {
	var datacenterID, vcpuType, volumeType string

	cmd := &cobra.Command{
		Use:   "spot-configs",
		Short: "List available spot instance configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			req := c.Client.ComputeInstancesConfigurationsAPI.GetSpotConfigs(ctx)
			if datacenterID != "" {
				req = req.DataCenterId(datacenterID)
			}
			if vcpuType != "" {
				req = req.VCpuType(vcpuType)
			}
			if volumeType != "" {
				req = req.VolumeType(volumeType)
			}
			if c.ProjectID != nil {
				req = req.ProjectId(*c.ProjectID)
			}

			var allRows [][]string
			var allData []interface{}
			page := int32(0)
			for {
				resp, _, err := req.Page(page).Size(100).Execute()
				if err != nil {
					return apierrors.Format(err)
				}
				for _, cfg := range resp.Content {
					cost := ""
					if cfg.Cost != nil && cfg.Cost.PricePerUnit != nil {
						cost = fmt.Sprintf("%.4f %s/%s", *cfg.Cost.PricePerUnit, derefStr(cfg.Cost.Currency), derefStr(cfg.Cost.Unit))
					}
					allRows = append(allRows, []string{
						derefStr(cfg.ProviderName),
						derefStr(cfg.DataCenterName),
						derefStr(cfg.VCpuType),
						i32str(cfg.VCpu),
						i32str(cfg.RamGb),
						i32str(cfg.VolumeGb),
						derefStr(cfg.VolumeType),
						cost,
					})
				}
				allData = append(allData, resp.Content)
				if resp.Last != nil && *resp.Last {
					break
				}
				page++
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, allData, output.TableView{
				Headers: []string{"PROVIDER", "DATACENTER", "VCPU-TYPE", "VCPU", "RAM(GB)", "VOL(GB)", "VOL-TYPE", "COST"},
				Rows:    allRows,
			})
		},
	}

	cmd.Flags().StringVar(&datacenterID, "datacenter-id", "", "Filter by datacenter ID")
	cmd.Flags().StringVar(&vcpuType, "vcpu-type", "", "Filter by vCPU type")
	cmd.Flags().StringVar(&volumeType, "volume-type", "", "Filter by volume type")
	return cmd
}

func (c *CLI) newK8sConfigsCmd() *cobra.Command {
	var connectionType, datacenterID, vcpuType, volumeType string

	cmd := &cobra.Command{
		Use:   "k8s-configs",
		Short: "List available Kubernetes node configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			req := c.Client.ComputeInstancesConfigurationsAPI.GetKuberNodesConfigs(ctx, normalizeConnectionType(connectionType))
			if datacenterID != "" {
				req = req.DataCenterId(datacenterID)
			}
			if vcpuType != "" {
				req = req.VCpuType(vcpuType)
			}
			if volumeType != "" {
				req = req.VolumeType(volumeType)
			}
			if c.ProjectID != nil {
				req = req.ProjectId(*c.ProjectID)
			}

			var allRows [][]string
			var allData []interface{}
			page := int32(0)
			for {
				resp, _, err := req.Page(page).Size(100).Execute()
				if err != nil {
					return apierrors.Format(err)
				}
				for _, cfg := range resp.Content {
					cost := ""
					if cfg.Cost != nil && cfg.Cost.PricePerUnit != nil {
						cost = fmt.Sprintf("%.4f %s/%s", *cfg.Cost.PricePerUnit, derefStr(cfg.Cost.Currency), derefStr(cfg.Cost.Unit))
					}
					allRows = append(allRows, []string{
						derefStr(cfg.ProviderName),
						derefStr(cfg.DataCenterName),
						derefStr(cfg.VCpuType),
						i32str(cfg.VCpu),
						i32str(cfg.RamGb),
						i32str(cfg.VolumeGb),
						derefStr(cfg.VolumeType),
						cost,
					})
				}
				allData = append(allData, resp.Content)
				if resp.Last != nil && *resp.Last {
					break
				}
				page++
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, allData, output.TableView{
				Headers: []string{"PROVIDER", "DATACENTER", "VCPU-TYPE", "VCPU", "RAM(GB)", "VOL(GB)", "VOL-TYPE", "COST"},
				Rows:    allRows,
			})
		},
	}

	cmd.Flags().StringVar(&connectionType, "connection-type", "", "K8s connection type: internet_connect or direct_connect (required)")
	cmd.Flags().StringVar(&datacenterID, "datacenter-id", "", "Filter by datacenter ID")
	cmd.Flags().StringVar(&vcpuType, "vcpu-type", "", "Filter by vCPU type")
	cmd.Flags().StringVar(&volumeType, "volume-type", "", "Filter by volume type")
	_ = cmd.MarkFlagRequired("connection-type")
	return cmd
}

func (c *CLI) newVolumeConfigsCmd() *cobra.Command {
	var datacenterID, volumeType string

	cmd := &cobra.Command{
		Use:   "volume-configs",
		Short: "List available volume configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			req := c.Client.VolumesConfigurationsAPI.GetVolumeConfigs(ctx).IsBootable(false)
			if datacenterID != "" {
				req = req.DataCenterId(datacenterID)
			}
			if volumeType != "" {
				req = req.VolumeType(volumeType)
			}
			if c.ProjectID != nil {
				req = req.ProjectId(*c.ProjectID)
			}

			var allRows [][]string
			var allData []interface{}
			page := int32(0)
			for {
				resp, _, err := req.Page(page).Size(100).Execute()
				if err != nil {
					return apierrors.Format(err)
				}
				for _, cfg := range resp.Content {
					cost := ""
					if cfg.Cost != nil && cfg.Cost.PricePerUnit != nil {
						cost = fmt.Sprintf("%.4f %s/%s", *cfg.Cost.PricePerUnit, derefStr(cfg.Cost.Currency), derefStr(cfg.Cost.Unit))
					}
					allRows = append(allRows, []string{
						derefStr(cfg.ProviderName),
						derefStr(cfg.DataCenterName),
						i32str(cfg.VolumeGb),
						derefStr(cfg.VolumeType),
						cost,
					})
				}
				allData = append(allData, resp.Content)
				if resp.Last != nil && *resp.Last {
					break
				}
				page++
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, allData, output.TableView{
				Headers: []string{"PROVIDER", "DATACENTER", "VOL(GB)", "VOL-TYPE", "COST"},
				Rows:    allRows,
			})
		},
	}

	cmd.Flags().StringVar(&datacenterID, "datacenter-id", "", "Filter by datacenter ID")
	cmd.Flags().StringVar(&volumeType, "volume-type", "", "Filter by volume type")
	return cmd
}

// i32str converts an *int32 to string, returning "" for nil.
func i32str(v *int32) string {
	if v == nil {
		return ""
	}
	return strconv.Itoa(int(*v))
}
