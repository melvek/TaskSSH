package cmd

import (
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
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		with := action.FetchWith{
			File:   fetchFile,
			Dest:   fetchDest,
			Zip:    fetchZip,
			TmpDir: fetchTmpDir,
		}
		return runBuiltinTask(cmd.Context(), "fetch", args, with)
	},
}

func init() {
	fetchCmd.Flags().StringVarP(&fetchFile, "file", "f", "", "Remote file or directory")
	fetchCmd.Flags().StringVarP(&fetchDest, "dest", "d", "./", "Local destination")
	fetchCmd.Flags().BoolVarP(&fetchZip, "zip", "z", false, "Zip directory on remote before download (requires zip on remote)")
	fetchCmd.Flags().StringVarP(&fetchTmpDir, "tmp-dir", "T", "/tmp", "Remote temp directory for zip (only with -z)")
}
