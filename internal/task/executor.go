package task

import (
	"fmt"
	"time"

	"mestrap.com/taskssh/internal/action"
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/log"
	"mestrap.com/taskssh/internal/resolve"
)

// Result 是单台主机的执行结果。
type Result struct {
	Host    string
	Success bool
	Error   error
}

// Executor 执行任务。
type Executor struct {
	actions *action.Registry
}

// NewExecutor 创建执行器。
func NewExecutor(actions *action.Registry) *Executor {
	return &Executor{actions: actions}
}

// Run 对一批主机执行任务。
func (e *Executor) Run(task *config.Task, hosts []HostEntry, globalVars resolve.Vars) []Result {
	var results []Result

	for _, entry := range hosts {
		log.EmptyLine()
		log.Info("Processing %s (%s@%s) ", entry.Name, entry.Host.Username, entry.Host.Host)

		err := e.runOnHost(task, &entry.Host, globalVars, entry.Vars)
		results = append(results, Result{
			Host:    entry.Name,
			Success: err == nil,
			Error:   err,
		})

		if err != nil {
			log.Error("%s: %v", entry.Name, err)
		} else {
			log.Success("[OK] %s", entry.Name)
		}
	}

	return results
}

// runOnHost 对单台主机执行整条任务。
func (e *Executor) runOnHost(task *config.Task, host *config.Host,
	globalVars, hostVars resolve.Vars) error {

	// 合并变量池：global -> host
	vars := make(resolve.Vars, len(globalVars)+len(hostVars))
	for k, v := range globalVars {
		vars[k] = v
	}
	for k, v := range hostVars {
		vars[k] = v
	}

	for i, step := range task.Steps {
		log.EmptyLine()
		log.Info("[STEP %d/%d] %s", i+1, len(task.Steps), step.Name)

		act := e.actions.Get(step.Action)
		if act == nil {
			return fmt.Errorf("unknown action: %s", step.Action)
		}

		// 构造执行上下文
		ctx := &action.Context{
			Host: host,
			Vars: vars,
			With: step.With,
		}

		if err := act.Execute(ctx); err != nil {
			return fmt.Errorf("step %s: %w", step.Name, err)
		}

		// delay
		if step.Delay > 0 {
			log.Info("Waiting %ds...", step.Delay)
			time.Sleep(time.Duration(step.Delay) * time.Second)
		}
	}

	return nil
}

// HostEntry 是目标主机的一项。
type HostEntry struct {
	Name string
	Host config.Host
	Vars resolve.Vars
}
