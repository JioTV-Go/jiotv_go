package scheduler

import (
	"fmt"
	"time"

	"github.com/madflojo/tasks"
)

var (
	// Scheduler is the task scheduler
	Scheduler *tasks.Scheduler
)

func Init() {
	// Create a new task scheduler
	Scheduler = tasks.New()
}

func Stop() {
	// Stop the task scheduler
	Scheduler.Stop()
}

func Add(interval time.Time, task func() error) {
	// Add a task
	id, err := Scheduler.Add(&tasks.Task{
		Interval: time.Until(interval),
		TaskFunc: task,
		ErrFunc: func(err error) {
			fmt.Printf("Task failed: %v\n", err)
		},
		StartAfter: interval, // Convert interval to time.Time value
	})
	if err != nil {
		fmt.Printf("Failed to add task: %v\n", err)
		return
	}
	fmt.Printf("Task added with ID: %v\n", id)
}

