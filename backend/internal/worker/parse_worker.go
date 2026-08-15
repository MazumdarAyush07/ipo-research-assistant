package worker

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/hibiken/asynq"
	"github.com/sqlc-dev/pqtype"
)

const TaskParseDocument = "document:parse"

type ParseDocumentPayload struct {
	IPOID    int64  `json:"ipo_id"`
	FilePath string `json:"file_path"`
}

func (p *Processor) ProcessTaskParseDocument(ctx context.Context, task *asynq.Task) error {
	var payload ParseDocumentPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		log.Printf("failed to unmarshal payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Starting parsing for IPO ID: %d from file: %s", payload.IPOID, payload.FilePath)

	parserURL := os.Getenv("PDF_PARSER_URL")
	if parserURL == "" {
		parserURL = "http://localhost:8000"
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"file_path": payload.FilePath,
		"ipo_id":    payload.IPOID,
		"doc_type":  "DRHP",
	})

	client := &http.Client{
		Timeout: 30 * time.Minute,
	}

	resp, err := client.Post(parserURL+"/parse", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("failed to call pdf-parser: %v", err)
		return fmt.Errorf("failed to call pdf-parser: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("pdf-parser returned status %d", resp.StatusCode)
		return fmt.Errorf("pdf-parser returned status %d", resp.StatusCode)
	}

	var parseResp struct {
		Status     string `json:"status"`
		IPOID      int64  `json:"ipo_id"`
		Financials []struct {
			Year        int     `json:"year"`
			Revenue     float64 `json:"revenue"`
			PAT         float64 `json:"pat"`
			EBITDA      float64 `json:"ebitda"`
			TotalAssets float64 `json:"total_assets"`
			TotalDebt   float64 `json:"total_debt"`
			Equity      float64 `json:"equity"`
		} `json:"financials"`
		ObjectsOfIssue string             `json:"objects_of_issue"`
		RiskFactors    string             `json:"risk_factors"`
		Promoters      string             `json:"promoters"`
		Confidence     map[string]float64 `json:"confidence_scores"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&parseResp); err != nil {
		log.Printf("failed to decode pdf-parser response: %v", err)
		return fmt.Errorf("failed to decode pdf-parser response: %w", err)
	}
	
	log.Printf("Extraction Confidence Scores for IPO %d: %+v", payload.IPOID, parseResp.Confidence)

	// Insert Financials
	for _, f := range parseResp.Financials {
		_, err := p.Queries.CreateOrUpdateFinancials(ctx, models.CreateOrUpdateFinancialsParams{
			IpoID:       payload.IPOID,
			Year:        int32(f.Year),
			Revenue:     sql.NullString{String: fmt.Sprintf("%f", f.Revenue), Valid: true},
			Pat:         sql.NullString{String: fmt.Sprintf("%f", f.PAT), Valid: true},
			Ebitda:      sql.NullString{String: fmt.Sprintf("%f", f.EBITDA), Valid: true},
			TotalAssets: sql.NullString{String: fmt.Sprintf("%f", f.TotalAssets), Valid: true},
			TotalDebt:   sql.NullString{String: fmt.Sprintf("%f", f.TotalDebt), Valid: true},
			Equity:      sql.NullString{String: fmt.Sprintf("%f", f.Equity), Valid: true},
		})
		if err != nil {
			log.Printf("Failed to insert financials for IPO %d: %v", payload.IPOID, err)
		}
	}

	// Insert Raw JSON into AI Analysis table for Phase 9
	rawJSON, _ := json.Marshal(map[string]string{
		"objects_of_issue": parseResp.ObjectsOfIssue,
		"risk_factors":     parseResp.RiskFactors,
		"promoters":        parseResp.Promoters,
	})

	_, err = p.Queries.CreateOrUpdateAIAnalysis(ctx, models.CreateOrUpdateAIAnalysisParams{
		IpoID:   payload.IPOID,
		RawJson: pqtype.NullRawMessage{RawMessage: rawJSON, Valid: true},
	})
	if err != nil {
		log.Printf("Failed to save AI analysis context for IPO %d: %v", payload.IPOID, err)
	}

	log.Printf("Successfully parsed document for IPO ID: %d", payload.IPOID)
	return nil
}
