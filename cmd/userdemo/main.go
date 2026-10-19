package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"taskprocessor/internal/db"
	"taskprocessor/internal/model"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	database, err := db.InitDB("postgres://postgres:secret@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		fmt.Printf("Failed to initialize DB: %v\n", err)
		return
	}
	defer database.Close()

	printUsers(ctx, database)

	if err := model.TransferFunds(ctx, database, 4, 9, 10, 3); err != nil {
		fmt.Printf("Failed to transfer funds: %v\n", err)
		return
	}

	printUsers(ctx, database)
}

func printUsers(ctx context.Context, database *sql.DB) {
	for i := 1; i <= 10; i++ {
		userInfo, err := model.ShowUserInfo(ctx, database, i)
		if err != nil {
			fmt.Printf("Error fetching user info for ID %d: %v\n", i, err)
			continue
		}
		fmt.Printf("User ID: %d, Name: %s, Balance: %.2f\n", userInfo.ID, userInfo.Name, userInfo.Balance)
	}
}
