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

func (c *CLI) newSSHKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sshkey",
		Short: "Manage SSH keys",
	}
	cmd.AddCommand(c.newSSHKeyListCmd())
	cmd.AddCommand(c.newSSHKeyGetCmd())
	cmd.AddCommand(c.newSSHKeyCreateCmd())
	cmd.AddCommand(c.newSSHKeyUpdateCmd())
	cmd.AddCommand(c.newSSHKeyDeleteCmd())
	return cmd
}

func sshKeyToRow(k emma.SshKey) []string {
	name := derefStr(k.Name)
	id := ""
	if k.Id != nil {
		id = strconv.Itoa(int(*k.Id))
	}
	keyType := derefStr(k.KeyType)
	fingerprint := derefStr(k.Fingerprint)
	createdAt := derefStr(k.CreatedAt)
	return []string{name, id, keyType, fingerprint, createdAt}
}

func (c *CLI) newSSHKeyListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List SSH keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			keys, _, err := c.Client.SSHKeysAPI.SshKeys(ctx).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			rows := make([][]string, 0, len(keys))
			for _, k := range keys {
				rows = append(rows, sshKeyToRow(k))
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, keys, output.TableView{
				Headers: []string{"NAME", "ID", "TYPE", "FINGERPRINT", "CREATED-AT"},
				Rows:    rows,
			})
		},
	}
	return cmd
}

func (c *CLI) newSSHKeyGetCmd() *cobra.Command {
	var id int32

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get an SSH key by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			key, _, err := c.Client.SSHKeysAPI.GetSshKey(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, key, output.TableView{
				Headers: []string{"NAME", "ID", "TYPE", "FINGERPRINT", "CREATED-AT"},
				Rows:    [][]string{sshKeyToRow(*key)},
			})
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "SSH key ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func (c *CLI) newSSHKeyCreateCmd() *cobra.Command {
	var name, publicKey, keyType string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create or import an SSH key",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			var createReq emma.SshKeysCreateImportRequest
			if publicKey != "" {
				importKey := emma.SshKeyImport{
					Name: name,
					Key:  publicKey,
				}
				createReq.SshKeyImport = &importKey
			} else {
				if keyType == "" {
					keyType = "RSA"
				}
				newKey := emma.SshKeyCreate{
					Name:    name,
					KeyType: keyType,
				}
				createReq.SshKeyCreate = &newKey
			}

			resp, _, err := c.Client.SSHKeysAPI.SshKeysCreateImport(ctx).SshKeysCreateImportRequest(createReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			var key emma.SshKey
			if resp.SshKey != nil {
				key = *resp.SshKey
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, resp, output.TableView{
				Headers: []string{"NAME", "ID", "TYPE", "FINGERPRINT", "CREATED-AT"},
				Rows:    [][]string{sshKeyToRow(key)},
			})
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Key name (required)")
	cmd.Flags().StringVar(&publicKey, "key", "", "Public key content (optional, generates new key if omitted)")
	cmd.Flags().StringVar(&keyType, "key-type", "RSA", "Key type for new key: RSA or ED25519")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func (c *CLI) newSSHKeyUpdateCmd() *cobra.Command {
	var id int32
	var name string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an SSH key name",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			updateReq := emma.SshKeyUpdate{
				Name: name,
			}

			key, _, err := c.Client.SSHKeysAPI.SshKeyUpdate(ctx, id).SshKeyUpdate(updateReq).Execute()
			if err != nil {
				return apierrors.Format(err)
			}

			return output.Render(c.Out, c.OutputFmt, c.NoColor, key, output.TableView{
				Headers: []string{"NAME", "ID", "TYPE", "FINGERPRINT", "CREATED-AT"},
				Rows:    [][]string{sshKeyToRow(*key)},
			})
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "SSH key ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "New name (required)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func (c *CLI) newSSHKeyDeleteCmd() *cobra.Command {
	var id int32
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an SSH key",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmdutil.ConfirmDelete(c.Err, cmd.InOrStdin(), "SSH key", id, yes) {
				fmt.Fprintln(c.Out, "Aborted.")
				return nil
			}

			ctx := context.Background()
			_, err := c.Client.SSHKeysAPI.SshKeyDelete(ctx, id).Execute()
			if err != nil {
				return apierrors.Format(err)
			}
			fmt.Fprintf(c.Out, "Deleted SSH key %d.\n", id)
			return nil
		},
	}

	cmd.Flags().Int32Var(&id, "id", 0, "SSH key ID")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}
