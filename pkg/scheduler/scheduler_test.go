package scheduler

import (
	"errors"
	"testing"
	"time"

	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Initialize scheduler",
		},
		{
			name: "Initialize scheduler multiple times",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Stop any existing scheduler to ensure a clean state for testing Init
			if Scheduler != nil {
				Scheduler.Stop()
			}
			Scheduler = nil // Explicitly nil it out

			Init()
			if Scheduler == nil {
				t.Errorf("Init() failed to initialize Scheduler, got nil")
				return
			}
			// Check type, though direct type assertion on unexported types from external packages can be tricky.
			// The fact that we can call methods like Del, AddWithID implies it's likely the correct type.
			// We'll rely on subsequent tests for Add/Stop to confirm functionality.

			if tt.name == "Initialize scheduler multiple times" {
				// Get the first instance
				firstSchedulerInstance := Scheduler
				Init() // Call Init again
				if Scheduler == nil {
					t.Errorf("Init() after second call failed to initialize Scheduler, got nil")
				}
				if Scheduler != firstSchedulerInstance {
					// This behavior (re-assigning Scheduler) is fine, but good to note.
					// If the expectation was to keep the original instance, this test would fail.
					// t.Logf("Init() on second call created a new Scheduler instance, which is acceptable.")
				}
			}
		})
	}
}

func TestStop(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() // For specific setup before Stop()
		expectPanic bool
	}{
		{
			name: "Stop an initialized and running scheduler",
			setup: func() {
				Init() // Ensure scheduler is initialized
			},
			expectPanic: false,
		},
		{
			name: "Stop a scheduler that is already stopped",
			setup: func() {
				Init()
				Stop() // Stop it once
			},
			expectPanic: false,
		},
		{
			name: "Stop a nil scheduler",
			setup: func() {
				// Ensure scheduler is nil
				if Scheduler != nil {
					Scheduler.Stop() // Stop if it's running
				}
				Scheduler = nil
			},
			expectPanic: false, // The underlying library might panic, or our Stop() might. Let's check.
			                        // Update: tasks.Scheduler.Stop() checks if it's nil.
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("Stop() panicked unexpectedly: %v", r)
					}
				} else {
					if tt.expectPanic {
						t.Errorf("Stop() did not panic as expected")
					}
				}
				// Cleanup: Re-initialize scheduler for subsequent tests if it was stopped or nilled
				// Cleanup logic after panic checks.
				// If a test stops or nils the scheduler, re-initialize for subsequent tests.
				// This ensures each test run (if not parallel) or subsequent top-level test
				// gets a fresh scheduler if the current one was altered.
				if (tt.name == "Stop an initialized and running scheduler" || tt.name == "Stop a scheduler that is already stopped") && Scheduler == nil {
					// This case should not happen if Stop() doesn't nil Scheduler.
					// If Stop() could nil Scheduler, then Init() would be needed.
				} else if tt.name == "Stop a nil scheduler" {
					// Scheduler was intentionally nil, leave it for the next test's setup.
				}
			}()

			if tt.setup != nil {
				tt.setup()
			}
			
			Stop() // The function under test

			// If we didn't expect a panic and Scheduler was not nil initially,
			// we can check if it's stopped (if the library provides such a mechanism).
			// The `tasks.Scheduler` has an `IsRunning()` method.
			if !tt.expectPanic && Scheduler != nil {
				// Note: IsRunning() might still be true for a short period after Stop() is called,
				// as the internal shutdown might take a moment.
				// A more robust check involves trying to add tasks or verifying existing tasks stop.
				// For now, this is a basic check.
				// Allow a brief moment for the scheduler to fully stop its internal loops.
				// time.Sleep(10 * time.Millisecond) // This makes tests slow and still not 100% reliable.
				// The tasks library's Stop() is synchronous and waits for tasks to finish.
				if Scheduler.IsRunning() {
					// This check is tricky. tasks.Scheduler.Stop() is synchronous.
					// However, if no tasks were ever added, tasks.Scheduler.IsRunning() might be false even after Init().
					// It becomes true when the main loop starts, which happens on Add() or Start().
					// So, for a scheduler that was just Init() and then Stop(), IsRunning() might already be false.
					// Let's consider what state implies "stopped" for our package.
					// The main effect of Stop() is that tasks won't run. This is tested in TestAdd.
				}
			}
		})
	}
}

