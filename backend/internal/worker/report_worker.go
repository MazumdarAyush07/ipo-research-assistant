package worker

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/MazumdarAyush07/ipo-research/internal/storage"
	"github.com/MazumdarAyush07/ipo-research/internal/utils"
	"github.com/hibiken/asynq"
)

const TaskGenerateReport = "report:generate"

type GenerateReportPayload struct {
	IPOID int64 `json:"ipo_id"`
}

type ReportData struct {
	IPO         models.Ipo
	Financials  []models.Financial
	Peers       []models.PeerCompany
	Valuation   models.Valuation
	Score       models.Score
	AI          models.AiAnalysis
	GeneratedAt string

	// Clean fields for template
	FinalScoreInt        int
	RecommendationString string
	FinancialsScoreInt   int
	ValuationScoreInt    int
	SubscriptionScoreInt int
	GmpScoreInt          int
	PromoterScoreInt     int
	IndustryScoreInt     int
	RiskScoreInt         int
	MarketDemandScore    int
	AIAssessmentScore    int

	FinancialsWidth   float64
	ValuationWidth    float64
	MarketDemandWidth float64
	AIAssessmentWidth float64

	RedFlagsList []string
}

func (p *Processor) HandleGenerateReportTask(ctx context.Context, t *asynq.Task) (err error) {
	log.Printf("Starting Task: %s", t.Type())
	defer func() {
		if err != nil {
			log.Printf("Task %s failed: %v", t.Type(), err)
		}
	}()
	
	var payload GenerateReportPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	
	ipoID := payload.IPOID

	// 1. Fetch all data for the IPO
	ipo, err := p.Queries.GetIPO(ctx, ipoID)
	if err != nil {
		return fmt.Errorf("failed to fetch IPO %d: %w", ipoID, err)
	}

	financials, err := p.Queries.GetFinancialsByIPO(ctx, ipoID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: failed to fetch financials for report: %v", err)
	}

	peers, err := p.Queries.GetPeerCompaniesByIPO(ctx, ipoID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: failed to fetch peers for report: %v", err)
	}

	val, err := p.Queries.GetValuationByIPO(ctx, ipoID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: failed to fetch valuation for report: %v", err)
	}

	score, err := p.Queries.GetScoreByIPO(ctx, ipoID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: failed to fetch score for report: %v", err)
	}

	aiData, err := p.Queries.GetAIAnalysisByIPO(ctx, ipoID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: failed to fetch AI analysis for report: %v", err)
	}

	parseInt := func(ns sql.NullString) int {
		if !ns.Valid {
			return 0
		}
		val, _ := strconv.ParseFloat(ns.String, 64)
		return int(val)
	}

	// 2. Prepare Template Data
	data := ReportData{
		IPO:         ipo,
		Financials:  financials,
		Peers:       peers,
		Valuation:   val,
		Score:       score,
		AI:          aiData,
		GeneratedAt: time.Now().Format("January 02, 2006 15:04:05"),
		
		FinalScoreInt:        parseInt(score.FinalScore),
		RecommendationString: score.Recommendation.String,
		FinancialsScoreInt:   parseInt(score.FinancialsScore),
		ValuationScoreInt:    parseInt(score.ValuationScore),
		SubscriptionScoreInt: parseInt(score.SubscriptionScore),
		GmpScoreInt:          parseInt(score.GmpScore),
		PromoterScoreInt:     parseInt(score.PromoterScore),
		IndustryScoreInt:     parseInt(score.IndustryScore),
		RiskScoreInt:         parseInt(score.RiskScore),
	}
	data.MarketDemandScore = data.SubscriptionScoreInt + data.GmpScoreInt
	data.AIAssessmentScore = data.PromoterScoreInt + data.IndustryScoreInt + data.RiskScoreInt

	data.FinancialsWidth = float64(data.FinancialsScoreInt) * 2.5
	data.ValuationWidth = float64(data.ValuationScoreInt) * 5.0
	data.MarketDemandWidth = float64(data.MarketDemandScore) * 10.0
	data.AIAssessmentWidth = float64(data.AIAssessmentScore) * (100.0 / 30.0)

	var redFlagsStr string
	if aiData.RedFlags.Valid {
		redFlagsStr = aiData.RedFlags.String
	} else if aiData.KeyRisks.Valid {
		redFlagsStr = aiData.KeyRisks.String
	}
	
	re := regexp.MustCompile(`\s*(?:\d+\.)\s+`)
	for _, p := range re.Split(redFlagsStr, -1) {
		p = strings.TrimSpace(p)
		if p != "" {
			data.RedFlagsList = append(data.RedFlagsList, p)
		}
	}
	if len(data.RedFlagsList) == 0 && redFlagsStr != "" {
		data.RedFlagsList = []string{redFlagsStr}
	}

	// 3. Load and parse the HTML template
	tmpl, err := template.ParseFiles("templates/report.html")
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// 4. Render to buffer
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// 5. Save HTML file to R2
	slug := utils.GenerateSlug(ipo.Name)
	s3Key := fmt.Sprintf("reports/%s.html", slug)
	
	r2Client, err := storage.NewR2Client(ctx)
	if err != nil {
		return fmt.Errorf("failed to init R2 client: %v", err)
	}

	if err := r2Client.UploadStream(ctx, s3Key, &buf, "text/html"); err != nil {
		return fmt.Errorf("failed to upload report to R2: %w", err)
	}
	
	log.Printf("Successfully generated report at R2 key: %s", s3Key)

	// 6. Update Database
	_, err = p.Queries.CreateOrUpdateReport(ctx, models.CreateOrUpdateReportParams{
		IpoID:    ipoID,
		FilePath: s3Key,
		Format:   "HTML",
	})
	if err != nil {
		return fmt.Errorf("failed to update report DB record: %w", err)
	}

	return nil
}
