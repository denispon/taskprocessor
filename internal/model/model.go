package model

import (
	"context"
	"database/sql"
	"fmt"
	"taskprocessor/internal/db"
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

func FindTaskByID(ctx context.Context, database *sql.DB, id int) (*Task, error) {
	taskDb, err := db.FetchTask(ctx, database, id)
	if err != nil {
		return nil, err
	}

	task := &Task{
		ID:      taskDb.ID,
		Data:    taskDb.Data,
		Timeout: time.Duration(taskDb.TimeoutNS) * time.Nanosecond,
	}

	return task, nil
}
