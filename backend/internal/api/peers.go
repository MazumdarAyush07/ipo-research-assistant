package api

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/MazumdarAyush07/ipo-research/internal/models"
	"github.com/MazumdarAyush07/ipo-research/internal/services"
	"github.com/gofiber/fiber/v2"
)

// GetPeers handles GET /api/ipos/:id/peers
func (h *IPOHandler) GetPeers(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid ipo id"})
	}
	ipoID := int64(id)

	// Fetch IPO
	ipo, err := h.Queries.GetIPO(c.Context(), ipoID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "IPO not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Fetch financials to get IPO's computed PE (we need Issue Price which is in IPO table? Wait, IPO table doesn't have Issue Price yet. Let's just mock IPO PE for now if we don't have it)
	// We'll compute it from Valuation table if it exists, or just pass 0.0
	// For this phase, we'll focus on fetching peers.
	
	// Check if peers exist in DB
	dbPeers, err := h.Queries.GetPeerCompaniesByIPO(c.Context(), ipoID)
	if err != nil && err != sql.ErrNoRows {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var peerData []services.PeerData
	
	if len(dbPeers) == 0 {
		// If no peers in DB, fetch via PeerService using IPO Sector
		// Wait, IPO sector is NULL initially. Let's fallback to "Healthcare" for testing if it's NULL
		sector := ipo.Sector.String
		if !ipo.Sector.Valid || sector == "" {
			// Try to infer from name for testing, or just use a default
			sector = "Healthcare"
		}

		if h.PeerService != nil {
			fetchedPeers, err := h.PeerService.FetchPeersForSector(c.Context(), sector)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}

			// Store in DB
			for _, p := range fetchedPeers {
				_, _ = h.Queries.InsertPeerCompany(c.Context(), models.InsertPeerCompanyParams{
					IpoID:     ipoID,
					Name:      p.Name,
					Ticker:    sql.NullString{String: p.Ticker, Valid: true},
					Pe:        sql.NullString{String: fmt.Sprintf("%f", p.PE), Valid: true},
					Pb:        sql.NullString{String: fmt.Sprintf("%f", p.PB), Valid: true},
					MarketCap: sql.NullString{String: fmt.Sprintf("%f", p.MarketCap), Valid: true},
				})
				
				peerData = append(peerData, p)
			}
		}
	} else {
		// Convert DB peers to service DTO
		for _, p := range dbPeers {
			pe, _ := strconv.ParseFloat(p.Pe.String, 64)
			pb, _ := strconv.ParseFloat(p.Pb.String, 64)
			cap, _ := strconv.ParseFloat(p.MarketCap.String, 64)

			peerData = append(peerData, services.PeerData{
				Ticker:    p.Ticker.String,
				Name:      p.Name,
				PE:        pe,
				PB:        pb,
				MarketCap: cap,
			})
		}
	}

	// Compute Comparison
	ipoPE := 0.0

	// Fetch dynamic PE from the valuation table
	val, err := h.Queries.GetValuationByIPO(c.Context(), ipoID)
	if err == nil && val.PeRatio.Valid {
		parsedPE, err := strconv.ParseFloat(val.PeRatio.String, 64)
		if err == nil {
			ipoPE = parsedPE
		}
	}

	comparison := services.CompareIPOToPeers(ipoPE, peerData)

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"ipo_pe": ipoPE,
			"peers": peerData,
			"comparison": comparison,
		},
	})
}
