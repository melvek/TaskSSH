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
	Use:   "push [hosts...]",
	Short: "Upload a file or directory to remote server",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if pushFile == "" {
			if len(args) > 0 {
				pushFile = args[0]
			} else {
				return fmt.Errorf("push action requires 'file' parameter")
			}
		}
		if pushDest == "" {
			if len(args) > 1 {
				pushDest = args[1]
			} else {
				return fmt.Errorf("push action requires 'dest' parameter")
			}
		}

		with := action.PushWith{
			File:  pushFile,
			Dest:  pushDest,
			Force: pushForce,
			Zip:   pushZip,
		}
		return runBuiltinTask(cmd.Context(), "push", with)
	},
}

func init() {
	pushCmd.Flags().StringVarP(&pushFile, "file", "f", "", "Local file or directory")
	pushCmd.Flags().StringVarP(&pushDest, "dest", "d", "", "Remote destination")
	pushCmd.Flags().BoolVarP(&pushForce, "force", "F", false, "Overwrite existing files")
	pushCmd.Flags().BoolVarP(&pushZip, "zip", "z", false, "Zip directory before upload (requires unzip on remote)")
}
