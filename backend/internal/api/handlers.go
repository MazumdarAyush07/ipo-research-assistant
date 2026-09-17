package api

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/MazumdarAyush07/ipo-research/internal/scoring"
	"github.com/MazumdarAyush07/ipo-research/internal/services"
	"github.com/MazumdarAyush07/ipo-research/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
)

// IPOHandler holds dependencies for API routes
type IPOHandler struct {
	Queries        *models.Queries
	AsynqClient    *asynq.Client
	AsynqInspector *asynq.Inspector
	PeerService    *services.PeerService
}

func NewIPOHandler(q *models.Queries, client *asynq.Client, inspector *asynq.Inspector, ps *services.PeerService) *IPOHandler {
	return &IPOHandler{Queries: q, AsynqClient: client, AsynqInspector: inspector, PeerService: ps}
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
		task := asynq.NewTask("ipo:sync", nil, asynq.Retention(24*time.Hour))
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

// GetParsingAudit handles GET /api/admin/parsing-audit
func (h *IPOHandler) GetParsingAudit(c *fiber.Ctx) error {
	ctx := c.Context()

	totalIPOs, err := h.Queries.CountIPOs(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count IPOs"})
	}

	financialsCount, err := h.Queries.CountIPOsWithFinancials(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count financials"})
	}

	return c.JSON(fiber.Map{
		"total_ipos":       totalIPOs,
		"total_financials": financialsCount,
		"missing_parsing":  totalIPOs - financialsCount,
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

	// Check if AI Analysis already exists to prevent wasting tokens
	analysis, err := h.Queries.GetAIAnalysisByIPO(c.Context(), int64(id))
	if err == nil && analysis.ID != 0 && analysis.RedFlags.Valid {
		return c.JSON(fiber.Map{
			"status":  "skipped",
			"message": "IPO already analyzed",
		})
	}

	slug := utils.GenerateSlug(ipo.Name)
	filePath := filepath.Join("../storage", slug, "drhp.pdf")

	if h.AsynqClient != nil {
		payload, _ := json.Marshal(map[string]interface{}{"ipo_id": id, "file_path": filePath})
		task := asynq.NewTask("task:analyze_document", payload, asynq.Retention(24*time.Hour))
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
		taskGMP := asynq.NewTask("tracker:sync_gmp", nil, asynq.Retention(24*time.Hour))
		h.AsynqClient.Enqueue(taskGMP)
		taskSub := asynq.NewTask("tracker:sync_subscriptions", nil, asynq.Retention(24*time.Hour))
		h.AsynqClient.Enqueue(taskSub)
		taskVal := asynq.NewTask("tracker:sync_valuation", nil, asynq.Retention(24*time.Hour))
		h.AsynqClient.Enqueue(taskVal)
		taskPeers := asynq.NewTask("tracker:sync_peers", nil, asynq.Retention(24*time.Hour))
		h.AsynqClient.Enqueue(taskPeers)
	}
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Tracker sync triggered for all active IPOs",
	})
}

// TriggerSyncPeers handles POST /api/trackers/peers/sync
func (h *IPOHandler) TriggerSyncPeers(c *fiber.Ctx) error {
	if h.AsynqClient != nil {
		taskPeers := asynq.NewTask("tracker:sync_peers", nil, asynq.Retention(24*time.Hour))
		h.AsynqClient.Enqueue(taskPeers)
	}
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Peer sync triggered for all IPOs",
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
	ipoID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ipo id"})
	}

	if h.AsynqClient != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"ipo_id": ipoID,
		})
		task := asynq.NewTask("report:generate", payload, asynq.Retention(24*time.Hour))
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

// GetAudit handles GET /api/admin/audit
func (h *IPOHandler) GetAudit(c *fiber.Ctx) error {
	ipos, err := h.Queries.ListIPOs(c.Context(), models.ListIPOsParams{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch IPOs"})
	}

	totalExpected := len(ipos)
	downloaded := 0
	missing := []string{}

	for _, ipo := range ipos {
		slug := utils.GenerateSlug(ipo.Name)
		path := filepath.Join("../storage", slug, "drhp.pdf")

		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			downloaded++
		} else {
			missing = append(missing, ipo.Name)
		}
	}

	return c.JSON(fiber.Map{
		"downloaded":     downloaded,
		"total_expected": totalExpected,
		"below_50_pages": 0, // Simplified for native Go implementation
		"missing":        missing,
		"unexpected_eof": 0,
		"errors":         []string{},
	})
}

