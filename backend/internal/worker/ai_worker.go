package worker

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/MazumdarAyush07/ipo-research/internal/ai"
	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/hibiken/asynq"
	"github.com/sqlc-dev/pqtype"
)

const TaskAnalyzeDocument = "task:analyze_document"

type AnalyzeDocumentPayload struct {
	IPOID    int64  `json:"ipo_id"`
	FilePath string `json:"file_path"`
}

func (processor *Processor) ProcessTaskAnalyzeDocument(ctx context.Context, task *asynq.Task) error {
	var payload AnalyzeDocumentPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	log.Printf("Starting AI analysis for IPO ID: %d", payload.IPOID)

	// 1. Ask Python sidecar to extract full text
	text, err := extractTextFromPDF(payload.FilePath)
	if err != nil {
		return fmt.Errorf("failed to extract text from PDF: %w", err)
	}

	if len(text) < 1000 {
		return fmt.Errorf("extracted text too short to analyze")
	}

	// 2. Initialize Gemini Client
	geminiClient, err := ai.NewGeminiClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to init Gemini client: %w", err)
	}
	defer geminiClient.Close()

	// 3. Read Prompts
	analystPrompt, err := os.ReadFile("/prompts/analyst_v1.md")
	if err != nil {
		return fmt.Errorf("failed to read analyst prompt: %w", err)
	}
	mergePrompt, err := os.ReadFile("/prompts/analyst_v1_merge.md")
	if err != nil {
		return fmt.Errorf("failed to read merge prompt: %w", err)
	}

	// 4. Chunk text (approx 50k tokens max per chunk, roughly 100k-150k characters)
	// A basic chunker splitting by rough character length for simplicity.
	chunks := chunkText(text, 150000)
	log.Printf("Split DRHP into %d chunks for IPO %d", len(chunks), payload.IPOID)

	var partialResults []ai.AIAnalysisResult
	for i, chunk := range chunks {
		log.Printf("Analyzing chunk %d/%d for IPO %d", i+1, len(chunks), payload.IPOID)

		var res *ai.AIAnalysisResult
		var err error
		maxRetries := 3
		for attempt := 1; attempt <= maxRetries; attempt++ {
			res, err = geminiClient.AnalyzeChunk(ctx, chunk, string(analystPrompt))
			if err == nil {
				break
			}
			log.Printf("Error analyzing chunk %d (attempt %d/%d): %v", i+1, attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(time.Duration(10*attempt) * time.Second) // Exponential-ish backoff
			}
		}

		if err != nil {
			return fmt.Errorf("failed to analyze chunk %d after %d attempts: %w", i+1, maxRetries, err)
		}
		partialResults = append(partialResults, *res)

		// Hard sleep to respect 250k TPM limit (avoiding 429 Too Many Requests)
		if i < len(chunks)-1 {
			time.Sleep(15 * time.Second)
		}
	}

	// 5. Merge Partials if more than 1 chunk
	var finalAnalysis *ai.AIAnalysisResult
	if len(partialResults) > 1 {
		log.Printf("Merging %d partial analyses for IPO %d", len(partialResults), payload.IPOID)
		maxRetriesMerge := 3
		for attempt := 1; attempt <= maxRetriesMerge; attempt++ {
			finalAnalysis, err = geminiClient.MergeAnalyses(ctx, partialResults, string(mergePrompt))
			if err == nil {
				break
			}
			log.Printf("Error merging analyses (attempt %d/%d): %v", attempt, maxRetriesMerge, err)
			if attempt < maxRetriesMerge {
				time.Sleep(time.Duration(10*attempt) * time.Second)
			}
		}
		if err != nil {
			return fmt.Errorf("failed to merge analyses after %d attempts: %w", maxRetriesMerge, err)
		}
	} else if len(partialResults) == 1 {
		finalAnalysis = &partialResults[0]
	} else {
		return fmt.Errorf("no partial analyses generated")
	}

	// 6. Save to DB
	rawJSON, err := json.Marshal(finalAnalysis)
	if err != nil {
		return fmt.Errorf("failed to marshal raw json: %w", err)
	}

	arg := models.CreateOrUpdateAIAnalysisParams{
		IpoID:                 payload.IPOID,
		BusinessModel:         sql.NullString{String: finalAnalysis.BusinessModel, Valid: true},
		Moat:                  sql.NullString{String: finalAnalysis.Moat, Valid: true},
		PromoterRisk:          sql.NullString{String: finalAnalysis.PromoterRisk, Valid: true},
		LegalCases:            sql.NullString{String: finalAnalysis.LegalCases, Valid: true},
		CustomerConcentration: sql.NullString{String: finalAnalysis.CustomerConcentration, Valid: true},
		DebtAssessment:        sql.NullString{String: finalAnalysis.DebtAssessment, Valid: true},
		KeyRisks:              sql.NullString{String: finalAnalysis.KeyRisks, Valid: true},
		RedFlags:              sql.NullString{String: finalAnalysis.RedFlags, Valid: true},
		ManagementAssumptions: sql.NullString{String: finalAnalysis.ManagementAssumptions, Valid: true},
		IndustryOutlook:       sql.NullString{String: finalAnalysis.IndustryOutlook, Valid: true},
		RawJson:               pqtype.NullRawMessage{RawMessage: rawJSON, Valid: true},
	}

	_, err = processor.Queries.CreateOrUpdateAIAnalysis(ctx, arg)
	if err != nil {
		return fmt.Errorf("failed to save ai analysis to db: %w", err)
	}

	log.Printf("Successfully completed AI analysis for IPO ID: %d", payload.IPOID)
	return nil
}

func extractTextFromPDF(filePath string) (string, error) {
	url := os.Getenv("PDF_PARSER_URL")
	if url == "" {
		url = "http://localhost:8000"
	}
	url = url + "/extract-text"

	reqBody := fmt.Sprintf(`{"file_path": "%s"}`, filePath)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("pdf parser returned status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Status string `json:"status"`
		Text   string `json:"text"`
		Error  string `json:"error"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", err
	}

	if result.Status != "success" {
		return "", fmt.Errorf("parser error: %s", result.Error)
	}

	return result.Text, nil
}

func chunkText(text string, chunkSize int) []string {
	var chunks []string
	runes := []rune(text)
	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	return chunks
}
