package main

import (
	"log"
	"os"

	"github.com/MazumdarAyush07/ipo-research/internal/api"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("../.env")

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = os.Getenv("REDIS_ADDR")
	}
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer asynqClient.Close()

	// Setup API handlers
	ipoHandler := api.NewIPOHandler(nil, asynqClient)
	app.Get("/api/ipos", ipoHandler.ListIPOs)
	app.Post("/api/ipos", ipoHandler.ManualSyncIPOs)
	app.Post("/api/ipos/:id/documents/trigger", ipoHandler.TriggerDocumentDownload)

	log.Fatal(app.Listen(":8080"))
}
