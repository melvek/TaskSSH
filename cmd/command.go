package cmd

import (
	"github.com/spf13/cobra"
)

var executeFlag string

var commandCmd = &cobra.Command{
	Use:   "command [hosts...]",
	Short: "Execute a remote command",
	Long:  `在目标主机上执行指定的远程命令。`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		with := map[string]any{
			"command": executeFlag,
		}
		return runBuiltinTask("command", args, with)
	},
}

func init() {
	commandCmd.Flags().StringVarP(&executeFlag, "execute", "e",
		"", "Command to execute")
}
