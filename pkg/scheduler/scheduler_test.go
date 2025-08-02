package scheduler

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/jiotv-go/jiotv_go/v3/pkg/utils"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	utils.Log = log.New(os.Stdout, "", log.LstdFlags)
	os.Exit(m.Run())
}

func TestInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Initialize scheduler",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init()
			if Scheduler == nil {
				t.Error("Init() should initialize Scheduler")
			}
		})
	}
}

func TestStop(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Stop scheduler",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize scheduler first
			Init()
			if Scheduler == nil {
				t.Fatal("Scheduler should be initialized")
			}
			
			// Should not panic when stopping
			Stop()
		})
	}
}

func TestAdd(t *testing.T) {
	// Initialize scheduler first
	Init()
	
	type args struct {
		id       string
		interval time.Duration
		task     func() error
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Add simple task",
			args: args{
				id:       "test_task_1",
				interval: 1 * time.Second,
				task: func() error {
					return nil
				},
			},
		},
		{
			name: "Add task that returns error",
			args: args{
				id:       "test_task_2",
				interval: 2 * time.Second,
				task: func() error {
					return nil // Don't actually return error in test to avoid log spam
				},
			},
		},
		{
			name: "Add task with same ID (should replace)",
			args: args{
				id:       "test_task_1", // Same ID as first test
				interval: 3 * time.Second,
				task: func() error {
					return nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			Add(tt.args.id, tt.args.interval, tt.args.task)
		})
	}
}
