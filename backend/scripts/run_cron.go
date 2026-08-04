//go:build ignore

package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/MazumdarAyush07/ipo-research/internal/worker"
	"github.com/hibiken/asynq"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("No .env file found or error loading it, relying on system environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx := context.Background()
	
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	queries := models.New(db)
	
	// Connect to local Redis for Asynq (default addr)
	redisOpt := asynq.RedisClientOpt{Addr: "localhost:6379"}
	client := asynq.NewClient(redisOpt)
	defer client.Close()

	proc := worker.NewProcessor(queries, client)

	log.Println("Manually triggering IPO Sync Task...")
	dummyTask := asynq.NewTask(worker.TaskSyncIPOs, nil)
	err = proc.HandleSyncIPOsTask(ctx, dummyTask)
	if err != nil {
		log.Fatalf("Failed to sync IPOs: %v", err)
	}
	
	log.Println("Cron job executed successfully.")
	os.Exit(0)
}
