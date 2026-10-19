package main

import (
	"context"
	"fmt"
	"time"
)

func ProcessSubTask(ctx context.Context, subTaskID int) error {

	workDone := make(chan struct{})
	go func() {
		time.Sleep(1 * time.Second) // Simulate work
		close(workDone)
	}()

	select {

	case <-ctx.Done():
		return fmt.Errorf("subtask %d processing cancelled: %v", subTaskID, ctx.Err())

	case <-workDone:
		fmt.Printf("Subtask %d processed successfully\n", subTaskID)
		return nil

	}
}
