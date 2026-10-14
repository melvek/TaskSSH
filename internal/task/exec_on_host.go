package task

import (
	"context"
	"fmt"
	"maps"
	"time"

	"mestrap.com/taskssh/internal/action"
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/resolve"
	"mestrap.com/taskssh/internal/ssh"
)

// runOnHost 对单台主机执行整条任务。
//
// 变量合并顺序：
//
//	global_vars -> 组 vars -> 主机 Extra -> 标准变量 -> CLI -D -> 运行时（date / exec_id）
func (e *Executor) runOnHost(ctx context.Context, task *config.Task, entry HostEntry,
	globalVars resolve.Vars) error {

	// 1. 清单变量：global -> 组 -> 主机
	vars := mergeVars(globalVars, entry.Vars)

	// 2. 标准变量（host_name / group_name / host_addr / host_port / host_user）
	maps.Copy(vars, entry.StandardVars())

	// 3. CLI -D
	maps.Copy(vars, e.cliVars)

	// 4. 运行时变量（不可覆盖）
	vars["exec_id"] = e.execID

	var client *ssh.Client

	if !e.dryRun {
		c, err := ssh.Connect(&entry.Host, e.connectTimeout)
		if err != nil {
			console.Error("%v", err)
			return err
		}
		client = c
		defer client.Close()
	}

	for i, step := range task.Steps {

		act := e.actions.Get(step.Action)
		if act == nil {
			console.Error("unknown action: %s", step.Action)
			return fmt.Errorf("unknown action: %s", step.Action)
		}

		if len(task.Steps) > 1 {
			console.Step(i+1, len(task.Steps), step.Name)
		}

		if step.When != "" {
			ok, expanded, err := evalWhen(step.When, vars)
			if err != nil {
				console.Error("eval when: %v", err)
				return fmt.Errorf("step %s: eval when: %w", step.Name, err)
			}
			if !ok {
				console.Skipped("ignore: %s", expanded)
				continue
			}
		}

		actx := &action.ActionContext{
			Host:   &entry.Host,
			Vars:   vars,
			With:   step.With,
			Client: client,
			DryRun: e.dryRun,
		}

		err := withRetry(ctx, func() error {
			return act.Execute(actx)
		}, step.Interval, step.Timeout)

		if err != nil {
			if step.IgnoreErrors {
				console.Ignored("Ignored with error: %v", err)
			} else {
				console.Error("Failure: %v", err)
				return fmt.Errorf("step %s: %w", step.Name, err)
			}
		}

		if step.Delay > 0 {
			console.Info("Waiting %ds...", step.Delay)
			if !e.dryRun {
				select {
				case <-time.After(time.Duration(step.Delay) * time.Second):
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}

	return nil
}

// mergeVars 合并变量池：global -> host，host 优先。
func mergeVars(globalVars, hostVars resolve.Vars) resolve.Vars {
	vars := make(resolve.Vars, len(globalVars)+len(hostVars))
	maps.Copy(vars, globalVars)
	maps.Copy(vars, hostVars)
	return vars
}
