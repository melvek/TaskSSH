package cmd

import (
	"github.com/spf13/cobra"
)

var (
	fetchFile   string
	fetchDest   string
	fetchForce  bool
	fetchBackup bool
)

var fetchCmd = &cobra.Command{
	Use:   "fetch [hosts...]",
	Short: "Download a file from remote server",
	Long:  `从远程服务器下载文件到本地。`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		with := map[string]any{
			"file":   fetchFile,
			"dest":   fetchDest,
			"force":  fetchForce,
			"backup": fetchBackup,
		}
		return runBuiltinTask("fetch", args, with)
	},
}

func init() {
	fetchCmd.Flags().StringVarP(&fetchFile, "file", "f", "", "Remote file")
	fetchCmd.Flags().StringVarP(&fetchDest, "dest", "d", "./", "Local destination")
	fetchCmd.Flags().BoolVarP(&fetchForce, "force", "F", false, "Overwrite existing files")
	fetchCmd.Flags().BoolVarP(&fetchBackup, "backup", "B", false, "Backup before overwrite")
}
