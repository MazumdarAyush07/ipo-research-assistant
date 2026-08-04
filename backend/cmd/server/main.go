package main

import (
	"log"

	"github.com/MazumdarAyush07/ipo-research/internal/api"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Setup Phase 3 API handlers
	ipoHandler := api.NewIPOHandler(nil)
	app.Get("/api/ipos", ipoHandler.ListIPOs)
	app.Post("/api/ipos", ipoHandler.ManualSyncIPOs)

	log.Fatal(app.Listen(":8080"))
}
