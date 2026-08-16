package api

import (
	"path/filepath"
	"strings"

	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/MazumdarAyush07/ipo-research/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
)

// IPOHandler holds dependencies for API routes
type IPOHandler struct {
	Queries     *models.Queries
	AsynqClient *asynq.Client
	PeerService *services.PeerService
}

func NewIPOHandler(q *models.Queries, client *asynq.Client, ps *services.PeerService) *IPOHandler {
	return &IPOHandler{Queries: q, AsynqClient: client, PeerService: ps}
}

// ListIPOs handles GET /api/ipos
func (h *IPOHandler) ListIPOs(c *fiber.Ctx) error {
	if h.Queries == nil {
		return c.JSON(fiber.Map{"status": "success", "data": []string{}})
	}

	ipos, err := h.Queries.ListIPOs(c.Context(), models.ListIPOsParams{
		Limit:  1000,
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
	if h.AsynqClient != nil {
		task := asynq.NewTask("ipo:sync", nil)
		if _, err := h.AsynqClient.Enqueue(task); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to enqueue sync task"})
		}
	}

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

	metrics, _ := services.CalculateMetrics(financials)

	return c.JSON(fiber.Map{
		"status":  "success",
		"data":    financials,
		"metrics": metrics,
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

// TriggerAIAnalysis handles POST /api/ipos/:id/analysis/trigger
func (h *IPOHandler) TriggerAIAnalysis(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ipo id"})
	}

	ipo, err := h.Queries.GetIPO(c.Context(), int64(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "IPO not found"})
	}

	slug := strings.ReplaceAll(strings.ToLower(ipo.Name), " ", "-")
	filePath := filepath.Join("../storage", slug, "drhp.pdf")

	if h.AsynqClient != nil {
		payload := []byte(`{"ipo_id":` + c.Params("id") + `,"file_path":"` + filePath + `"}`)
		task := asynq.NewTask("task:analyze_document", payload)
		if _, err := h.AsynqClient.Enqueue(task); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to enqueue ai analysis task"})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "AI analysis triggered",
	})
}

// GetGMP handles GET /api/ipos/:id/gmp
func (h *IPOHandler) GetGMP(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ipo id"})
	}

	gmp, err := h.Queries.GetLatestGMP(c.Context(), int64(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   gmp,
	})
}

// TriggerTrackers handles POST /api/trackers/sync
func (h *IPOHandler) TriggerTrackers(c *fiber.Ctx) error {
	if h.AsynqClient != nil {
		taskGMP := asynq.NewTask("tracker:sync_gmp", nil)
		h.AsynqClient.Enqueue(taskGMP)
		taskSub := asynq.NewTask("tracker:sync_subscriptions", nil)
		h.AsynqClient.Enqueue(taskSub)
	}
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Tracker sync triggered for all active IPOs",
	})
}

// GetSubscriptions handles GET /api/ipos/:id/subscriptions
func (h *IPOHandler) GetSubscriptions(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ipo id"})
	}

	subs, err := h.Queries.GetLatestSubscription(c.Context(), int64(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   subs,
	})
}
