package lifecycle

import (
	"context"
	"log"
	"time"

	"backend/config"
	"backend/store"
)

type Manager struct {
	Store  *store.Store
	Config *config.Config
}

func NewManager(s *store.Store, cfg *config.Config) *Manager {
	return &Manager{Store: s, Config: cfg}
}

func (m *Manager) Run(ctx context.Context) {
	interval := time.Duration(m.Config.GetFinalizerIntervalMs()) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkExpiredListings()
			// Dynamically adjust ticker interval if config changed
			newInterval := time.Duration(m.Config.GetFinalizerIntervalMs()) * time.Millisecond
			if newInterval != interval {
				ticker.Reset(newInterval)
				interval = newInterval
			}
		}
	}
}

func (m *Manager) checkExpiredListings() {
	if !m.Config.GetAutoFinalize() {
		return
	}

	now := Now()
	listings, err := m.Store.GetOpenListingsPastEndTime(now)
	if err != nil {
		log.Printf("LIFECYCLE error fetching expired listings: %v", err)
		return
	}

	for _, listing := range listings {
		endTime, err := time.Parse(time.RFC3339, listing.EndTime)
		if err != nil {
			log.Printf("LIFECYCLE error parsing endTime for listing %d: %v", listing.ListingID, err)
			continue
		}
		elapsed := now.Sub(endTime)

		// Apply status transition delay
		delayMs := m.Config.GetStatusTransitionDelayMs()
		if elapsed < time.Duration(delayMs)*time.Millisecond {
			continue
		}

		// Determine terminal status: SOLD only if bids were placed and reserve met
		if listing.BidsCount > 0 {
			m.Store.UpdateListingStatus(listing.ListingID, "SOLD", listing.CurrentPriceUsd)
			log.Printf("LIFECYCLE listing=%d OPEN→SOLD salePrice=$%.2f",
				listing.ListingID, float64(listing.CurrentPriceUsd)/1_000_000)
		} else {
			m.Store.UpdateListingStatus(listing.ListingID, "CLOSED", listing.CurrentPriceUsd)
			log.Printf("LIFECYCLE listing=%d OPEN→CLOSED",
				listing.ListingID)
		}
	}
}
