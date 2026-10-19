package task

import (
	"context"
	"errors"
	"testing"

	"mestrap.com/taskssh/internal/action"
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/host"
	"mestrap.com/taskssh/internal/resolve"
)

type waitForCancellationAction struct {
	started chan struct{}
}

func (a *waitForCancellationAction) Execute(ctx *action.ActionContext) error {
	close(a.started)
	<-ctx.Context.Done()
	return ctx.Context.Err()
}

func TestRunParallelCancellationStopsDispatch(t *testing.T) {
	const actionName = "test-wait-for-cancellation"
	started := make(chan struct{})
	action.Register(actionName, &waitForCancellationAction{started: started})

	executor := &Executor{concurrency: 1, dryRun: true}
	taskConfig := &config.Task{Steps: []config.Step{{Action: actionName}}}
	hosts := []host.HostEntry{
		{Name: "host-1", Host: config.Host{Host: "host-1"}},
		{Name: "host-2", Host: config.Host{Host: "host-2"}},
		{Name: "host-3", Host: config.Host{Host: "host-3"}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan []Result, 1)
	go func() {
		done <- executor.runParallel(ctx, taskConfig, hosts, resolve.Vars{})
	}()

	<-started
	cancel()
	results := <-done
	if len(results) != 1 {
		t.Fatalf("runParallel returned %d results after cancellation, want only the dispatched host", len(results))
	}
	if results[0].Host != "host-1" || !errors.Is(results[0].Error, context.Canceled) {
		t.Fatalf("unexpected canceled result: %+v", results[0])
	}
}

func TestRunParallelAlreadyCanceledDoesNotDispatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	executor := &Executor{concurrency: 1, dryRun: true}
	taskConfig := &config.Task{}
	hosts := []host.HostEntry{{Name: "host-1", Host: config.Host{Host: "host-1"}}}

	results := executor.runParallel(ctx, taskConfig, hosts, resolve.Vars{})
	if len(results) != 0 {
		t.Fatalf("runParallel returned %d results, want 0", len(results))
	}
}
