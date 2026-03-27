package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	emma "github.com/emma-community/emma-go-sdk"
	apierrors "github.com/emma-community/emma-cli/internal/apierrors"
	"github.com/emma-community/emma-cli/internal/output"
	"github.com/emma-community/emma-cli/internal/poller"
	"github.com/spf13/cobra"
)

func (c *CLI) newVMCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vm",
		Short: "Manage virtual machines",
	}
	cmd.AddCommand(c.newVMListCmd())
	cmd.AddCommand(c.newVMGetCmd())
	cmd.AddCommand(c.newVMCreateCmd())
	cmd.AddCommand(c.newVMDeleteCmd())
	cmd.AddCommand(c.newVMActionsCmd())
	return cmd
}

func vmToRow(vm emma.Vm) []string {
	name := derefStr(vm.Name)
	id := ""
	if vm.Id != nil {
		id = strconv.Itoa(int(*vm.Id))
	}
	status := derefStr(vm.Status)
	osName := ""
	if vm.Os != nil && vm.Os.Type != nil {
		osName = *vm.Os.Type
	}
	vcpu := ""
	if vm.VCpu != nil {
		vcpu = strconv.Itoa(int(*vm.VCpu))
	}
	ram := ""
	if vm.RamGb != nil {
		ram = strconv.Itoa(int(*vm.RamGb))
	}
	dc := ""
	if vm.DataCenter != nil && vm.DataCenter.Name != nil {
		dc = *vm.DataCenter.Name
	}
	ip := ""
	for _, net := range vm.Networks {
		if net.Ip != nil && *net.Ip != "" {
			ip = *net.Ip
			break
		}
	}
	return []string{name, id, status, osName, vcpu, ram, dc, ip}
}

func (c *CLI) newVMListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List virtual machines",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			req := c.Client.VirtualMachinesAPI.GetVms(ctx)
			if c.ProjectID != nil {
				req = req.ProjectId(*c.ProjectID)
			}
			vms, _, err := req.Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			rows := make([][]string, 0, len(vms))
			for _, vm := range vms {
				rows = append(rows, vmToRow(vm))
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, vms, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "OS", "VCPU", "RAM(GB)", "DATACENTER", "IP"},
				Rows:    rows,
			})
		},
	}
	return cmd
}