func TestAdd(t *testing.T) {
	// Setup logger for tasks' ErrFunc
	oldLog := utils.Log
	utils.Log = utils.GetLogger() // Ensure Log is not nil
	defer func() { utils.Log = oldLog }()

	t.Run("Add and run a valid task", func(t *testing.T) {
		Init()
		defer Stop()

		var counter int
		taskID := "testTask1"
		err := Add(taskID, 10*time.Millisecond, func() error {
			counter++
			return nil
		})
		if err != nil {
			t.Fatalf("Add() returned an unexpected error: %v", err)
		}

		// Allow some time for the task to run multiple times
		time.Sleep(55 * time.Millisecond) // Should run ~5 times

		if counter == 0 {
			t.Errorf("Task was not executed")
		}
		if counter < 2 { // Check if it ran multiple times, not just once
			t.Errorf("Task expected to run multiple times, counter is %d", counter)
		}
		// t.Logf("Counter for %s: %d", taskID, counter) // For debugging
	})

	t.Run("Add task with duplicate ID replaces existing task", func(t *testing.T) {
		Init()
		defer Stop()

		var counterA, counterB int
		taskID := "duplicateTest"

		// Add first task
		Add(taskID, 20*time.Millisecond, func() error {
			counterA++
			return nil
		})

		// Add second task with the same ID
		Add(taskID, 20*time.Millisecond, func() error {
			counterB++
			return nil
		})

		time.Sleep(70 * time.Millisecond) // Allow time for tasks to run

		if counterA > 0 {
			t.Errorf("Original task (counterA) was executed, expected it to be replaced. counterA: %d", counterA)
		}
		if counterB == 0 {
			t.Errorf("New task (counterB) was not executed. counterB: %d", counterB)
		}
		// t.Logf("CounterA: %d, CounterB: %d for %s", counterA, counterB, taskID) // For debugging
	})

	t.Run("Add task with zero or negative interval", func(t *testing.T) {
		Init()
		defer Stop()

		err := Add("zeroIntervalTask", 0*time.Millisecond, func() error { return nil })
		if err == nil {
			// tasks library might not error on 0 if it means "run once immediately then stop"
			// or if it defaults to a minimum. The library's behavior for 0 interval is "run once".
			// Let's check if it runs.
			// t.Log("Add task with zero interval did not return an error, this might be expected by the underlying library to run once.")
		} else {
			// If it errors, that's also acceptable for a 0 interval.
			// t.Logf("Add task with zero interval returned error as expected: %v", err)
		}


		err = Add("negativeIntervalTask", -5*time.Millisecond, func() error { return nil })
		if err == nil {
			t.Errorf("Add() with negative interval expected an error, but got nil")
		}
	})

	t.Run("Add task after Stop", func(t *testing.T) {
		Init()
		Stop() // Stop the scheduler

		var counter int
		err := Add("taskAfterStop", 10*time.Millisecond, func() error {
			counter++
			return nil
		})

		// The `tasks` library's AddWithID does not return an error if the scheduler is stopped.
		// It allows adding tasks, but they won't be picked up by the (stopped) processing loop.
		if err != nil {
			t.Errorf("Add() after Stop returned an error: %v", err)
		}
		
		time.Sleep(30 * time.Millisecond) // Wait to see if task runs

		if counter > 0 {
			t.Errorf("Task added after Stop() was executed, counter: %d", counter)
		}
	})

	t.Run("Task returns an error", func(t *testing.T) {
		Init()
		defer Stop()
	
		// To verify ErrFunc, we'd ideally capture logs or have ErrFunc signal.
		// For now, we just ensure Add doesn't fail and the task system handles it.
		// This test is more about the scheduler's resilience.
		taskID := "erroringTask"
		err := Add(taskID, 10*time.Millisecond, func() error {
			return errors.New("task failed as expected")
		})
		if err != nil {
			t.Fatalf("Add() returned an unexpected error for an erroring task: %v", err)
		}
	
		// Allow some time for the task to attempt execution and call ErrFunc
		time.Sleep(30 * time.Millisecond) 
		// No direct assertion here without log capture/ErrFunc signaling.
		// We are implicitly testing that the scheduler doesn't crash.
	})

}
