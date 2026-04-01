package cmd

import (
	"context"
	"fmt"

	emma "github.com/emma-community/emma-go-sdk"
	apierrors "github.com/emma-community/emma-cli/internal/apierrors"
	"github.com/emma-community/emma-cli/internal/output"
	"github.com/spf13/cobra"
)

func (c *CLI) newMCNCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcn",
		Short: "Manage multi-cloud networks",
	}
	cmd.AddCommand(c.newMCNListCmd())
	cmd.AddCommand(c.newMCNEnableNetworkCmd())
	cmd.AddCommand(c.newMCNDisableNetworkCmd())
	cmd.AddCommand(c.newMCNCloudConnectCmd())
	cmd.AddCommand(c.newMCNCrossConnectCmd())
	return cmd
}

func (c *CLI) newMCNListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List multi-cloud network configuration",
		Long:  "Shows Direct Connect network topology. Use --output json for full nested detail.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			mcn, _, err := c.Client.NetworkAPI.GetMultiCloudNetworks(ctx).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			// Summary table: show networks by macro region
			rows := [][]string{}
			if mcn != nil && mcn.DirectConnect != nil {
				for _, net := range mcn.DirectConnect.Networks {
					region := derefStr(net.MacroRegion)
					status := derefStr(net.Status)
					cloudConnects := fmt.Sprintf("%d", len(net.CloudConnects))
					rows = append(rows, []string{region, status, cloudConnects})
				}
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, mcn, output.TableView{
				Headers: []string{"REGION", "STATUS", "CLOUD-CONNECTS"},
				Rows:    rows,
			})
		},
	}
	return cmd
}

func (c *CLI) newMCNEnableNetworkCmd() *cobra.Command {
	var macroRegion string
	var connectivityCenterID int32

	cmd := &cobra.Command{
		Use:   "enable-network",
		Short: "Enable a direct network in a macro region",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			req := emma.PostMultiCloudNetworkActionRequest{
				McnDirectNetworkEnable: &emma.McnDirectNetworkEnable{
					NetworkType:          "directConnect",
					MacroRegion:          macroRegion,
					ConnectivityCenterId: connectivityCenterID,
					Action:               "enableNetwork",
				},
			}

			result, _, err := c.Client.NetworkAPI.PostMultiCloudNetworkAction(ctx).PostMultiCloudNetworkActionRequest(req).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			fmt.Fprintf(c.Out, "Network enabled in region %s.\n", macroRegion)
			_ = result
			return nil
		},
	}

	cmd.Flags().StringVar(&macroRegion, "macro-region", "", "Macro region: EMEA, AMER, APAC (required)")
	cmd.Flags().Int32Var(&connectivityCenterID, "connectivity-center-id", 0, "Connectivity center ID (required)")
	_ = cmd.MarkFlagRequired("macro-region")
	_ = cmd.MarkFlagRequired("connectivity-center-id")
	return cmd
}

func (c *CLI) newMCNDisableNetworkCmd() *cobra.Command {
	var macroRegion string

	cmd := &cobra.Command{
		Use:   "disable-network",
		Short: "Disable a direct network in a macro region",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			req := emma.PostMultiCloudNetworkActionRequest{
				McnDirectNetworkDisable: &emma.McnDirectNetworkDisable{
					NetworkType: "directConnect",
					MacroRegion: macroRegion,
					Action:      "disableNetwork",
				},
			}

			_, _, err := c.Client.NetworkAPI.PostMultiCloudNetworkAction(ctx).PostMultiCloudNetworkActionRequest(req).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			fmt.Fprintf(c.Out, "Network disabled in region %s.\n", macroRegion)
			return nil
		},
	}

	cmd.Flags().StringVar(&macroRegion, "macro-region", "", "Macro region: EMEA, AMER, APAC (required)")
	_ = cmd.MarkFlagRequired("macro-region")
	return cmd
}

func (c *CLI) newMCNCloudConnectCmd() *cobra.Command {
	var macroRegion, provider, action string

	cmd := &cobra.Command{
		Use:   "cloud-connect",
		Short: "Enable or disable a cloud connection in a region",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			req := emma.PostMultiCloudNetworkActionRequest{
				McnDirectCloudConnectAction: &emma.McnDirectCloudConnectAction{
					NetworkType:             "directConnect",
					MacroRegion:             macroRegion,
					CloudConnectionProvider: provider,
					Action:                  action,
				},
			}

			_, _, err := c.Client.NetworkAPI.PostMultiCloudNetworkAction(ctx).PostMultiCloudNetworkActionRequest(req).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			fmt.Fprintf(c.Out, "Cloud connect %s for provider %s in region %s.\n", action, provider, macroRegion)
			return nil
		},
	}

	cmd.Flags().StringVar(&macroRegion, "macro-region", "", "Macro region (required)")
	cmd.Flags().StringVar(&provider, "provider", "", "Cloud provider name (required)")
	cmd.Flags().StringVar(&action, "action", "", "Action: enableCloudConnect or disableCloudConnect (required)")
	_ = cmd.MarkFlagRequired("macro-region")
	_ = cmd.MarkFlagRequired("provider")
	_ = cmd.MarkFlagRequired("action")
	return cmd
}

func (c *CLI) newMCNCrossConnectCmd() *cobra.Command {
	var macroRegion1, macroRegion2, action string

	cmd := &cobra.Command{
		Use:   "cross-connect",
		Short: "Enable or disable cross-region connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			req := emma.PostMultiCloudNetworkActionRequest{
				McnDirectCrossConnectAction: &emma.McnDirectCrossConnectAction{
					NetworkType: "directConnect",
					MacroRegions: []string{macroRegion1, macroRegion2},
					Action:       action,
				},
			}

			_, _, err := c.Client.NetworkAPI.PostMultiCloudNetworkAction(ctx).PostMultiCloudNetworkActionRequest(req).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			fmt.Fprintf(c.Out, "Cross-connect %s between %s and %s.\n", action, macroRegion1, macroRegion2)
			return nil
		},
	}

	cmd.Flags().StringVar(&macroRegion1, "region-1", "", "First macro region (required)")
	cmd.Flags().StringVar(&macroRegion2, "region-2", "", "Second macro region (required)")
	cmd.Flags().StringVar(&action, "action", "", "Action: enableCrossRegionConnect or disableCrossRegionConnect (required)")
	_ = cmd.MarkFlagRequired("region-1")
	_ = cmd.MarkFlagRequired("region-2")
	_ = cmd.MarkFlagRequired("action")
	return cmd
}
