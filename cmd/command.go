package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mestrap.com/taskssh/internal/action"
)

var executeFlag string

var commandCmd = &cobra.Command{
	Use:   "command [hosts...]",
	Short: "Execute a remote command",
	Long:  `在目标主机上执行指定的远程命令。`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if executeFlag == "" {
			return fmt.Errorf("command requires -e/--execute")
		}

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

// 让 action 包被引用，避免未使用
var _ = action.CommandAction{}
