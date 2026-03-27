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

func (c *CLI) newSpotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spot",
		Short: "Manage spot instances",
	}
	cmd.AddCommand(c.newSpotListCmd())
	cmd.AddCommand(c.newSpotGetCmd())
	cmd.AddCommand(c.newSpotCreateCmd())
	cmd.AddCommand(c.newSpotDeleteCmd())
	return cmd
}

func spotToRow(s emma.SpotVm) []string {
	name := derefStr(s.Name)
	id := ""
	if s.Id != nil {
		id = strconv.Itoa(int(*s.Id))
	}
	status := derefStr(s.Status)
	osName := ""
	if s.Os != nil && s.Os.Type != nil {
		osName = *s.Os.Type
	}
	vcpu := ""
	if s.VCpu != nil {
		vcpu = strconv.Itoa(int(*s.VCpu))
	}
	ram := ""
	if s.RamGb != nil {
		ram = strconv.Itoa(int(*s.RamGb))
	}
	dc := ""
	if s.DataCenter != nil && s.DataCenter.Name != nil {
		dc = *s.DataCenter.Name
	}
	ip := ""
	for _, net := range s.Networks {
		if net.Ip != nil && *net.Ip != "" {
			ip = *net.Ip
			break
		}
	}
	return []string{name, id, status, osName, vcpu, ram, dc, ip}
}

func (c *CLI) newSpotListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List spot instances",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			req := c.Client.SpotInstancesAPI.GetSpots(ctx)
			if c.ProjectID != nil {
				req = req.ProjectId(*c.ProjectID)
			}
			spots, _, err := req.Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			rows := make([][]string, 0, len(spots))
			for _, s := range spots {
				rows = append(rows, spotToRow(s))
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, spots, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "OS", "VCPU", "RAM(GB)", "DATACENTER", "IP"},
				Rows:    rows,
			})
		},
	}
	return cmd
}

func (c *CLI) newSpotGetCmd() *cobra.Command {
	var id int32

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a spot instance by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			spot, _, err := c.Client.SpotInstancesAPI.GetSpot(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, spot, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "OS", "VCPU", "RAM(GB)", "DATACENTER", "IP"},
				Rows:    [][]string{spotToRow(*spot)},
			})
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "Spot instance ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (c *CLI) newSpotCreateCmd() *cobra.Command {
	var name, datacenterID, cloudNetworkType, volumeType, vcpuType string
	var osID, vcpu, ram, volumeSize int32
	var price float32

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a spot instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			createReq := emma.SpotCreate{
				Name:             name,
				DataCenterId:     datacenterID,
				OsId:             osID,
				CloudNetworkType: cloudNetworkType,
				VCpuType:         vcpuType,
				VCpu:             vcpu,
				RamGb:            ram,
				VolumeType:       volumeType,
				VolumeGb:         volumeSize,
				Price:            price,
			}

			spot, _, err := c.Client.SpotInstancesAPI.SpotCreate(ctx).SpotCreate(createReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, spot, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "OS", "VCPU", "RAM(GB)", "DATACENTER", "IP"},
				Rows:    [][]string{spotToRow(*spot)},
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Instance name (required)")
	cmd.Flags().StringVar(&datacenterID, "datacenter-id", "", "Datacenter ID (required)")
	cmd.Flags().Int32Var(&osID, "os-id", 0, "OS ID (required)")
	cmd.Flags().Int32Var(&vcpu, "vcpu", 0, "Number of vCPUs (required)")
	cmd.Flags().Int32Var(&ram, "ram", 0, "RAM in GB (required)")
	cmd.Flags().Int32Var(&volumeSize, "volume-size", 0, "Volume size in GB (required)")
	cmd.Flags().StringVar(&volumeType, "volume-type", "", "Volume type (required)")
	cmd.Flags().StringVar(&cloudNetworkType, "cloud-network-type", "", "Cloud network type (required)")
	cmd.Flags().StringVar(&vcpuType, "vcpu-type", "shared", "vCPU type")
	cmd.Flags().Float32Var(&price, "price", 0, "Max price per hour")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("datacenter-id")
	_ = cmd.MarkFlagRequired("os-id")
	_ = cmd.MarkFlagRequired("vcpu")
	_ = cmd.MarkFlagRequired("ram")
	_ = cmd.MarkFlagRequired("volume-size")
	_ = cmd.MarkFlagRequired("volume-type")
	_ = cmd.MarkFlagRequired("cloud-network-type")

	return cmd
}

func (c *CLI) newSpotDeleteCmd() *cobra.Command {
	var id int32
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a spot instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				fmt.Fprintf(c.Err, "Delete spot instance %d? [y/N] ", id)
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
			_, _, err := c.Client.SpotInstancesAPI.SpotDelete(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			fmt.Fprintf(c.Out, "Deleted spot instance %d.\n", id)
			return nil
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "Spot instance ID")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}
