package cmd

import (
	"github.com/spf13/cobra"
	"mestrap.com/taskssh/internal/action"
)

var (
	executeFlag  string
	executeLocal bool
)

// 执行远程命令
// -e/--execute 需要执行的远程命令，支持变量替换
// -L/--local   指定命令在本地环境执行
var commandCmd = &cobra.Command{
	Use:   "command [hosts...]",
	Short: "Execute a remote command",
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
	// 为了配合一些需要在本地执行的命令，比如进行一些复杂任务时，需要对本地文件进行清理，或者执行不同项目的打包之类
	commandCmd.Flags().BoolVarP(&executeLocal, "local", "L", false, "Execute command locally instead of remote")
}
