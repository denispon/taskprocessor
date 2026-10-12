package model

import (
	"fmt"
	"time"
)

type Task struct {
	ID      int
	Data    string
	Timeout time.Duration
}

type Result struct {
	TaskID int
	Output string
	Err    error
}

func (r Result) ShowResult() string {
	if r.Err != nil {
		return fmt.Sprintf("\n\n\tTaskID: %d, Output: %s, Error: %v", r.TaskID, r.Output, r.Err)
	}
	return fmt.Sprintf("\n\n\tTaskID: %d, Output: %s", r.TaskID, r.Output)
}
