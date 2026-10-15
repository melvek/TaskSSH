package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"mestrap.com/taskssh/internal/action"
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
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if scriptFile == "" {
			if len(args) > 0 {
				scriptFile = args[0]
			} else {
				return fmt.Errorf("push action requires 'file' parameter")
			}
		}
		if scriptDest == "" {
			if len(args) > 1 {
				scriptDest = args[1]
			} else {
				scriptDest = "/tmp"
			}
		}
		with := action.ScriptWith{
			File:   scriptFile,
			Dest:   scriptDest,
			Remove: scriptRemove,
			Force:  scriptForce,
		}
		return runBuiltinTask(cmd.Context(), "script", with)
	},
}

func init() {
	scriptCmd.Flags().StringVarP(&scriptFile, "file", "f", "", "Local script file")
	scriptCmd.Flags().StringVarP(&scriptDest, "dest", "d", "", "Remote directory (default /tmp)")
	scriptCmd.Flags().BoolVarP(&scriptRemove, "remove", "r", false, "Remove script after execution")
	scriptCmd.Flags().BoolVarP(&scriptForce, "force", "F", false, "Overwrite existing script")
}
