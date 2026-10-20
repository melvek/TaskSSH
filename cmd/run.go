package cmd

import (
	"context"
	"fmt"
	"maps"

	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/host"
	"mestrap.com/taskssh/internal/resolve"
	"mestrap.com/taskssh/internal/ssh"
	"mestrap.com/taskssh/internal/task"
)

// runBuiltinTask 执行内置任务
func runBuiltinTask(ctx context.Context, taskName string, with any, args []string) error {
	if err := setTargetHostPattern(args); err != nil {
		return err
	}

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

	return executeTask(ctx, &t, inv)
}

// runDynamicTask 从 inventory.yaml 动态查找自定义的 task
func runDynamicTask(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task name is required")
	}
	if len(args) != 2 {
		return fmt.Errorf("expected task name and one host-pattern argument")
	}
	if err := setTargetHostPattern(args[1:]); err != nil {
		return err
	}

	inv, err := config.Load(inventoryFile)
	if err != nil {
		return err
	}

	taskName := args[0]
	t, ok := inv.Tasks[taskName]
	if !ok {
		return fmt.Errorf("task not found: %s", taskName)
	}

	return executeTask(ctx, &t, inv)
}

// executeTask 是公共执行入口。
func executeTask(ctx context.Context, t *config.Task, inv *config.Inventory) error {
	ssh.ClearPasswordCache()
	defer ssh.ClearPasswordCache()

	// 未指定具体主机时，报错，因为目前默认是不做确认的，为防止忘了写，导致对全部主机执行一些不可逆的操作
	// 如果确实需要对所有服务器执行操作，需要显式地指定 "*"
	if targetHostPattern == "" {
		return fmt.Errorf("no target hosts, specify one host pattern as a positional argument")
	}

	// dry-run 提醒
	if dryRun {
		console.Default.Warn("Dry run (no changes will be made)")
	}

	cliVarMap, err := parseCLIVars(cliVars)
	if err != nil {
		return err
	}

	ov := host.Overrides{
		Port:         portFlag,
		Username:     userFlag,
		Password:     passwordFlag,
		IdentityFile: identityFileFlag,
	}

	// 解析主机匹配
	entries, err := host.ResolveHosts(targetHostPattern, inv, ov)

	if err != nil {
		return err
	}

	// 如果解析后，未匹配到任何服务器，则报错退出
	if len(entries) == 0 {
		return fmt.Errorf("no target hosts")
	}

	if listOnly || confirmFlag || dryRun {
		lines := make([]string, 0, len(entries))
		for _, e := range entries {
			if e.Name == e.Host.Host {
				lines = append(lines, fmt.Sprintf("%s", e.Name))
			} else {
				lines = append(lines, fmt.Sprintf("%s: %s", e.Name, e.Host.Host))
			}
		}
		console.Default.Targets(lines)
	}

	if listOnly {
		return nil
	}

	// dry run 不需要确认操作
	if confirmFlag && !dryRun {
		if !console.Confirm("Confirm to proceed") {
			return nil
		}
	}

	globalVars := make(resolve.Vars)
	maps.Copy(globalVars, inv.Defaults.Extra)

	// 自动分批处理
	serial := t.Serial
	if serialFlag >= 0 {
		serial = serialFlag
	}

	minAvailable := t.MinAvailable
	if minAvailableFlag >= 0 {
		minAvailable = minAvailableFlag
	}

	executor := task.NewExecutor(concurrency, connectTimeout, dryRun, cliVarMap,
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

		console.Default.Summary(len(results), success, failed, failedHosts)
	}

	console.Default.ExecutionID(executor.ExecID())

	return nil
}
