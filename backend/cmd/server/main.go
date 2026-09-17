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
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/hibiken/asynq"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("../.env")

	app := fiber.New()
	
	// Default CORS allows all origins
	app.Use(cors.New())

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

	asynqInspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: redisAddr})
	defer asynqInspector.Close()

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
	ipoHandler := api.NewIPOHandler(queries, asynqClient, asynqInspector, peerService)
	// Register Routes
	app.Get("/api/ipos", ipoHandler.ListIPOs)
	app.Post("/api/ipos", ipoHandler.ManualSyncIPOs)
	
	app.Get("/api/ipos/:id/financials", ipoHandler.GetFinancials)
	app.Get("/api/ipos/:id/analysis", ipoHandler.GetAIAnalysis)
	app.Post("/api/ipos/:id/analysis/trigger", ipoHandler.TriggerAIAnalysis)
	app.Get("/api/ipos/:id/peers", ipoHandler.GetPeers)
	app.Get("/api/ipos/:id/gmp", ipoHandler.GetGMP)
	app.Get("/api/ipos/:id/subscriptions", ipoHandler.GetSubscriptions)
	app.Post("/api/trackers/sync", ipoHandler.TriggerTrackers)
	
	app.Get("/api/ipos/:id/score", ipoHandler.GetScore)
	app.Post("/api/ipos/:id/score/trigger", ipoHandler.TriggerScore)

	// Report endpoints
	app.Post("/api/ipos/:id/report/generate", ipoHandler.TriggerReportGeneration)
	app.Get("/api/ipos/:id/report", ipoHandler.GetReport)

	// Tracker endpoints
	app.Post("/api/admin/trackers/trigger", ipoHandler.TriggerTrackers)
	app.Post("/api/admin/trackers/peers/trigger", ipoHandler.TriggerSyncPeers)

	// Document endpoints
	app.Post("/api/ipos/:id/documents/trigger", ipoHandler.TriggerDocumentDownload)
	app.Post("/api/ipos/:id/parse/trigger", ipoHandler.TriggerDocumentParse)

	// Admin Audit endpoints
	app.Get("/api/admin/audit", ipoHandler.GetAudit)
	app.Get("/api/admin/parsing-audit", ipoHandler.GetParsingAudit)
	app.Get("/api/admin/analysis-audit", ipoHandler.GetAnalysisAudit)
	app.Get("/api/admin/tracker-audit", ipoHandler.GetTrackerAudit)
	app.Get("/api/admin/scoring-audit", ipoHandler.GetScoringAudit)
	app.Get("/api/admin/report-audit", ipoHandler.GetReportAudit)
	app.Get("/api/admin/queues", ipoHandler.GetQueueStats)

	log.Fatal(app.Listen(":8080"))
}
