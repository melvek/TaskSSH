package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"mestrap.com/taskssh/internal/fileutil"
	"mestrap.com/taskssh/internal/secret"
)

var keygenForce bool

var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate a new vault key",
	Long: `Generate a new vault key.

The key is written to ~/.taskssh/vault-key by default.
Use -V to specify a different path.

WARNING: overwriting an existing key makes all existing
encrypted secrets unrecoverable.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := fileutil.ExpandHome(secret.DefaultKeyFile)
		if secret.SecretKeyFile != "" {
			path = secret.SecretKeyFile
		}

		if fileutil.FileExists(path) && !keygenForce {
			return fmt.Errorf(
				"vault key already exists: %s\n"+
					"       use --force to overwrite (existing encrypted secrets will become unreadable)",
				path)
		}

		if err := secret.GenerateKeyFile(path); err != nil {
			return err
		}

		fmt.Printf("Vault key generated: %s\n", path)
		fmt.Println("Keep this file safe. If lost, existing encrypted secrets cannot be recovered.")
		return nil
	},
}

func init() {
	keygenCmd.Flags().BoolVarP(&keygenForce, "force", "F", false, "Overwrite existing key file")
}
