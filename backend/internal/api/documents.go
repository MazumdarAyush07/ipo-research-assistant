package api

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/MazumdarAyush07/ipo-research/internal/worker"
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
)

// TriggerDocumentDownload handles POST /api/ipos/:id/documents/trigger
func (h *IPOHandler) TriggerDocumentDownload(c *fiber.Ctx) error {
	idParam := c.Params("id")
	ipoID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ipo id"})
	}

	if h.AsynqClient == nil {
		return c.Status(500).JSON(fiber.Map{"error": "asynq client not configured"})
	}

	payload, _ := json.Marshal(worker.DownloadDocumentsPayload{IPOID: ipoID})
	task := asynq.NewTask(worker.TaskDownloadDocuments, payload, asynq.MaxRetry(3))
	
	info, err := h.AsynqClient.Enqueue(task)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to enqueue task: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "download triggered",
		"task_id": info.ID,
	})
}

// TriggerDocumentParse handles POST /api/ipos/:id/parse/trigger
func (h *IPOHandler) TriggerDocumentParse(c *fiber.Ctx) error {
	idParam := c.Params("id")
	ipoID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ipo id"})
	}

	if h.AsynqClient == nil {
		return c.Status(500).JSON(fiber.Map{"error": "asynq client not configured"})
	}

	// Assuming the file is always in storage/<slug>/drhp.pdf.
	// We should fetch the latest document path from the database to be safe.
	// However, the worker expects a FilePath. We can let the parser worker or the handler figure it out.
	// In Phase 5, ParseDocumentPayload takes IPOID and FilePath.
	// We need to fetch the filepath from the documents table.
	ctx := c.Context()
	ipo, err := h.Queries.GetIPO(ctx, ipoID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "IPO not found"})
	}

	// Reconstruct the file path (same logic as document worker)
	// Alternatively, if the file doesn't exist, we can error early
	slug := strings.ToLower(strings.ReplaceAll(ipo.Name, " ", "-"))
	slug = strings.ReplaceAll(slug, ".", "")
	
	// The path from the worker's perspective is relative to /app (which is ../storage)
	// So we pass "../storage/<slug>/drhp.pdf" to the worker.
	filePath := "../storage/" + slug + "/drhp.pdf"

	payload, _ := json.Marshal(worker.ParseDocumentPayload{
		IPOID:    ipoID,
		FilePath: filePath,
	})
	task := asynq.NewTask(worker.TaskParseDocument, payload, asynq.MaxRetry(3))
	
	info, err := h.AsynqClient.Enqueue(task)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to enqueue parse task: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "parse triggered",
		"task_id": info.ID,
	})
}
