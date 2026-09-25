package cmd

import (
	"github.com/spf13/cobra"
)

var (
	fetchFile   string
	fetchDest   string
	fetchZip    bool
	fetchTmpDir string
)

var fetchCmd = &cobra.Command{
	Use:   "fetch [hosts...]",
	Short: "Download a file or directory from remote server",
	Long:  `从远程服务器下载文件或目录。目录下载默认递归，加 -z 则远程打包后下载。本地已存在文件默认覆盖。`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		with := map[string]any{
			"file":    fetchFile,
			"dest":    fetchDest,
			"zip":     fetchZip,
			"tmp_dir": fetchTmpDir,
		}
		return runBuiltinTask("fetch", args, with)
	},
}

func init() {
	fetchCmd.Flags().StringVarP(&fetchFile, "file", "f", "", "Remote file or directory")
	fetchCmd.Flags().StringVarP(&fetchDest, "dest", "d", "./", "Local destination")
	fetchCmd.Flags().BoolVarP(&fetchZip, "zip", "z", false, "Zip directory on remote before download (requires zip on remote)")
	fetchCmd.Flags().StringVarP(&fetchTmpDir, "tmp-dir", "T", "/tmp", "Remote temp directory for zip (only with -z)")
}
