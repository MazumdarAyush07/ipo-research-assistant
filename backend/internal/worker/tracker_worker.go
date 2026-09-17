package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/MazumdarAyush07/ipo-research/internal/ai"
	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/MazumdarAyush07/ipo-research/internal/scraper"
	"github.com/hibiken/asynq"
)

const (
	TaskSyncGMP           = "tracker:sync_gmp"
	TaskSyncSubscriptions = "tracker:sync_subscriptions"
	TaskSyncValuation     = "tracker:sync_valuation"
	TaskSyncPeers         = "tracker:sync_peers"
)

func (p *Processor) HandleSyncGMPTask(ctx context.Context, t *asynq.Task) error {
	log.Printf("Starting Task: %s", t.Type())

	// Read allowed sectors from peers config
	var allowedSectors []string
	if b, err := os.ReadFile("config/peers.json"); err == nil {
		var peersConfig map[string]interface{}
		if err := json.Unmarshal(b, &peersConfig); err == nil {
			for k := range peersConfig {
				allowedSectors = append(allowedSectors, k)
			}
		}
	}

	// Initialize Gemini client for Sector Inference
	aiClient, aiErr := ai.NewGeminiClient(ctx)
	if aiErr != nil {
		log.Printf("Warning: Failed to initialize AI client. Sector inference will be skipped: %v", aiErr)
	} else {
		defer aiClient.Close()
	}

	// Fetch the 100 most recent IPOs to ensure we backfill recently CLOSED ones too
	activeIPOs, err := p.Queries.GetActiveIPOs(ctx)
	if err != nil {
		log.Printf("Failed to fetch IPOs for GMP: %v", err)
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

		// Opportunistically save the price band and listing date into the IPO and valuation tables
		if (gmpData.PriceBand > 0 && (!ipo.PriceBandHigh.Valid || ipo.PriceBandHigh.String == "" || ipo.PriceBandHigh.String == "0.00")) || (gmpData.ListingDate != "" && !ipo.ListingDate.Valid) {
			priceBandStr := fmt.Sprintf("%.2f", gmpData.PriceBand)
			listingTime := ipo.ListingDate
			if gmpData.ListingDate != "" {
				layouts := []string{"Monday, January 2, 2006", "January 2, 2006", "02-Jan-2006", "2-Jan-2006", "02-January-2006", "2-January-2006", "January 02, 2006", "January 02 2006"}
				for _, l := range layouts {
					// Need to append the year if it's missing, but ipowatch detail pages usually have "August 19, 2026"
					if t, err := time.Parse(l, gmpData.ListingDate); err == nil {
						listingTime = sql.NullTime{Time: t, Valid: true}
						break
					}
				}
			}

			var pbLow, pbHigh sql.NullString
			if gmpData.PriceBand > 0 {
				pbLow = sql.NullString{String: priceBandStr, Valid: true}
				pbHigh = sql.NullString{String: priceBandStr, Valid: true}
			} else {
				pbLow = ipo.PriceBandLow
				pbHigh = ipo.PriceBandHigh
			}

			sectorStr := ipo.Sector.String
			if (sectorStr == "" || sectorStr == "Others") && aiClient != nil && gmpData.AboutCompany != "" && len(allowedSectors) > 0 {
				// Avoid hitting the 15 RPM free tier limit or quota failures
				time.Sleep(10 * time.Second)

				inferredSector, err := aiClient.InferSector(ctx, gmpData.AboutCompany, allowedSectors)
				if err == nil && inferredSector != "" {
					sectorStr = inferredSector
				}
			}
			sector := sql.NullString{String: sectorStr, Valid: sectorStr != ""}

			_, _ = p.Queries.UpdateIPODetails(ctx, models.UpdateIPODetailsParams{
				ID:            ipo.ID,
				Sector:        sector,
				PriceBandLow:  pbLow,
				PriceBandHigh: pbHigh,
				ListingDate:   listingTime,
			})

			if gmpData.PriceBand > 0 {
				// Also save into the valuation table
				_, _ = p.Queries.CreateOrUpdateValuation(ctx, models.CreateOrUpdateValuationParams{
					IpoID:      ipo.ID,
					IssuePrice: sql.NullString{String: priceBandStr, Valid: true},
					MarketCap:  sql.NullString{},
					PeRatio:    sql.NullString{},
					PbRatio:    sql.NullString{},
				})
			}
		}
	}

	log.Printf("Finished syncing GMP. Updated %d/%d active IPOs.", updatedCount, len(activeIPOs))
	return nil
}

