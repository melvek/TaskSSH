package task

import (
	"context"
	"fmt"
	"maps"
	"time"

	"mestrap.com/taskssh/internal/action"
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/host"
	"mestrap.com/taskssh/internal/resolve"
	"mestrap.com/taskssh/internal/ssh"
)

// runOnHost 对单台主机执行整条任务。
//
// 变量合并顺序：
//
//	defaults -> 组 vars -> 主机 Extra -> 标准变量 -> CLI -D -> 运行时（date / exec_id）
func (e *Executor) runOnHost(ctx context.Context, task *config.Task, entry host.HostEntry,
	globalVars resolve.Vars, logger *console.Logger) error {

	// 合并清单变量：global -> 组 -> 主机
	vars := entry.BaseVars(globalVars)

	// 合并命令行变量
	maps.Copy(vars, e.cliVars)

	// 加载执行时参数
	vars["exec_id"] = e.execID

	var client *ssh.Client

	if !e.dryRun {
		c, err := ssh.ConnectContext(ctx, &entry.Host, e.connectTimeout)
		if err != nil {
			logger.Error("host=%s connect: %v", entry.Name, err)
			return err
		}
		client = c
		defer client.Close()
	}

	for i, step := range task.Steps {
		if err := ctx.Err(); err != nil {
			return err
		}

		act := action.Get(step.Action)
		if act == nil {
			logger.Error("host=%s step=%s: unknown action %s", entry.Name, step.Name, step.Action)
			return fmt.Errorf("unknown action: %s", step.Action)
		}

		if len(task.Steps) > 1 {
			logger.Step(i+1, len(task.Steps), step.Name)
		}

		if step.When != "" {
			ok, expanded, err := evalWhen(step.When, vars)
			if err != nil {
				logger.Error("host=%s step=%s eval when: %v", entry.Name, step.Name, err)
				return fmt.Errorf("step %s: eval when: %w", step.Name, err)
			}
			if !ok {
				logger.Skipped("ignore: %s", expanded)
				continue
			}
		}

		actx := &action.ActionContext{
			Context: ctx,
			Host:    &entry.Host,
			Vars:    vars,
			With:    step.With,
			Client:  client,
			DryRun:  e.dryRun,
			Logger:  logger,
		}

		err := withRetryLogger(ctx, logger, func() error {
			return act.Execute(actx)
		}, step.Interval, step.Timeout)

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err != nil {
			if step.IgnoreErrors {
				logger.Ignored("host=%s step=%s: %v", entry.Name, step.Name, err)
			} else {
				logger.Error("host=%s step=%s: %v", entry.Name, step.Name, err)
				return fmt.Errorf("step %s: %w", step.Name, err)
			}
		} else {
			logger.Success("Step completed: %s", step.Name)
		}

		if step.Delay > 0 {
			logger.Info("Waiting %ds...", step.Delay)
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
