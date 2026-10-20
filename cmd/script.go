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
	Use:   "script <host-pattern> --file <file> [--dest <dest>]",
	Short: "Execute a local shell script on remote hosts",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if scriptFile == "" {
			return fmt.Errorf("script requires -f/--file")
		}
		if scriptDest == "" {
			if cmd.Flags().Changed("dest") {
				return fmt.Errorf("script -d/--dest cannot be empty")
			}
			scriptDest = "/tmp"
		}
		with := action.ScriptWith{
			File:   scriptFile,
			Dest:   scriptDest,
			Remove: scriptRemove,
			Force:  scriptForce,
		}
		return runBuiltinTask(cmd.Context(), "script", with, args)
	},
}

func init() {
	scriptCmd.Flags().StringVarP(&scriptFile, "file", "f", "", "Local script file")
	scriptCmd.Flags().StringVarP(&scriptDest, "dest", "d", "", "Remote directory (default /tmp)")
	scriptCmd.Flags().BoolVarP(&scriptRemove, "remove", "r", false, "Remove script after execution")
	scriptCmd.Flags().BoolVarP(&scriptForce, "force", "F", false, "Overwrite existing script")
}
