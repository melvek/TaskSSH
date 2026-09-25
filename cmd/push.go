package cmd

import (
	"github.com/spf13/cobra"
)

var (
	pushFile  string
	pushDest  string
	pushForce bool
	pushZip   bool
)

var pushCmd = &cobra.Command{
	Use:   "push [hosts...]",
	Short: "Upload a file or directory to remote server",
	Long:  `将本地文件或目录上传到远程服务器。目录上传默认递归，加 -z 则打包上传。`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		with := map[string]any{
			"file":  pushFile,
			"dest":  pushDest,
			"force": pushForce,
			"zip":   pushZip,
		}
		return runBuiltinTask("push", args, with)
	},
}

func init() {
	pushCmd.Flags().StringVarP(&pushFile, "file", "f", "", "Local file or directory")
	pushCmd.Flags().StringVarP(&pushDest, "dest", "d", "", "Remote destination")
	pushCmd.Flags().BoolVarP(&pushForce, "force", "F", false, "Overwrite existing files")
	pushCmd.Flags().BoolVarP(&pushZip, "zip", "z", false, "Zip directory before upload (requires unzip on remote)")
}
