package model

import (
	"context"
	"database/sql"
	"fmt"
	"taskprocessor/internal/db"
)

type User struct {
	ID      int
	Name    string
	Balance float64
}

func ShowUserInfo(ctx context.Context, database *sql.DB, userID int) (*User, error) {
	userDb, err := db.FetchUserInfo(ctx, database, userID)
	if err == nil {
		return &User{
			ID:      userDb.ID,
			Name:    userDb.Name,
			Balance: userDb.Balance,
		}, nil
	}
	return nil, err
}

func TransferFunds(ctx context.Context,
	database *sql.DB,
	fromUserID,
	toUserID int,
	amount float64,
	maxRetries int,
) error {
	err := db.ExecuteOperationWithRetry(ctx, database, maxRetries, func(tx *sql.Tx) error {
		fromUser, err := db.FetchUserInfo(ctx, tx, fromUserID)
		if err != nil {
			return err
		}
		toUser, err := db.FetchUserInfo(ctx, tx, toUserID)
		if err != nil {
			return err
		}
		if fromUser.Balance < amount {
			return fmt.Errorf("insufficient funds for user %d", fromUserID)
		}

		fromUser.Balance -= amount
		toUser.Balance += amount
		if err := db.UpdateUserBalance(ctx, tx, fromUser); err != nil {
			return err
		}
		return db.UpdateUserBalance(ctx, tx, toUser)
	})
	if err != nil {
		return err
	}
	fmt.Printf("Transferred %.2f from user %d to user %d\n", amount, fromUserID, toUserID)
	return nil
}
