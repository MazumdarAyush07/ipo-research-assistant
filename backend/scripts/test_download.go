//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/MazumdarAyush07/ipo-research/internal/worker"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("No .env file found")
	}

	redisOpt := asynq.RedisClientOpt{Addr: "localhost:6379"}
	client := asynq.NewClient(redisOpt)
	defer client.Close()

	payload, _ := json.Marshal(worker.DownloadDocumentsPayload{IPOID: 1})
	task := asynq.NewTask(worker.TaskDownloadDocuments, payload)

	info, err := client.EnqueueContext(context.Background(), task)
	if err != nil {
		log.Fatalf("Could not enqueue task: %v", err)
	}
	log.Printf("Successfully enqueued task: id=%s queue=%s", info.ID, info.Queue)
}