// GetQueueStats returns stats from the Asynq inspector
func (h *IPOHandler) GetQueueStats(c *fiber.Ctx) error {
	if h.AsynqInspector == nil {
		return c.Status(500).JSON(fiber.Map{"error": "asynq inspector not configured"})
	}

	queues, err := h.AsynqInspector.Queues()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to list queues", "details": err.Error()})
	}

	// Always ensure we check our core queues even if the set was somehow dropped from Redis
	knownQueues := []string{"default", "critical"}
	queueSet := make(map[string]bool)
	for _, q := range queues {
		queueSet[q] = true
	}
	for _, q := range knownQueues {
		if !queueSet[q] {
			queues = append(queues, q)
		}
	}

	var stats []interface{}
	for _, qname := range queues {
		info, err := h.AsynqInspector.GetQueueInfo(qname)
		if err != nil {
			continue
		}
		stats = append(stats, fiber.Map{
			"queue":     info.Queue,
			"active":    info.Active,
			"pending":   info.Pending,
			"scheduled": info.Scheduled,
			"retry":     info.Retry,
			"archived":  info.Archived,
			"completed": info.Completed,
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   stats,
	})
}

// GetAnalysisAudit handles GET /api/admin/analysis-audit
func (h *IPOHandler) GetAnalysisAudit(c *fiber.Ctx) error {
	ctx := c.Context()

	totalIPOs, err := h.Queries.CountIPOs(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count IPOs"})
	}

	analysisCount, err := h.Queries.CountIPOsWithAIAnalysis(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count analysis"})
	}

	return c.JSON(fiber.Map{
		"total_ipos": totalIPOs,
		"completed":  analysisCount,
		"missing":    totalIPOs - analysisCount,
	})
}

// GetTrackerAudit handles GET /api/admin/tracker-audit
func (h *IPOHandler) GetTrackerAudit(c *fiber.Ctx) error {
	ctx := c.Context()

	hoursStr := c.Query("hours", "0")
	var hours int32
	if h, err := strconv.Atoi(hoursStr); err == nil {
		hours = int32(h)
	}

	peersTracked, err := h.Queries.CountRecentPeersTracked(ctx, hours)
	if err != nil {
		log.Printf("Tracker audit db error (peers): %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count peers tracked"})
	}

	gmpTracked, err := h.Queries.CountRecentGMPTracked(ctx, hours)
	if err != nil {
		log.Printf("Tracker audit db error (gmp): %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count gmp tracked"})
	}

	subscriptionsTracked, err := h.Queries.CountRecentSubscriptionsTracked(ctx, hours)
	if err != nil {
		log.Printf("Tracker audit db error (subscriptions): %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count subscriptions tracked"})
	}

	valuationsTracked, err := h.Queries.CountRecentValuationsTracked(ctx, hours)
	if err != nil {
		log.Printf("Tracker audit db error (valuations): %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count valuations tracked"})
	}

	return c.JSON(fiber.Map{
		"peers_tracked":         peersTracked,
		"gmp_tracked":           gmpTracked,
		"subscriptions_tracked": subscriptionsTracked,
		"valuations_tracked":    valuationsTracked,
	})
}

// GetScoringAudit handles GET /api/admin/scoring-audit
func (h *IPOHandler) GetScoringAudit(c *fiber.Ctx) error {
	ctx := c.Context()

	totalIPOs, err := h.Queries.CountIPOs(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count IPOs"})
	}

	scoringCount, err := h.Queries.CountIPOsWithScores(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count scores"})
	}

	return c.JSON(fiber.Map{
		"total_ipos": totalIPOs,
		"completed":  scoringCount,
		"missing":    totalIPOs - scoringCount,
	})
}

// GetReportAudit handles GET /api/admin/report-audit
func (h *IPOHandler) GetReportAudit(c *fiber.Ctx) error {
	ctx := c.Context()

	totalIPOs, err := h.Queries.CountIPOs(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count IPOs"})
	}

	ipos, err := h.Queries.ListIPOs(ctx, models.ListIPOsParams{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch IPOs"})
	}

	completed := int64(0)
	for _, ipo := range ipos {
		slug := utils.GenerateSlug(ipo.Name)
		path := filepath.Join("../storage", slug, "report.html")

		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			completed++
		}
	}

	return c.JSON(fiber.Map{
		"total_ipos": totalIPOs,
		"completed":  completed,
		"missing":    totalIPOs - completed,
	})
}
