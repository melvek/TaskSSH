package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"mestrap.com/taskssh/internal/action"
)

var (
	pushFile  string
	pushDest  string
	pushForce bool
	pushZip   bool
)

var pushCmd = &cobra.Command{
	Use:   "push <host-pattern> --file <file> --dest <dest>",
	Short: "Upload a file or directory to remote server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if pushFile == "" {
			return fmt.Errorf("push requires -f/--file")
		}
		if pushDest == "" {
			return fmt.Errorf("push requires -d/--dest")
		}

		with := action.PushWith{
			File:  pushFile,
			Dest:  pushDest,
			Force: pushForce,
			Zip:   pushZip,
		}
		return runBuiltinTask(cmd.Context(), "push", with, args)
	},
}

func init() {
	pushCmd.Flags().StringVarP(&pushFile, "file", "f", "", "Local file or directory")
	pushCmd.Flags().StringVarP(&pushDest, "dest", "d", "", "Remote destination")
	pushCmd.Flags().BoolVarP(&pushForce, "force", "F", false, "Overwrite existing files")
	pushCmd.Flags().BoolVarP(&pushZip, "zip", "z", false, "Zip directory before upload (requires unzip on remote)")
}
