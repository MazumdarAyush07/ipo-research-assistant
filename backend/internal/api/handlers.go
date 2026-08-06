package api

import (
	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
)

// IPOHandler holds dependencies for API routes
type IPOHandler struct {
	Queries     *models.Queries
	AsynqClient *asynq.Client
}

func NewIPOHandler(q *models.Queries, client *asynq.Client) *IPOHandler {
	return &IPOHandler{Queries: q, AsynqClient: client}
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
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Manual sync triggered",
	})
}

// GetFinancials handles GET /api/ipos/:id/financials
func (h *IPOHandler) GetFinancials(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ipo id"})
	}

	financials, err := h.Queries.GetFinancialsByIPO(c.Context(), int64(id))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   financials,
	})
}

// GetAIAnalysis handles GET /api/ipos/:id/analysis
func (h *IPOHandler) GetAIAnalysis(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ipo id"})
	}

	analysis, err := h.Queries.GetAIAnalysisByIPO(c.Context(), int64(id))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   analysis,
	})
}
