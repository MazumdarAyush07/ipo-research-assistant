package api

import (
	"encoding/json"
	"strconv"

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
