package config

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

var DB *pgx.Conn

func ConnectDB() {
	connStr := os.Getenv("DATABASE_URL")

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		os.Exit(1)
	}

	DB = conn
	fmt.Println("Successfully connected to the database!")
}
