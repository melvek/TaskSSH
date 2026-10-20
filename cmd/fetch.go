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
	Use:   "fetch <host-pattern> --file <file> [--dest <dest>]",
	Short: "Download a file or directory from remote server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if fetchFile == "" {
			return fmt.Errorf("fetch requires -f/--file")
		}

		if fetchDest == "" {
			if cmd.Flags().Changed("dest") {
				return fmt.Errorf("fetch -d/--dest cannot be empty")
			}
			// 如果没有指定本地路径，则使用当前路径
			fetchDest = "./"
		}

		with := action.FetchWith{
			File:   fetchFile,
			Dest:   fetchDest,
			Zip:    fetchZip,
			TmpDir: fetchTmpDir,
		}
		return runBuiltinTask(cmd.Context(), "fetch", with, args)
	},
}

func init() {
	fetchCmd.Flags().StringVarP(&fetchFile, "file", "f", "", "Remote file or directory")
	fetchCmd.Flags().StringVarP(&fetchDest, "dest", "d", "", "Local destination")
	fetchCmd.Flags().BoolVarP(&fetchZip, "zip", "z", false, "Zip directory on remote before download (requires zip on remote)")
	fetchCmd.Flags().StringVarP(&fetchTmpDir, "tmp-dir", "T", "/tmp", "Remote temp directory for zip (only with -z)")
}
