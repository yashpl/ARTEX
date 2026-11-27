package server

import (
	"context"
	"testing"
	"time"
)

func waitForActiveWorkers(t *testing.T, e *Engine, taskID string, want int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if got := e.ActiveWorkers(taskID); got == want {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("task %s active workers = %d, want %d after %v", taskID, e.ActiveWorkers(taskID), want, timeout)
}

func TestDynamicWorkerScalingUpDownAndBounce(t *testing.T) {
	m := &Manager{tasks: map[string]*Task{}}
	e := NewEngine(m)

	if err := m.SetWorkers(2); err != nil {
		t.Fatalf("SetWorkers(2): %v", err)
	}
	if got := m.Workers(); got != 2 {
		t.Fatalf("Workers() = %d, want 2", got)
	}

	task := &Task{
		ID:     "task-scale-1",
		notify: make(chan struct{}, 1),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e.Run(ctx, task)
	waitForActiveWorkers(t, e, task.ID, 2, 2*time.Second)

	// Scale UP: 2 -> 5. New workers spawn immediately.
	if err := m.SetWorkers(5); err != nil {
		t.Fatalf("SetWorkers(5): %v", err)
	}
	waitForActiveWorkers(t, e, task.ID, 5, 2*time.Second)

	// Scale DOWN: 5 -> 2. Idle excess workers exit cleanly at top of workerLoop.
	if err := m.SetWorkers(2); err != nil {
		t.Fatalf("SetWorkers(2): %v", err)
	}
	waitForActiveWorkers(t, e, task.ID, 2, 4*time.Second)

	// Bounce: 2 -> 4 -> 1 -> 3 rapidly to exercise generation tokens and slot tracking.
	if err := m.SetWorkers(4); err != nil {
		t.Fatalf("SetWorkers(4): %v", err)
	}
	waitForActiveWorkers(t, e, task.ID, 4, 2*time.Second)

	if err := m.SetWorkers(1); err != nil {
		t.Fatalf("SetWorkers(1): %v", err)
	}
	if err := m.SetWorkers(3); err != nil {
		t.Fatalf("SetWorkers(3): %v", err)
	}
	waitForActiveWorkers(t, e, task.ID, 3, 4*time.Second)

	// Verify exact slot set 1..3 is live.
	e.runtimeMu.Lock()
	rt := e.runtimes[task.ID]
	if rt == nil {
		e.runtimeMu.Unlock()
		t.Fatalf("expected runtime for %s", task.ID)
	}
	for idx := 1; idx <= 3; idx++ {
		if _, ok := rt.workers[idx]; !ok {
			e.runtimeMu.Unlock()
			t.Fatalf("missing expected live worker slot %d (got %v)", idx, rt.workers)
		}
	}
	e.runtimeMu.Unlock()

	// Shutdown via StopTask must drain cleanly with balanced WaitGroup.
	done := make(chan struct{})
	go func() {
		e.StopTask(task.ID)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("StopTask timed out (WaitGroup leak or deadlock)")
	}
	if got := e.ActiveWorkers(task.ID); got != 0 {
		t.Fatalf("ActiveWorkers after StopTask = %d, want 0", got)
	}

	// Engine.Wait must also complete cleanly.
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer waitCancel()
	if err := e.Wait(waitCtx); err != nil {
		t.Fatalf("Engine.Wait: %v", err)
	}
}

func TestReconcileWorkersNudgesTaskOnScaleUpOnly(t *testing.T) {
	m := &Manager{tasks: map[string]*Task{}}
	e := NewEngine(m)

	task := &Task{
		ID:     "task-nudge-1",
		notify: make(chan struct{}, 1),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Seed a runtime with 2 live workers and no plannerLoop so task.notify is
	// observed deterministically without racing a planner consumer.
	e.runtimeMu.Lock()
	rtCtx, rtCancel := context.WithCancelCause(ctx)
	rt := &taskRuntime{
		ctx:     rtCtx,
		cancel:  rtCancel,
		task:    task,
		workers: map[int]uint64{1: 1, 2: 2},
	}
	e.desiredWorkers = 2
	e.workerGen = 2
	e.runtimes[task.ID] = rt
	e.runtimeMu.Unlock()

	// Scale UP 2 -> 4 must spawn slots 3..4 and nudge the task.
	if err := m.SetWorkers(4); err != nil {
		t.Fatalf("SetWorkers(4): %v", err)
	}
	select {
	case <-task.notify:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("expected task.Notify() nudge on scale-up")
	}

	// Scale DOWN 4 -> 2 must NOT spuriously nudge the task.
	if err := m.SetWorkers(2); err != nil {
		t.Fatalf("SetWorkers(2): %v", err)
	}
	select {
	case <-task.notify:
		t.Fatalf("unexpected task.Notify() on scale-down")
	default:
	}

	// Clean up spawned worker goroutines.
	e.StopTask(task.ID)
}

func TestSetWorkersRejectsNonPositive(t *testing.T) {
	m := &Manager{tasks: map[string]*Task{}}
	if err := m.SetWorkers(0); err == nil {
		t.Fatalf("expected error for SetWorkers(0)")
	}
	if err := m.SetWorkers(-2); err == nil {
		t.Fatalf("expected error for SetWorkers(-2)")
	}
}
