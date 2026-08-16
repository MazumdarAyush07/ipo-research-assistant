package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/MazumdarAyush07/ipo-research/internal/api"
	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/MazumdarAyush07/ipo-research/internal/services"
	"github.com/redis/go-redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
	_ "github.com/jackc/pgx/v5/stdlib"
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

	// DB Connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	queries := models.New(db)

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	peerService, err := services.NewPeerService(redisClient, "config/peers.json")
	if err != nil {
		log.Printf("Warning: Failed to init PeerService: %v", err)
	}

	// Setup API handlers
	ipoHandler := api.NewIPOHandler(queries, asynqClient, peerService)
	// Register Routes
	app.Get("/api/ipos", ipoHandler.ListIPOs)
	app.Post("/api/ipos", ipoHandler.ManualSyncIPOs)
	
	app.Get("/api/ipos/:id/financials", ipoHandler.GetFinancials)
	app.Get("/api/ipos/:id/analysis", ipoHandler.GetAIAnalysis)
	app.Get("/api/ipos/:id/peers", ipoHandler.GetPeers)
	app.Get("/api/ipos/:id/gmp", ipoHandler.GetGMP)
	app.Get("/api/ipos/:id/subscriptions", ipoHandler.GetSubscriptions)
	app.Post("/api/trackers/sync", ipoHandler.TriggerTrackers)

	// Document endpoints
	app.Post("/api/ipos/:id/documents/trigger", ipoHandler.TriggerDocumentDownload)
	app.Post("/api/ipos/:id/parse/trigger", ipoHandler.TriggerDocumentParse)

	log.Fatal(app.Listen(":8080"))
}
