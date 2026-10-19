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

func (w *Worker) ExecuteTask(ctx context.Context, task *Task, results chan *Result) error {

	cancel := context.CancelFunc(func() {})
	if task.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, task.Timeout)

	}
	defer cancel()

	result := &Result{
		TaskID: task.ID,
		Output: fmt.Sprintf("%s performed task %d, data: %s", w.Name, task.ID, task.Data),
	}

	seconds := rand.Intn(3) + 1
	sleepDuration := time.Duration(seconds) * time.Second

	select {
	case <-ctx.Done():
		result.Err = fmt.Errorf("task %d was cancelled", task.ID)
		results <- result
		return result.Err
	case <-time.After(sleepDuration):
		results <- result
	}

	return nil

}
