package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mestrap.com/taskssh/internal/config"
)

// runDynamicTask 从 inventory.yaml 动态查找 task。
func runDynamicTask(cmd *cobra.Command, args []string) error {
	inv, err := config.Load(inventoryFile)
	if err != nil {
		return err
	}

	taskName := args[0]
	t, ok := inv.Tasks[taskName]
	if !ok {
		return fmt.Errorf("task not found: %s", taskName)
	}

	hostNames := args[1:]
	if len(hostNames) == 0 {
		return fmt.Errorf("no target hosts")
	}

	return executeTask(&t, hostNames, inv)
}
