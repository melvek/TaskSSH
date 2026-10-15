package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"mestrap.com/taskssh/internal/action"
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
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if fetchFile == "" {
			if len(args) > 0 {
				fetchFile = args[0]
			} else {
				return fmt.Errorf("push action requires 'file' parameter")
			}
		}

		if fetchDest == "" {
			if len(args) > 1 {
				fetchDest = args[1]
			} else {
				// 如果没有指定本地路径，则使用当前路径
				fetchDest = "./"
			}
		}

		with := action.FetchWith{
			File:   fetchFile,
			Dest:   fetchDest,
			Zip:    fetchZip,
			TmpDir: fetchTmpDir,
		}
		return runBuiltinTask(cmd.Context(), "fetch", with)
	},
}

func init() {
	fetchCmd.Flags().StringVarP(&fetchFile, "file", "f", "", "Remote file or directory")
	fetchCmd.Flags().StringVarP(&fetchDest, "dest", "d", "", "Local destination")
	fetchCmd.Flags().BoolVarP(&fetchZip, "zip", "z", false, "Zip directory on remote before download (requires zip on remote)")
	fetchCmd.Flags().StringVarP(&fetchTmpDir, "tmp-dir", "T", "/tmp", "Remote temp directory for zip (only with -z)")
}
