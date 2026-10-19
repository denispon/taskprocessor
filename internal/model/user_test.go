package model

import (
	"context"
	"fmt"
	"testing"

	"taskprocessor/internal/db"

	"github.com/DATA-DOG/go-sqlmock"
	pgerror "github.com/lib/pq"
)

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		code string
		want bool
	}{
		{name: "serialization failure", code: "40001", want: true},
		{name: "deadlock detected", code: "40P01", want: true},
		{name: "other PostgreSQL error", code: "23505", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgErr := &pgerror.Error{Code: pgerror.ErrorCode(tt.code)}
			err := fmt.Errorf("wrapped: %w", pgErr)

			if got := db.IsRetryableError(err); got != tt.want {
				t.Errorf("IsRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}

}

func TestTransferFunds(t *testing.T) {
	database, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("create database mock: %v", err)
	}
	defer database.Close()

	ctx := context.Background()
	fromUserID, toUserID := 1, 2
	amount := 25.0
	selectUser := "SELECT id, name, balance FROM users WHERE id = $1"
	updateBalance := "UPDATE users SET balance = $1 WHERE id = $2"

	mock.ExpectBegin()
	mock.ExpectQuery(selectUser).WithArgs(fromUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance"}).AddRow(fromUserID, "Sender", "100.00"))
	mock.ExpectQuery(selectUser).WithArgs(toUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance"}).AddRow(toUserID, "Recipient", "50.00"))
	mock.ExpectExec(updateBalance).WithArgs(75.0, fromUserID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(updateBalance).WithArgs(75.0, toUserID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := TransferFunds(ctx, database, fromUserID, toUserID, amount, 1); err != nil {
		t.Fatalf("TransferFunds() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet database expectations: %v", err)
	}
}

func TestTransferFundsRetriesSerializationFailure(t *testing.T) {
	database, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("create database mock: %v", err)
	}
	defer database.Close()

	ctx := context.Background()
	fromUserID, toUserID := 1, 2
	selectUser := "SELECT id, name, balance FROM users WHERE id = $1"
	updateBalance := "UPDATE users SET balance = $1 WHERE id = $2"
	addTransfer := func() {
		mock.ExpectBegin()
		mock.ExpectQuery(selectUser).WithArgs(fromUserID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance"}).AddRow(fromUserID, "Sender", "100.00"))
		mock.ExpectQuery(selectUser).WithArgs(toUserID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance"}).AddRow(toUserID, "Recipient", "50.00"))
		mock.ExpectExec(updateBalance).WithArgs(75.0, fromUserID).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(updateBalance).WithArgs(75.0, toUserID).WillReturnResult(sqlmock.NewResult(0, 1))
	}

	addTransfer()
	mock.ExpectCommit().WillReturnError(&pgerror.Error{Code: pgerror.ErrorCode("40001")})
	addTransfer()
	mock.ExpectCommit()

	if err := TransferFunds(ctx, database, fromUserID, toUserID, 25.0, 2); err != nil {
		t.Fatalf("TransferFunds() error after retry = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet database expectations: %v", err)
	}
}

func TestTransferFundsReturnsErrorAfterMaxRetries(t *testing.T) {
	database, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("create database mock: %v", err)
	}
	defer database.Close()

	ctx := context.Background()
	fromUserID, toUserID := 1, 2
	maxRetries := 2
	selectUser := "SELECT id, name, balance FROM users WHERE id = $1"
	updateBalance := "UPDATE users SET balance = $1 WHERE id = $2"
	retryErr := &pgerror.Error{Code: pgerror.ErrorCode("40001")}

	for attempt := 0; attempt < maxRetries; attempt++ {
		mock.ExpectBegin()
		mock.ExpectQuery(selectUser).WithArgs(fromUserID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance"}).AddRow(fromUserID, "Sender", "100.00"))
		mock.ExpectQuery(selectUser).WithArgs(toUserID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance"}).AddRow(toUserID, "Recipient", "50.00"))
		mock.ExpectExec(updateBalance).WithArgs(75.0, fromUserID).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(updateBalance).WithArgs(75.0, toUserID).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit().WillReturnError(retryErr)
	}

	gotErr := TransferFunds(ctx, database, fromUserID, toUserID, 25.0, maxRetries)
	if gotErr != retryErr {
		t.Fatalf("TransferFunds() error = %v, want final retryable error %v", gotErr, retryErr)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet database expectations: %v", err)
	}
}

func TestTransferFundsInsufficientFunds(t *testing.T) {
	database, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("create database mock: %v", err)
	}
	defer database.Close()

	ctx := context.Background()
	fromUserID, toUserID := 1, 2
	selectUser := "SELECT id, name, balance FROM users WHERE id = $1"

	mock.ExpectBegin()
	mock.ExpectQuery(selectUser).WithArgs(fromUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance"}).AddRow(fromUserID, "Sender", "10.00"))
	mock.ExpectQuery(selectUser).WithArgs(toUserID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "balance"}).AddRow(toUserID, "Recipient", "50.00"))
	mock.ExpectRollback()

	err = TransferFunds(ctx, database, fromUserID, toUserID, 25.0, 1)
	if err == nil || err.Error() != "insufficient funds for user 1" {
		t.Fatalf("TransferFunds() error = %v, want insufficient funds error", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet database expectations: %v", err)
	}
}
