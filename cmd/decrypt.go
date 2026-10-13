package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"mestrap.com/taskssh/internal/prompt"
	"mestrap.com/taskssh/internal/secret"
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt [cipher]",
	Short: "Decrypt a string",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var cipherText string
		var err error

		if len(args) > 0 {
			cipherText = args[0]
		} else {
			cipherText, err = prompt.Line("Ciphertext: ")
			if err != nil {
				return err
			}
		}

		plain, err := secret.Decrypt(cipherText)
		if err != nil {
			return err
		}
		fmt.Println(plain)
		return nil
	},
}
