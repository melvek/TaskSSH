package cmd

import (
	"github.com/spf13/cobra"
)

var (
	pushFile   string
	pushDest   string
	pushForce  bool
	pushBackup bool
)

var pushCmd = &cobra.Command{
	Use:   "push [hosts...]",
	Short: "Upload a file to remote server",
	Long:  `将本地文件上传到远程服务器。`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		with := map[string]any{
			"file":   pushFile,
			"dest":   pushDest,
			"force":  pushForce,
			"backup": pushBackup,
		}
		return runBuiltinTask("push", args, with)
	},
}

func init() {
	pushCmd.Flags().StringVarP(&pushFile, "file", "f", "", "Local file")
	pushCmd.Flags().StringVarP(&pushDest, "dest", "d", "", "Remote destination")
	pushCmd.Flags().BoolVarP(&pushForce, "force", "F", false, "Overwrite existing files")
	pushCmd.Flags().BoolVarP(&pushBackup, "backup", "B", false, "Backup before overwrite")
}
