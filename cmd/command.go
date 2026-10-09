package cmd

import (
	"github.com/spf13/cobra"
	"mestrap.com/taskssh/internal/action"
)

var (
	executeFlag  string
	executeLocal bool
)
var commandCmd = &cobra.Command{
	Use:   "command [hosts...]",
	Short: "Execute a remote command",
	Long:  `在目标主机上执行指定的远程命令。`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		with := action.CommandWith{
			Command: executeFlag,
			Local:   executeLocal,
		}
		return runBuiltinTask(cmd.Context(), "command", args, with)
	},
}

func init() {
	commandCmd.Flags().StringVarP(&executeFlag, "execute", "e", "", "Command to execute")
	commandCmd.Flags().BoolVarP(&executeLocal, "local", "L", false, "Execute command locally instead of remote")
}
