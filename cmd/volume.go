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

func (c *CLI) newVolumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume",
		Short: "Manage volumes",
	}
	cmd.AddCommand(c.newVolumeListCmd())
	cmd.AddCommand(c.newVolumeGetCmd())
	cmd.AddCommand(c.newVolumeCreateCmd())
	cmd.AddCommand(c.newVolumeDeleteCmd())
	cmd.AddCommand(c.newVolumeAttachCmd())
	cmd.AddCommand(c.newVolumeDetachCmd())
	return cmd
}

func volumeToRow(v emma.Volume) []string {
	name := derefStr(v.Name)
	id := ""
	if v.Id != nil {
		id = strconv.Itoa(int(*v.Id))
	}
	status := derefStr(v.Status)
	size := ""
	if v.SizeGb != nil {
		size = strconv.Itoa(int(*v.SizeGb))
	}
	volType := derefStr(v.Type)
	dc := ""
	if v.DataCenter != nil && v.DataCenter.Name != nil {
		dc = *v.DataCenter.Name
	}
	attachedTo := ""
	if v.AttachedToId != nil {
		attachedTo = strconv.Itoa(int(*v.AttachedToId))
	}
	return []string{name, id, status, size, volType, dc, attachedTo}
}

func (c *CLI) newVolumeListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List volumes",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			// GetVolumes doesn't support ProjectId filter — project determined by auth
			volumes, _, err := c.Client.VolumesAPI.GetVolumes(ctx).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			rows := make([][]string, 0, len(volumes))
			for _, v := range volumes {
				rows = append(rows, volumeToRow(v))
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, volumes, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "SIZE(GB)", "TYPE", "DATACENTER", "ATTACHED-TO"},
				Rows:    rows,
			})
		},
	}
	return cmd
}

func (c *CLI) newVolumeGetCmd() *cobra.Command {
	var id int32

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a volume by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			vol, _, err := c.Client.VolumesAPI.GetVolume(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, vol, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "SIZE(GB)", "TYPE", "DATACENTER", "ATTACHED-TO"},
				Rows:    [][]string{volumeToRow(*vol)},
			})
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "Volume ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (c *CLI) newVolumeCreateCmd() *cobra.Command {
	var datacenterID, volumeType string
	var size int32

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a volume",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			createReq := emma.VolumeCreate{
				DataCenterId: datacenterID,
				VolumeType:   volumeType,
				VolumeGb:     size,
			}

			vol, _, err := c.Client.VolumesAPI.VolumeCreate(ctx).VolumeCreate(createReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, vol, output.TableView{
				Headers: []string{"NAME", "ID", "STATUS", "SIZE(GB)", "TYPE", "DATACENTER", "ATTACHED-TO"},
				Rows:    [][]string{volumeToRow(*vol)},
			})
		},
	}

	cmd.Flags().StringVar(&datacenterID, "datacenter-id", "", "Datacenter ID (required)")
	cmd.Flags().StringVar(&volumeType, "type", "", "Volume type (required)")
	cmd.Flags().Int32Var(&size, "size", 0, "Size in GB (required)")

	_ = cmd.MarkFlagRequired("datacenter-id")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("size")

	return cmd
}

func (c *CLI) newVolumeDeleteCmd() *cobra.Command {
	var id int32
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a volume",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmdutil.ConfirmDelete(c.Err, cmd.InOrStdin(), "volume", id, yes) {
				fmt.Fprintln(c.Out, "Aborted.")
				return nil
			}

			ctx := context.Background()
			_, _, err := c.Client.VolumesAPI.VolumeDelete(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			fmt.Fprintf(c.Out, "Deleted volume %d.\n", id)
			return nil
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "Volume ID")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (c *CLI) newVolumeAttachCmd() *cobra.Command {
	var volumeID, vmID int32

	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach a volume to a VM",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			attachReq := emma.VolumeAttach{
				Action:   "attach",
				VolumeId: volumeID,
			}
			actionReq := emma.VmActionsRequest{
				VolumeAttach: &attachReq,
			}

			_, _, err := c.Client.VirtualMachinesAPI.VmActions(ctx, vmID).VmActionsRequest(actionReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			fmt.Fprintf(c.Out, "Attached volume %d to VM %d.\n", volumeID, vmID)
			return nil
		},
	}

	cmd.Flags().Int32Var(&volumeID, "volume-id", 0, "Volume ID")
	cmd.Flags().Int32Var(&vmID, "vm-id", 0, "VM ID")
	_ = cmd.MarkFlagRequired("volume-id")
	_ = cmd.MarkFlagRequired("vm-id")
	return cmd
}

func (c *CLI) newVolumeDetachCmd() *cobra.Command {
	var volumeID, vmID int32

	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Detach a volume from a VM",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			detachReq := emma.VolumeDetach{
				Action:   "detach",
				VolumeId: volumeID,
			}
			actionReq := emma.VmActionsRequest{
				VolumeDetach: &detachReq,
			}

			_, _, err := c.Client.VirtualMachinesAPI.VmActions(ctx, vmID).VmActionsRequest(actionReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			fmt.Fprintf(c.Out, "Detached volume %d from VM %d.\n", volumeID, vmID)
			return nil
		},
	}

	cmd.Flags().Int32Var(&volumeID, "volume-id", 0, "Volume ID")
	cmd.Flags().Int32Var(&vmID, "vm-id", 0, "VM ID")
	_ = cmd.MarkFlagRequired("volume-id")
	_ = cmd.MarkFlagRequired("vm-id")
	return cmd
}
