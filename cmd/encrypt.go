package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mestrap.com/taskssh/internal/secret"
)

var encryptCmd = &cobra.Command{
	Use:   "encrypt <text>",
	Short: "Encrypt a string",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cipher, err := secret.Encrypt(args[0])
		if err != nil {
			return err
		}
		fmt.Println(cipher)
		return nil
	},
}

var decryptCmd = &cobra.Command{
	Use:   "decrypt <cipher>",
	Short: "Decrypt a string",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		plain, err := secret.Decrypt(args[0])
		if err != nil {
			return err
		}
		fmt.Println(plain)
		return nil
	},
}
