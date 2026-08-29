package api

import (
	"path/filepath"
	"strings"

	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/MazumdarAyush07/ipo-research/internal/scoring"
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
		taskVal := asynq.NewTask("tracker:sync_valuation", nil)
		h.AsynqClient.Enqueue(taskVal)
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

// GetScore handles GET /api/ipos/:id/score
func (h *IPOHandler) GetScore(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ipo id"})
	}

	score, err := h.Queries.GetScoreByIPO(c.Context(), int64(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "score not found"})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   score,
	})
}

// TriggerScore handles POST /api/ipos/:id/score/trigger
func (h *IPOHandler) TriggerScore(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ipo id"})
	}

	res, err := scoring.ScoreIPO(c.Context(), h.Queries, h.PeerService, int64(id))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to calculate score", "details": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Score calculated successfully",
		"data":    res,
	})
}

// TriggerReportGeneration handles POST /api/ipos/:id/report/generate
func (h *IPOHandler) TriggerReportGeneration(c *fiber.Ctx) error {
	_, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ipo id"})
	}

	if h.AsynqClient != nil {
		payload := []byte(`{"ipo_id":` + c.Params("id") + `}`)
		task := asynq.NewTask("report:generate", payload)
		if _, err := h.AsynqClient.Enqueue(task); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "failed to enqueue report generation task"})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Report generation triggered",
	})
}

// GetReport handles GET /api/ipos/:id/report
func (h *IPOHandler) GetReport(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ipo id"})
	}

	report, err := h.Queries.GetReportByIPO(c.Context(), models.GetReportByIPOParams{
		IpoID:  int64(id),
		Format: "HTML",
	})
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "report not found"})
	}

	// Make the path absolute or relative to the current working directory
	return c.SendFile(filepath.Clean(filepath.Join("..", report.FilePath)))
}
