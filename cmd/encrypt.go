package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/prompt"
	"mestrap.com/taskssh/internal/secret"
)

// 加密指定的字符串
var encryptCmd = &cobra.Command{
	Use:   "encrypt [text]",
	Short: "Encrypt a string",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var passwd string
		var err error

		if len(args) > 0 {
			console.Warn("Passing input via command line is insecure and may be recorded in shell history; use interactive input instead")
			passwd = args[0]
		} else {
			passwd, err = prompt.HiddenConfirm("Text to encrypt: ")
			if err != nil {
				return err
			}
		}

		key, err := secret.LoadVaultKey()
		if err != nil {
			return err
		}
		secret.SetVaultKey(key)

		cipher, err := secret.Encrypt(passwd)
		if err != nil {
			return err
		}
		fmt.Println(cipher)
		return nil
	},
}
