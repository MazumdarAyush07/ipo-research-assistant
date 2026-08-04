//go:build ignore

package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env in the parent directory
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("No .env file found or error loading it, relying on system environment variables")
	}

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// Connect to Neon DB
	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Clear the ipos table (CASCADE will also clear financials, peers, etc. linked to it)
	_, err = pool.Exec(context.Background(), "TRUNCATE ipos CASCADE;")
	if err != nil {
		log.Fatalf("Failed to truncate table: %v", err)
	}

	log.Println("✅ Successfully cleared all IPOs from the database!")
}
