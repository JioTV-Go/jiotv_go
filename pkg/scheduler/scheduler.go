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

func Add(id string, interval time.Duration, task func() error) {
	// delete any existing task with the same ID
	Scheduler.Del(id)
	// Add a task
	err := Scheduler.AddWithID(id, &tasks.Task{
		Interval: interval,
		TaskFunc: task,
		ErrFunc: func(err error) {
			utils.Log.Printf("Task failed: %v\n", err)
		},
	})
	if err != nil {
		utils.Log.Printf("Failed to add task: %v\n", err)
		return
	}
	utils.Log.Printf("Task added with ID: %v\n", id)
}

