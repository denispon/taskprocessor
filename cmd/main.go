package main

import (
	"context"
	"fmt"
	"time"

	"taskprocessor/internal/db"
	"taskprocessor/internal/model"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

			

	database, err := db.InitDB("postgres://postgres:secret@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		fmt.Printf("Failed to initialize DB: %v\n", err)
		return
	}

	defer database.Close()

	jobs := make(chan *model.Task, 10)
	results := make(chan *model.Result, 10)

	worker1 := model.Worker{ID: 1, Name: "worker1"}
	worker2 := model.Worker{ID: 2, Name: "worker2"}
	worker3 := model.Worker{ID: 3, Name: "worker3"}

	workers := []*model.Worker{&worker1, &worker2, &worker3}
	g, gCtx := errgroup.WithContext(ctx)

	for _, worker := range workers {
		g.Go(func() error {
			for task := range jobs {
				worker.ExecuteTask(gCtx, task, results)
			}
			return nil
		})
	}

	for i := 1; i <= 10; i++ {
		t, err := model.FindTaskByID(ctx, database, i)
		if err != nil {
			fmt.Printf("Error finding task by ID %d: %v\n", i, err)
			continue
		}

		jobs <- t
	}

	close(jobs)
	if err := g.Wait(); err != nil {
		fmt.Printf("Worker error: %v\n", err)
	}

	close(results)

	for result := range results {
		fmt.Println(result.ShowResult())
	}

}
