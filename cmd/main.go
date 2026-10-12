package main

import (
	"fmt"
	"sync"
	"time"

	"taskprocessor/internal/model"
)

func main() {

	jobs := make(chan model.Task, 10)
	results := make(chan *model.Result, 10)

	worker1 := model.Worker{ID: 1, Name: "worker1"}
	worker2 := model.Worker{ID: 2, Name: "worker2"}
	worker3 := model.Worker{ID: 3, Name: "worker3"}

	workers := []*model.Worker{&worker1, &worker2, &worker3}

	var wg sync.WaitGroup

	for _, worker := range workers {
		wg.Add(1)
		go func(w *model.Worker) {
			defer wg.Done()
			for task := range jobs {
				w.ExecuteTask(&task, results)
			}
		}(worker)
	}

	for i := 1; i <= 10; i++ {
		jobs <- model.Task{
			ID:      i,
			Data:    fmt.Sprintf("Task data %d", i),
			Timeout: time.Duration(1) * time.Second,
		}
	}

	close(jobs)

	wg.Wait()

	close(results)

	for result := range results {
		fmt.Println(result.ShowResult())
	}
}
