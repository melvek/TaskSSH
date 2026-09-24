package cmd

import (
	"github.com/spf13/cobra"
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
		with := map[string]any{
			"file":   scriptFile,
			"dest":   scriptDest,
			"remove": scriptRemove,
			"force":  scriptForce,
		}
		return runBuiltinTask("script", args, with)
	},
}

func init() {
	scriptCmd.Flags().StringVarP(&scriptFile, "file", "f", "", "Local script file")
	scriptCmd.Flags().StringVarP(&scriptDest, "dest", "d", "/tmp", "Remote directory")
	scriptCmd.Flags().BoolVarP(&scriptRemove, "remove", "R", false, "Remove script after execution")
	scriptCmd.Flags().BoolVarP(&scriptForce, "force", "F", false, "Overwrite existing script")
}
