package cmd

import (
	"context"
	"fmt"
	"maps"

	"mestrap.com/taskssh/internal/action"
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/resolve"
	"mestrap.com/taskssh/internal/ssh"
	"mestrap.com/taskssh/internal/task"
)

// runBuiltinTask 执行内置任务
func runBuiltinTask(ctx context.Context, taskName string, hosts []string, with any) error {
	inv, err := config.Load(inventoryFile)
	if err != nil {
		return err
	}

	step := config.Step{
		Name:   taskName,
		Action: taskName,
		With:   with,
	}
	t := config.Task{
		Steps: []config.Step{step},
	}

	return executeTask(ctx, &t, hosts, inv)
}

// runDynamicTask 从 inventory.yaml 动态查找自定义的 task
func runDynamicTask(ctx context.Context, args []string) error {
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

	return executeTask(ctx, &t, hostNames, inv)
}

// executeTask 是公共执行入口。
func executeTask(ctx context.Context, t *config.Task, hostNames []string, inv *config.Inventory) error {
	ssh.ClearPasswordCache()
	defer ssh.ClearPasswordCache()

	// dry-run 提醒
	if dryRun {
		console.Warn("Dry run (no changes will be made)")
	}

	cliVarMap, err := parseCLIVars(cliVars)
	if err != nil {
		return err
	}

	ov := task.Overrides{
		Port:         portFlag,
		Username:     userFlag,
		Password:     passwordFlag,
		IdentityFile: identityFileFlag,
	}

	entries, err := task.ResolveHosts(hostNames, inv, ov)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("no target hosts")
	}

	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("%s: %s", e.Name, e.Host.Host))
	}
	console.Targets(lines)

	if listOnly {
		return nil
	}

	// dry run 不需要确认操作
	if !yesFlag && !dryRun {
		if !confirm("Confirm to proceed") {
			return nil
		}
	}

	globalVars := make(resolve.Vars)
	maps.Copy(globalVars, inv.GlobalVars.Extra)

	// 一次性加载所有Action，之后根据任务创建执行器
	actions := action.NewRegistry()
	serial := t.Serial
	if serialFlag >= 0 {
		serial = serialFlag
	}

	minAvailable := t.MinAvailable
	if minAvailableFlag >= 0 {
		minAvailable = minAvailableFlag
	}

	executor := task.NewExecutor(actions, concurrency, connectTimeout, dryRun, cliVarMap,
		serial, minAvailable, stopOnFailureFlag)

	results, err := executor.Run(ctx, t, entries, globalVars)
	if err != nil {
		return err
	}

	success, failed := 0, 0

	if !dryRun {
		// 整理执行结果，输出执行摘要信息
		var failedHosts []string
		for _, r := range results {
			if r.Success {
				success++
			} else {
				failed++
				failedHosts = append(failedHosts, r.Host)
			}
		}

		console.Summary(len(results), success, failed, failedHosts)
	}

	console.ExecutionID(executor.ExecID())

	// if failed > 0 {
	// 	return fmt.Errorf("%d host(s) failed", failed)
	// }

	return nil
}
