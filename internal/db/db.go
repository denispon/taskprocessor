package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq" // PostgreSQL driver
)

type TaskDb struct {
	ID        int
	Data      string
	TimeoutNS int64 // Timeout in nanoseconds
}

type UserDb struct {
	ID      int
	Name    string
	Balance float64
}

type queryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

// 1. Setup the connection pool
func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	// TODO: Configure your connection pool parameters here
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection is alive
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

// 2. Implement the context-aware query function
func FetchUserInfo(ctx context.Context, db queryRower, userID int) (*UserDb, error) {
	user := &UserDb{}
	query := "SELECT id, name, balance FROM users WHERE id = $1"

	// Hint: Use db.QueryRowContext to respect the context timeout
	err := db.QueryRowContext(ctx, query, userID).Scan(&user.ID, &user.Name, &user.Balance)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// 2. Implement the context-aware query function
func FetchTask(ctx context.Context, db *sql.DB, taskId int) (TaskDb, error) {
	task := TaskDb{}
	query := "SELECT id, data, timeout_ns FROM tasks WHERE id = $1"

	// Hint: Use db.QueryRowContext to respect the context timeout
	err := db.QueryRowContext(ctx, query, taskId).Scan(&task.ID, &task.Data, &task.TimeoutNS)
	if err != nil {
		return task, err
	}

	return task, nil
}

func UpdateUserBalance(ctx context.Context, database execer, user *UserDb) error {
	query := "UPDATE users SET balance = $1 WHERE id = $2"
	_, err := database.ExecContext(ctx, query, user.Balance, user.ID)

	return err
}

func IsRetryableError(err error) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return pgErr.Code == "40001" || pgErr.Code == "40P01"
	}

	return false

}

func ExecuteOperationWithRetry(
	ctx context.Context,
	database *sql.DB,
	maxRetries int,
	operation func(*sql.Tx) error,
) error {
	if maxRetries <= 0 {
		return fmt.Errorf("maxRetries must be greater than zero")
	}
	if operation == nil {
		return fmt.Errorf("operation must not be nil")
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		tx, err := database.BeginTx(ctx, nil)
		if err != nil {
			if IsRetryableError(err) && attempt+1 < maxRetries {
				continue
			}
			return err
		}

		if err := operation(tx); err != nil {
			_ = tx.Rollback()
			if IsRetryableError(err) && attempt+1 < maxRetries {
				continue
			}
			return err
		}

		if err := tx.Commit(); err != nil {
			if IsRetryableError(err) && attempt+1 < maxRetries {
				continue
			}
			return err
		}
		return nil
	}

	return fmt.Errorf("operation failed after %d attempts", maxRetries)
}
