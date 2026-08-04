package worker

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/MazumdarAyush07/ipo-research/internal/scraper"
	"github.com/hibiken/asynq"
)

type Processor struct {
	Queries *models.Queries
}

func NewProcessor(q *models.Queries) *Processor {
	return &Processor{Queries: q}
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

		// Insert new IPO
		_, err = p.Queries.CreateIPO(ctx, models.CreateIPOParams{
			Name:         ipo.Name,
			ExchangeType: sql.NullString{String: ipo.ExchangeType, Valid: true},
			OpenDate:     parsedOpenDate,
			CloseDate:    parsedCloseDate,
			Status:       sql.NullString{String: "UPCOMING", Valid: true},
		})
		if err != nil {
			log.Printf("Failed to insert IPO %s: %v", ipo.Name, err)
			continue
		}
		insertedCount++
	}

	log.Printf("Successfully inserted %d new IPOs", insertedCount)
	return nil
}
