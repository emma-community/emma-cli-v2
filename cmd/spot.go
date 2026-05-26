package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	apierrors "github.com/emma-community/emma-cli/internal/apierrors"
	"github.com/emma-community/emma-cli/internal/cmdutil"
	"github.com/emma-community/emma-cli/internal/output"
	"github.com/emma-community/emma-cli/internal/poller"
	emma "github.com/emma-community/emma-go-sdk"
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
	cmd.AddCommand(c.newSpotActionsCmd())
	cmd.AddCommand(c.newSpotWaitCmd())
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

func spotToWideRow(s emma.SpotVm) []string {
	row := spotToRow(s)
	provider := ""
	if s.Provider != nil && s.Provider.Name != nil {
		provider = *s.Provider.Name
	}
	cost := ""
	if s.Cost != nil && s.Cost.Price != nil {
		currency := derefStr(s.Cost.Currency)
		cost = fmt.Sprintf("%.2f %s/%s", *s.Cost.Price, currency, derefStr(s.Cost.Unit))
	}
	return append(row, provider, cost)
}

func (c *CLI) newSpotListCmd() *cobra.Command {
	var statusFilter string

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
			wideRows := make([][]string, 0, len(spots))
			for _, s := range spots {
				if statusFilter != "" && !strings.EqualFold(derefStr(s.Status), statusFilter) {
					continue
				}
				rows = append(rows, spotToRow(s))
				wideRows = append(wideRows, spotToWideRow(s))
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, spots, output.TableView{
				Headers:     []string{"NAME", "ID", "STATUS", "OS", "VCPU", "RAM(GB)", "DATACENTER", "IP"},
				Rows:        rows,
				WideHeaders: []string{"NAME", "ID", "STATUS", "OS", "VCPU", "RAM(GB)", "DATACENTER", "IP", "PROVIDER", "COST"},
				WideRows:    wideRows,
			})
		},
	}

	cmd.Flags().StringVar(&statusFilter, "status", "", "Filter by status")
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
				Headers:     []string{"NAME", "ID", "STATUS", "OS", "VCPU", "RAM(GB)", "DATACENTER", "IP"},
				Rows:        [][]string{spotToRow(*spot)},
				WideHeaders: []string{"NAME", "ID", "STATUS", "OS", "VCPU", "RAM(GB)", "DATACENTER", "IP", "PROVIDER", "COST"},
				WideRows:    [][]string{spotToWideRow(*spot)},
			})
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "Spot instance ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (c *CLI) newSpotCreateCmd() *cobra.Command {
	var name, datacenterID, cloudNetworkType, volumeType, vcpuType, acceleratorTypeID string
	var osID, vcpu, ram, volumeSize, sshKeyID, securityGroupID int32
	var price, accelerators float32

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a spot instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			configReq := c.Client.ComputeInstancesConfigurationsAPI.GetSpotConfigs(ctx)
			configReq = configReq.DataCenterId(datacenterID)
			configReq = configReq.VCpu(vcpu)
			configReq = configReq.RamGb(ram)
			configReq = configReq.VolumeType(volumeType)
			configReq = configReq.VolumeGb(volumeSize)
			if c.ProjectID != nil {
				configReq = configReq.ProjectId(*c.ProjectID)
			}
			configResp, _, configErr := configReq.Execute()
			if configErr == nil && configResp != nil && len(configResp.Content) == 1 {
				cfg := configResp.Content[0]
				if cfg.Cost != nil && cfg.Cost.PricePerUnit != nil {
					currency := derefStr(cfg.Cost.Currency)
					unit := derefStr(cfg.Cost.Unit)
					fmt.Fprintf(c.Err, "Estimated cost: %.2f %s/%s\n", *cfg.Cost.PricePerUnit, currency, unit)
				}
			}

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
			if cmd.Flags().Changed("ssh-key-id") {
				createReq.SshKeyId = &sshKeyID
			}
			if cmd.Flags().Changed("security-group-id") {
				createReq.SecurityGroupId = &securityGroupID
			}
			if cmd.Flags().Changed("accelerator-type-id") {
				createReq.AcceleratorTypeId = &acceleratorTypeID
			}
			if cmd.Flags().Changed("accelerators") {
				createReq.Accelerators = &accelerators
			}

			if cmd.Flags().Changed("ssh-key-id") {
				createReq.SshKeyId = &sshKeyID
			}

			spot, _, err := c.Client.SpotInstancesAPI.SpotCreate(ctx).SpotCreate(createReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			fmt.Fprintf(c.Err, "Spot %q created (ID: %d). Waiting for IP assignment...\n", derefStr(spot.Name), derefInt32(spot.Id))

			spotID := *spot.Id
			pollErr := poller.WaitFor(ctx, 5*time.Second, 60*time.Second, func() (bool, error) {
				updated, _, err := c.Client.SpotInstancesAPI.GetSpot(ctx, spotID).Execute()
				if err != nil {
					if apierrors.IsNotFound(err) || apierrors.IsUnauthorized(err) {
						return false, apierrors.Format(err)
					}
					return false, nil
				}
				for _, net := range updated.Networks {
					if net.Ip != nil && *net.Ip != "" {
						spot = updated
						return true, nil
					}
				}
				fmt.Fprint(c.Err, ".")
				return false, nil
			})
			fmt.Fprintln(c.Err)
			if pollErr != nil {
				fmt.Fprintf(c.Err, "Warning: timed out waiting for IP: %v\n", pollErr)
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
	cmd.Flags().Int32Var(&sshKeyID, "ssh-key-id", 0, "SSH key ID to inject (see `emma sshkey list`)")
	cmd.Flags().Int32Var(&securityGroupID, "security-group-id", 0, "Security group ID (see `emma sg list`)")
	cmd.Flags().StringVar(&acceleratorTypeID, "accelerator-type-id", "", "GPU accelerator type ID")
	cmd.Flags().Float32Var(&accelerators, "accelerators", 0, "Number of GPU accelerators")

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
			if !cmdutil.ConfirmDelete(c.Err, cmd.InOrStdin(), "spot instance", id, yes) {
				fmt.Fprintln(c.Out, "Aborted.")
				return nil
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

func (c *CLI) newSpotActionsCmd() *cobra.Command {
	var id int32
	var action, newName string
	var price float32

	cmd := &cobra.Command{
		Use:   "actions",
		Short: "Perform an action on a spot instance",
		Long:  "Supported actions: start, shutdown, reboot, rename, change-price",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			var req emma.SpotActionsRequest
			switch strings.ToLower(action) {
			case "start":
				req.SpotStart = emma.NewSpotStart("start")
			case "shutdown":
				req.SpotShutdown = emma.NewSpotShutdown("shutdown")
			case "reboot":
				req.SpotReboot = emma.NewSpotReboot("reboot")
			case "rename":
				if newName == "" {
					return fmt.Errorf("--name is required for rename action")
				}
				req.SpotRename = emma.NewSpotRename("rename", newName)
			case "change-price":
				req.SpotChangePrice = emma.NewSpotChangePrice("changePrice", price)
			default:
				return fmt.Errorf("unknown action %q (supported: start, shutdown, reboot, rename, change-price)", action)
			}

			spot, _, err := c.Client.SpotInstancesAPI.SpotActions(ctx, id).SpotActionsRequest(req).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			if spot == nil {
				fmt.Fprintf(c.Out, "Action %q performed on spot instance %d.\n", action, id)
				return nil
			}
			return output.Render(c.Out, c.OutputFmt, c.NoColor, spot, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "OS", "VCPU", "RAM(GB)", "DATACENTER", "IP"},
				Rows:    [][]string{spotToRow(*spot)},
			})
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "Spot instance ID")
	cmd.Flags().StringVar(&action, "action", "", "Action: start, shutdown, reboot, rename, change-price")
	cmd.Flags().StringVar(&newName, "name", "", "New name (for rename)")
	cmd.Flags().Float32Var(&price, "price", 0, "New price (for change-price)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("action")
	return cmd
}

func (c *CLI) newSpotWaitCmd() *cobra.Command {
	var id int32
	var state string
	var timeout int

	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Wait for a spot instance to reach a target state",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			fmt.Fprintf(c.Err, "Waiting for spot instance %d to reach state %q...\n", id, state)

			err := poller.WaitFor(ctx, 5*time.Second, time.Duration(timeout)*time.Second, func() (bool, error) {
				spot, _, err := c.Client.SpotInstancesAPI.GetSpot(ctx, id).Execute()
				if err != nil {
					if apierrors.IsNotFound(err) || apierrors.IsUnauthorized(err) {
						return false, apierrors.Format(err)
					}
					return false, nil
				}
				if strings.EqualFold(derefStr(spot.Status), state) {
					return true, nil
				}
				fmt.Fprint(c.Err, ".")
				return false, nil
			})
			fmt.Fprintln(c.Err)
			if err != nil {
				return err
			}
			fmt.Fprintf(c.Out, "Spot instance %d reached state %q.\n", id, state)
			return nil
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "Spot instance ID (required)")
	cmd.Flags().StringVar(&state, "state", "", "Target state (required)")
	cmd.Flags().IntVar(&timeout, "timeout", 300, "Timeout in seconds")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("state")
	return cmd
}
