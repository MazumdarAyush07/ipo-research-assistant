package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/MazumdarAyush07/ipo-research/internal/scraper"
	"github.com/hibiken/asynq"
)

const (
	TaskSyncGMP           = "tracker:sync_gmp"
	TaskSyncSubscriptions = "tracker:sync_subscriptions"
)

func (p *Processor) HandleSyncGMPTask(ctx context.Context, t *asynq.Task) error {
	log.Printf("Starting Task: %s", t.Type())

	// Fetch all ACTIVE and UPCOMING IPOs
	activeIPOs, err := p.Queries.GetActiveIPOs(ctx)
	if err != nil {
		log.Printf("Failed to fetch active IPOs for GMP: %v", err)
		return err
	}

	updatedCount := 0
	for _, ipo := range activeIPOs {
		gmpData, err := scraper.FetchGMPData(ctx, ipo.Name)
		if err != nil {
			// Not all IPOs will have GMP data (e.g. some SMEs or too early)
			log.Printf("GMP not found for IPO %s: %v", ipo.Name, err)
			continue
		}

		// Insert into DB
		premiumPercent := sql.NullString{
			String: fmt.Sprintf("%.2f", gmpData.PremiumPercent),
			Valid:  true,
		}
		
		gmpAmount := sql.NullString{
			String: fmt.Sprintf("%.2f", gmpData.GMPAmount),
			Valid:  true,
		}

		_, err = p.Queries.CreateGMPHistory(ctx, models.CreateGMPHistoryParams{
			IpoID:          ipo.ID,
			GmpAmount:      gmpAmount,
			PremiumPercent: premiumPercent,
		})
		if err != nil {
			log.Printf("Failed to insert GMP for IPO %s: %v", ipo.Name, err)
			continue
		}
		updatedCount++
	}

	log.Printf("Finished syncing GMP. Updated %d/%d active IPOs.", updatedCount, len(activeIPOs))
	return nil
}

func (p *Processor) HandleSyncSubscriptionsTask(ctx context.Context, t *asynq.Task) error {
	log.Printf("Starting Task: %s", t.Type())

	// Fetch all ACTIVE and UPCOMING IPOs
	activeIPOs, err := p.Queries.GetActiveIPOs(ctx)
	if err != nil {
		log.Printf("Failed to fetch active IPOs for Subscriptions: %v", err)
		return err
	}

	updatedCount := 0
	for _, ipo := range activeIPOs {
		if !ipo.SourceUrl.Valid || ipo.SourceUrl.String == "" {
			continue // We need the specific Chittorgarh URL to scrape subscription
		}

		subData, err := scraper.FetchSubscriptionData(ctx, ipo.SourceUrl.String)
		if err != nil {
			log.Printf("Subscriptions not found for IPO %s: %v", ipo.Name, err)
			continue
		}

		// Insert into DB
		// Categories: QIB, NII, Retail, Total
		categories := map[string]float64{
			"QIB":    subData.QIB,
			"NII":    subData.NII,
			"Retail": subData.Retail,
			"Total":  subData.Total,
		}

		for cat, val := range categories {
			numericVal := sql.NullString{
				String: fmt.Sprintf("%.2f", val),
				Valid:  true,
			}

			_, err = p.Queries.CreateSubscriptionData(ctx, models.CreateSubscriptionDataParams{
				IpoID:           ipo.ID,
				Category:        cat,
				TimesSubscribed: numericVal,
			})
			if err != nil {
				log.Printf("Failed to insert Subscription %s for IPO %s: %v", cat, ipo.Name, err)
			}
		}
		updatedCount++
	}

	log.Printf("Finished syncing Subscriptions. Updated %d/%d active IPOs.", updatedCount, len(activeIPOs))
	return nil
}
