//go:build ignore

package main

import (
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
	_ = godotenv.Load("../.env")

	// Connect to Database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()
	queries := models.New(db)

	// Setup Asynq Worker Server
	redisOpt := asynq.RedisClientOpt{Addr: "localhost:6379"}
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
		},
	)

	// Register Handlers
	mux := asynq.NewServeMux()
	proc := worker.NewProcessor(queries, nil) // AsynqClient is nil for the worker itself since it doesn't enqueue tasks

	mux.HandleFunc(worker.TaskDownloadDocuments, proc.HandleDownloadDocumentsTask)

	log.Println("Starting Asynq Worker Server...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