func (c *CLI) newVMGetCmd() *cobra.Command {
	var id int32

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a virtual machine by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			vm, _, err := c.Client.VirtualMachinesAPI.GetVm(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, vm, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "OS", "VCPU", "RAM(GB)", "DATACENTER", "IP"},
				Rows:    [][]string{vmToRow(*vm)},
			})
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "VM ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (c *CLI) newVMCreateCmd() *cobra.Command {
	var name, datacenterID, networkType, cloudNetworkType, volumeType, vcpuType string
	var osID, vcpu, ram, volumeSize, sshKeyID int32

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a virtual machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			createReq := emma.VmCreate{
				Name:             name,
				DataCenterId:     datacenterID,
				OsId:             osID,
				CloudNetworkType: cloudNetworkType,
				VCpuType:         vcpuType,
				VCpu:             vcpu,
				RamGb:            ram,
				VolumeType:       volumeType,
				VolumeGb:         volumeSize,
			}
			if cmd.Flags().Changed("ssh-key-id") {
				createReq.SshKeyId = &sshKeyID
			}

			vm, _, err := c.Client.VirtualMachinesAPI.VmCreate(ctx).VmCreate(createReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			fmt.Fprintf(c.Err, "VM %q created (ID: %d). Waiting for IP assignment...\n", derefStr(vm.Name), derefInt32(vm.Id))

			// Poll for IP assignment
			vmID := *vm.Id
			pollErr := poller.WaitFor(ctx, 5*time.Second, 60*time.Second, func() (bool, error) {
				updated, _, err := c.Client.VirtualMachinesAPI.GetVm(ctx, vmID).Execute()
				if err != nil {
					// Propagate permanent errors (VM gone or access revoked); retry transient ones.
					if apierrors.IsNotFound(err) || apierrors.IsUnauthorized(err) {
						return false, apierrors.Format(err)
					}
					return false, nil
				}
				for _, net := range updated.Networks {
					if net.Ip != nil && *net.Ip != "" {
						vm = updated
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

			return output.Render(c.Out, c.OutputFmt, c.NoColor, vm, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "OS", "VCPU", "RAM(GB)", "DATACENTER", "IP"},
				Rows:    [][]string{vmToRow(*vm)},
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "VM name (required)")
	cmd.Flags().StringVar(&datacenterID, "datacenter-id", "", "Datacenter ID (required)")
	cmd.Flags().Int32Var(&osID, "os-id", 0, "OS ID (required)")
	cmd.Flags().Int32Var(&vcpu, "vcpu", 0, "Number of vCPUs (required)")
	cmd.Flags().Int32Var(&ram, "ram", 0, "RAM in GB (required)")
	cmd.Flags().Int32Var(&volumeSize, "volume-size", 0, "Volume size in GB (required)")
	cmd.Flags().StringVar(&volumeType, "volume-type", "", "Volume type (required)")
	cmd.Flags().StringVar(&cloudNetworkType, "cloud-network-type", "", "Cloud network type (required)")
	cmd.Flags().StringVar(&networkType, "network-type", "", "Network type")
	cmd.Flags().StringVar(&vcpuType, "vcpu-type", "shared", "vCPU type (default: shared)")
	cmd.Flags().Int32Var(&sshKeyID, "ssh-key-id", 0, "SSH key ID to inject (see `emma sshkey list`)")

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

func (c *CLI) newVMDeleteCmd() *cobra.Command {
	var id int32
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a virtual machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				fmt.Fprintf(c.Err, "Delete VM %d? [y/N] ", id)
				var input string
				fmt.Fscan(cmd.InOrStdin(), &input)
				if strings.ToLower(strings.TrimSpace(input)) != "y" {
					fmt.Fprintln(c.Out, "Aborted.")
					return nil
				}
			}

			ctx := context.Background()
			_, _, err := c.Client.VirtualMachinesAPI.VmDelete(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			fmt.Fprintf(c.Out, "Deleted VM %d.\n", id)
			return nil
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "VM ID")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (c *CLI) newVMActionsCmd() *cobra.Command {
	var id int32
	var action, newName string

	cmd := &cobra.Command{
		Use:   "actions",
		Short: "Perform an action on a virtual machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			var req emma.VmActionsRequest
			switch strings.ToLower(action) {
			case "clone":
				if newName == "" {
					return fmt.Errorf("--name is required for clone action")
				}
				clone := emma.NewVmClone("clone", newName)
				req.VmClone = clone
			case "start":
				req.VmStart = emma.NewVmStart("start")
			case "shutdown":
				req.VmShutdown = emma.NewVmShutdown("shutdown")
			case "reboot":
				req.VmReboot = emma.NewVmReboot("reboot")
			default:
				return fmt.Errorf("unknown action %q (supported: clone, start, shutdown, reboot)", action)
			}

			vm, _, err := c.Client.VirtualMachinesAPI.VmActions(ctx, id).VmActionsRequest(req).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			if vm == nil {
				fmt.Fprintf(c.Out, "Action %q performed on VM %d.\n", action, id)
				return nil
			}
			return output.Render(c.Out, c.OutputFmt, c.NoColor, vm, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "OS", "VCPU", "RAM(GB)", "DATACENTER", "IP"},
				Rows:    [][]string{vmToRow(*vm)},
			})
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "VM ID")
	cmd.Flags().StringVar(&action, "action", "", "Action to perform (clone, start, shutdown, reboot)")
	cmd.Flags().StringVar(&newName, "name", "", "New name (for clone)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("action")
	return cmd
}

// Helper functions

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}
