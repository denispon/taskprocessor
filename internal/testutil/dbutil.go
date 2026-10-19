package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	defaultAdminDSN = "postgres://postgres:secret@localhost:5432/postgres?sslmode=disable"
	defaultTestDSN  = "postgres://postgres:secret@localhost:5432/taskprocessor_test?sslmode=disable"
)

// TestDSN returns the connection string used by the integration tests.
func TestDSN() string {
	if dsn := os.Getenv("TASKPROCESSOR_TEST_DSN"); dsn != "" {
		return dsn
	}
	return defaultTestDSN
}

// RunWithTestDB prepares the shared test fixtures, runs the package tests,
// and removes the fixtures before returning the test exit code.
func RunWithTestDB(m *testing.M) int {
	if os.Getenv("TASKPROCESSOR_TEST_DSN") == "" {
		if err := ensureTestDatabase(); err != nil {
			fmt.Fprintf(os.Stderr, "create test database: %v\n", err)
			return 1
		}
	}

	db, err := sql.Open("postgres", TestDSN())
	if err != nil {
		fmt.Fprintf(os.Stderr, "open test database: %v\n", err)
		return 1
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "connect to test database: %v\n", err)
		return 1
	}

	if err := resetTestData(db); err != nil {
		fmt.Fprintf(os.Stderr, "prepare test database: %v\n", err)
		return 1
	}

	code := m.Run()
	if err := dropTestTables(db); err != nil {
		fmt.Fprintf(os.Stderr, "clean up test database: %v\n", err)
		return 1
	}
	return code
}

func ensureTestDatabase() error {
	db, err := sql.Open("postgres", defaultAdminDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return err
	}

	var exists bool
	if err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", "taskprocessor_test").Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = db.Exec("CREATE DATABASE taskprocessor_test")
	return err
}

func resetTestData(db *sql.DB) error {
	if err := dropTestTables(db); err != nil {
		return err
	}

	statements := []string{
		`CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			balance NUMERIC(10,2) NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE tasks (
			id SERIAL PRIMARY KEY,
			data TEXT NOT NULL,
			timeout_ns BIGINT NOT NULL DEFAULT 0
		)`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}

	for i := 1; i <= 10; i++ {
		if _, err := db.Exec("INSERT INTO users (id, name, balance) VALUES ($1, $2, $3)", i, fmt.Sprintf("Test User %d", i), float64(i*100)); err != nil {
			return err
		}
		if _, err := db.Exec("INSERT INTO tasks (id, data, timeout_ns) VALUES ($1, $2, $3)", i, fmt.Sprintf("test-task-%d", i), int64(i*1_000_000)); err != nil {
			return err
		}
	}
	return nil
}

func dropTestTables(db *sql.DB) error {
	if _, err := db.Exec("DROP TABLE IF EXISTS tasks"); err != nil {
		return err
	}
	_, err := db.Exec("DROP TABLE IF EXISTS users")
	return err
}
