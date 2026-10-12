package model

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type Worker struct {
	ID   int
	Name string
}

func (w *Worker) ExecuteTask(task *Task, results chan *Result) {

	ctx := context.Background()
	cancel := context.CancelFunc(func() {})
	if task.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), task.Timeout)

	}
	defer cancel()

	result := &Result{
		TaskID: task.ID,
		Output: fmt.Sprintf("Worker %d(%s) performed task %d", w.ID, w.Name, task.ID),
	}

	seconds := rand.Intn(3) + 1
	sleepDuration := time.Duration(seconds) * time.Second

	select {
	case <-ctx.Done():
		result.Err = fmt.Errorf("task %d was cancelled", task.ID)
		results <- result
	case <-time.After(sleepDuration):
		results <- result
	}

}
