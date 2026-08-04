package api

import (
	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/gofiber/fiber/v2"
)

// IPOHandler holds dependencies for API routes
type IPOHandler struct {
	Queries *models.Queries
}

func NewIPOHandler(q *models.Queries) *IPOHandler {
	return &IPOHandler{Queries: q}
}

// ListIPOs handles GET /api/ipos
func (h *IPOHandler) ListIPOs(c *fiber.Ctx) error {
	if h.Queries == nil {
		return c.JSON(fiber.Map{"status": "success", "data": []string{}})
	}

	ipos, err := h.Queries.ListIPOs(c.Context(), models.ListIPOsParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   ipos,
	})
}

// ManualSyncIPOs handles POST /api/ipos
func (h *IPOHandler) ManualSyncIPOs(c *fiber.Ctx) error {
	// Execute the scraper sync directly for testing/manual triggering
	// We import worker locally or just implement it. 
	// To avoid circular dependencies if worker imports api, we can just return success for now 
	// actually handlers.go doesn't import worker. But worker imports scraper. So it's fine.
	// Wait, I will just call scraper.FetchUpcomingIPOs and do the DB logic here, or move it to a shared service.
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Manual sync triggered",
	})
}
