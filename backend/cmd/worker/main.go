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
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	
	// Create an Asynq client so the worker can enqueue child tasks
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer client.Close()

	redisOpt := asynq.RedisClientOpt{Addr: redisAddr}
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			// Concurrency is set to 1 to avoid overwhelming the pdf-parser (OOM crashes)
			Concurrency: 1,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	// Register Handlers
	mux := asynq.NewServeMux()
	proc := worker.NewProcessor(queries, client)

	mux.HandleFunc(worker.TaskSyncIPOs, proc.HandleSyncIPOsTask)
	mux.HandleFunc(worker.TaskDownloadDocuments, proc.HandleDownloadDocumentsTask)
	mux.HandleFunc(worker.TaskParseDocument, proc.ProcessTaskParseDocument)
	mux.HandleFunc(worker.TaskSyncGMP, proc.HandleSyncGMPTask)
	mux.HandleFunc(worker.TaskSyncSubscriptions, proc.HandleSyncSubscriptionsTask)
	mux.HandleFunc(worker.TaskSyncValuation, proc.HandleSyncValuationTask)
	mux.HandleFunc(worker.TaskAnalyzeDocument, proc.ProcessTaskAnalyzeDocument)
	mux.HandleFunc(worker.TaskGenerateReport, proc.HandleGenerateReportTask)

	log.Println("Starting Asynq Worker Server...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