func (p *Processor) HandleSyncSubscriptionsTask(ctx context.Context, t *asynq.Task) error {
	log.Printf("Starting Task: %s", t.Type())

	// Fetch the 100 most recent IPOs to ensure we backfill recently CLOSED ones too
	activeIPOs, err := p.Queries.GetActiveIPOs(ctx)
	if err != nil {
		log.Printf("Failed to fetch IPOs for Subscriptions: %v", err)
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


func (p *Processor) HandleSyncValuationTask(ctx context.Context, t *asynq.Task) error {
	log.Printf("Starting Task: %s", t.Type())

	// Read allowed sectors from peers config
	var allowedSectors []string
	if b, err := os.ReadFile("config/peers.json"); err == nil {
		var peersConfig map[string]interface{}
		if err := json.Unmarshal(b, &peersConfig); err == nil {
			for k := range peersConfig {
				allowedSectors = append(allowedSectors, k)
			}
		}
	}

	// Initialize Gemini client for Sector Inference (fallback)
	aiClient, aiErr := ai.NewGeminiClient(ctx)
	if aiErr != nil {
		log.Printf("Warning: Failed to initialize AI client. Fallback sector inference will be skipped: %v", aiErr)
	} else {
		defer aiClient.Close()
	}

	// Fetch the 100 most recent IPOs to ensure we backfill CLOSED ones too
	activeIPOs, err := p.Queries.GetActiveIPOs(ctx)
	if err != nil {
		log.Printf("Failed to fetch IPOs for Valuation: %v", err)
		return err
	}

	updatedCount := 0
	for _, ipo := range activeIPOs {
		if !ipo.SourceUrl.Valid || ipo.SourceUrl.String == "" {
			continue
		}

		valData, err := scraper.FetchValuationData(ctx, ipo.SourceUrl.String)
		if err != nil {
			log.Printf("Valuation not found for IPO %s: %v", ipo.Name, err)
			continue
		}

		issuePrice := sql.NullString{String: fmt.Sprintf("%.2f", valData.IssuePrice), Valid: valData.IssuePrice > 0}
		marketCap := sql.NullString{String: fmt.Sprintf("%.2f", valData.MarketCap), Valid: valData.MarketCap > 0}
		peRatio := sql.NullString{String: fmt.Sprintf("%.2f", valData.PE), Valid: valData.PE > 0}
		pbRatio := sql.NullString{String: fmt.Sprintf("%.2f", valData.PB), Valid: valData.PB > 0}

		_, err = p.Queries.CreateOrUpdateValuation(ctx, models.CreateOrUpdateValuationParams{
			IpoID:      ipo.ID,
			IssuePrice: issuePrice,
			MarketCap:  marketCap,
			PeRatio:    peRatio,
			PbRatio:    pbRatio,
		})
		if err != nil {
			log.Printf("Failed to insert Valuation for IPO %s: %v", ipo.Name, err)
		} else {
			updatedCount++
		}

		// Retain existing data if scraper couldn't find it
		sectorStr := ipo.Sector.String
		if valData.Sector != "" {
			sectorStr = valData.Sector
		} else if (sectorStr == "" || sectorStr == "Others") && aiClient != nil && len(allowedSectors) > 0 {
			// AI Sector Inference FALLBACK (guessing strictly from name since we don't have description)
			time.Sleep(5 * time.Second) // Respect rate limits

			fallbackPrompt := fmt.Sprintf("We only have the name for this company: %s. Please make your best guess.", ipo.Name)
			inferredSector, err := aiClient.InferSector(ctx, fallbackPrompt, allowedSectors)
			if err == nil && inferredSector != "" {
				sectorStr = inferredSector
			}
		}
		sector := sql.NullString{String: sectorStr, Valid: sectorStr != ""}

		priceLow := ipo.PriceBandLow.String
		if valData.PriceBandLow > 0 {
			priceLow = fmt.Sprintf("%.2f", valData.PriceBandLow)
		}
		priceBandLow := sql.NullString{String: priceLow, Valid: priceLow != ""}

		priceHigh := ipo.PriceBandHigh.String
		if valData.PriceBandHigh > 0 {
			priceHigh = fmt.Sprintf("%.2f", valData.PriceBandHigh)
		}
		priceBandHigh := sql.NullString{String: priceHigh, Valid: priceHigh != ""}

		listingTime := ipo.ListingDate
		if valData.ListingDate != "" {
			layouts := []string{"Monday, January 2, 2006", "January 2, 2006", "02-Jan-2006", "2-Jan-2006"}
			for _, l := range layouts {
				if t, err := time.Parse(l, valData.ListingDate); err == nil {
					listingTime = sql.NullTime{Time: t, Valid: true}
					break
				}
			}
		}

		_, err = p.Queries.UpdateIPODetails(ctx, models.UpdateIPODetailsParams{
			ID:            ipo.ID,
			Sector:        sector,
			PriceBandLow:  priceBandLow,
			PriceBandHigh: priceBandHigh,
			ListingDate:   listingTime,
		})
		if err != nil {
			log.Printf("Failed to update IPO Details for %s: %v", ipo.Name, err)
		}
	}

	log.Printf("Finished syncing Valuation. Updated %d/%d active IPOs.", updatedCount, len(activeIPOs))
	return nil
}

func (p *Processor) HandleSyncPeersTask(ctx context.Context, t *asynq.Task) error {
	log.Printf("Starting Task: sync_peers")

	activeIPOs, err := p.Queries.GetActiveIPOs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active ipos: %w", err)
	}

	if p.PeerService == nil {
		return fmt.Errorf("PeerService is not initialized")
	}

	updatedCount := 0
	for _, ipo := range activeIPOs {
		sector := ipo.Sector.String
		if !ipo.Sector.Valid || sector == "" || sector == "Others" {
			continue
		}

		fetchedPeers, err := p.PeerService.FetchPeersForSector(ctx, sector)
		if err != nil {
			log.Printf("Warning: failed to fetch peers for IPO %d (Sector: %s): %v", ipo.ID, sector, err)
			continue
		}

		// Delete existing peers
		err = p.Queries.DeletePeerCompaniesByIPO(ctx, ipo.ID)
		if err != nil {
			log.Printf("Error deleting old peers for IPO %d: %v", ipo.ID, err)
			continue
		}

		// Insert updated peers
		for _, peer := range fetchedPeers {
			_, err = p.Queries.InsertPeerCompany(ctx, models.InsertPeerCompanyParams{
				IpoID:     ipo.ID,
				Name:      peer.Name,
				Ticker:    sql.NullString{String: peer.Ticker, Valid: peer.Ticker != ""},
				Pe:        sql.NullString{String: fmt.Sprintf("%.2f", peer.PE), Valid: peer.PE != 0},
				Pb:        sql.NullString{String: fmt.Sprintf("%.2f", peer.PB), Valid: peer.PB != 0},
				MarketCap: sql.NullString{String: fmt.Sprintf("%.2f", peer.MarketCap), Valid: peer.MarketCap != 0},
			})
			if err != nil {
				log.Printf("Error inserting peer %s for IPO %d: %v", peer.Ticker, ipo.ID, err)
			}
		}
		updatedCount++
		time.Sleep(5 * time.Second) // Be gentle to Yahoo Finance API
	}

	log.Printf("Finished syncing Peers. Updated %d/%d active IPOs.", updatedCount, len(activeIPOs))
	return nil
}
