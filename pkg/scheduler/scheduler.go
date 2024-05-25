package scheduler

import (
	"time"

	"github.com/madflojo/tasks"
	"github.com/rabilrbl/jiotv_go/v3/pkg/utils"
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
			utils.Log.Printf("Task failed: %v\n", err)
		},
		StartAfter: interval, // Convert interval to time.Time value
	})
	if err != nil {
		utils.Log.Printf("Failed to add task: %v\n", err)
		return
	}
	utils.Log.Printf("Task added with ID: %v\n", id)
}

