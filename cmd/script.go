package cmd

import (
	"github.com/spf13/cobra"
	"mestrap.com/taskssh/internal/action"
)

var (
	scriptFile   string
	scriptDest   string
	scriptRemove bool
	scriptForce  bool
)

var scriptCmd = &cobra.Command{
	Use:   "script [hosts...]",
	Short: "Execute a local shell script on remote hosts",
	Long:  `将本地 shell 脚本上传到目标主机并执行。`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		with := action.ScriptWith{
			File:   scriptFile,
			Dest:   scriptDest,
			Remove: scriptRemove,
			Force:  scriptForce,
		}
		return runBuiltinTask(cmd.Context(), "script", args, with)
	},
}

func init() {
	scriptCmd.Flags().StringVarP(&scriptFile, "file", "f", "", "Local script file")
	scriptCmd.Flags().StringVarP(&scriptDest, "dest", "d", "/tmp", "Remote directory")
	scriptCmd.Flags().BoolVarP(&scriptRemove, "remove", "r", false, "Remove script after execution")
	scriptCmd.Flags().BoolVarP(&scriptForce, "force", "F", false, "Overwrite existing script")
}
