package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/MazumdarAyush07/ipo-research/internal/scraper"
	"github.com/hibiken/asynq"
)

type Processor struct {
	Queries     *models.Queries
	AsynqClient *asynq.Client
}

func NewProcessor(q *models.Queries, client *asynq.Client) *Processor {
	return &Processor{Queries: q, AsynqClient: client}
}

func (p *Processor) HandleSyncIPOsTask(ctx context.Context, t *asynq.Task) error {
	log.Printf("Starting Task: %s", t.Type())

	url := "https://www.chittorgarh.com/report/ipo-in-india-list-main-board-sme/82/all/"
	ipos, err := scraper.FetchUpcomingIPOs(ctx, url)
	if err != nil {
		log.Printf("Failed to fetch IPOs: %v", err)
		return err
	}

	log.Printf("Scraped %d IPOs. Ready for DB insertion.", len(ipos))
	
	insertedCount := 0
	for _, ipo := range ipos {
		// Deduplication check
		_, err := p.Queries.GetIPOByName(ctx, ipo.Name)
		if err == nil {
			// IPO exists, skip
			continue
		}

		// Parse dates (e.g. "10-Aug-2026" -> sql.NullTime)
		var parsedOpenDate sql.NullTime
		if ipo.OpenDate != "" {
			if t, err := time.Parse("02-Jan-2006", ipo.OpenDate); err == nil {
				parsedOpenDate = sql.NullTime{Time: t, Valid: true}
			}
		}

		var parsedCloseDate sql.NullTime
		if ipo.CloseDate != "" {
			if t, err := time.Parse("02-Jan-2006", ipo.CloseDate); err == nil {
				parsedCloseDate = sql.NullTime{Time: t, Valid: true}
			}
		}

		var parsedSourceUrl sql.NullString
		if ipo.SourceUrl != "" {
			parsedSourceUrl = sql.NullString{String: ipo.SourceUrl, Valid: true}
		}

		// Insert new IPO
		insertedIPO, err := p.Queries.CreateIPO(ctx, models.CreateIPOParams{
			Name:         ipo.Name,
			ExchangeType: sql.NullString{String: ipo.ExchangeType, Valid: true},
			OpenDate:     parsedOpenDate,
			CloseDate:    parsedCloseDate,
			Status:       sql.NullString{String: "UPCOMING", Valid: true},
			SourceUrl:    parsedSourceUrl,
		})
		if err != nil {
			log.Printf("Failed to insert IPO %s: %v", ipo.Name, err)
			continue
		}
		insertedCount++

		// Enqueue Document Download Task for the newly inserted IPO
		if p.AsynqClient != nil {
			payload, _ := json.Marshal(DownloadDocumentsPayload{IPOID: insertedIPO.ID})
			task := asynq.NewTask(TaskDownloadDocuments, payload, asynq.MaxRetry(3))
			if _, err := p.AsynqClient.Enqueue(task); err != nil {
				log.Printf("Failed to enqueue download task for IPO %d: %v", insertedIPO.ID, err)
			} else {
				log.Printf("Enqueued download task for IPO %d", insertedIPO.ID)
			}
		}
	}

	log.Printf("Successfully inserted %d new IPOs", insertedCount)
	return nil
}
