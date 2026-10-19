package db

import (
	"context"
	"os"
	"testing"
	"time"

	"taskprocessor/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	os.Exit(testutil.RunWithTestDB(m))
}

func TestDBConnection(t *testing.T) {
	db, err := InitDB(testutil.TestDSN())
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("Failed to ping DB: %v", err)
	}
}

func TestFetchUserBalance(t *testing.T) {
	db, err := InitDB(testutil.TestDSN())
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	userID := 5
	expectedBalance := float64(userID * 100)

	balance, err := FetchUserBalance(ctx, db, userID)
	if err != nil {
		t.Fatalf("Failed to fetch user balance: %v", err)
	}

	if balance != expectedBalance {
		t.Errorf("Expected balance %f, got %f", expectedBalance, balance)
	}
}

func TestFetchTasks(t *testing.T) {
	db, err := InitDB(testutil.TestDSN())
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	taskId := 7

	task, err := FetchTask(ctx, db, taskId)
	if err != nil {
		t.Fatalf("Failed to fetch task: %v", err)
	}

	assert.Equal(t, taskId, task.ID, "Expected task ID %d, got %d", taskId, task.ID)

}
