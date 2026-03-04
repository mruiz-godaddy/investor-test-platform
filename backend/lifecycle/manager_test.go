package lifecycle

import (
	"context"
	"testing"
	"time"

	"backend/config"
	"backend/db"
	"backend/model"
	"backend/store"
)

func setupManagerTest(t *testing.T) (*Manager, *store.Store) {
	t.Helper()
	database := db.New(":memory:")
	database.SeedDefaults()
	t.Cleanup(func() { database.Close() })
	s := store.New(database)
	cfg := config.New()
	cfg.SetFinalizerIntervalMs(50) // Fast interval for tests
	return NewManager(s, cfg), s
}

func TestAutoFinalize_SOLD(t *testing.T) {
	m, s := setupManagerTest(t)
	Reset() // Ensure realtime clock
	defer Reset()

	// Create listing that ended 1 minute ago with a bid
	pastEnd := time.Now().UTC().Add(-1 * time.Minute).Format(time.RFC3339)
	id, _ := s.CreateListing(model.Listing{
		DomainName:      "test.com",
		ListingStatus:   model.StatusOpen,
		StartTime:       pastEnd,
		EndTime:         pastEnd,
		AskingPriceUsd:  5_000_000,
		BidsCount:       1,
		SellerShopperID: "shopper-seller",
	})

	// Run one check
	m.checkExpiredListings()

	listing, _ := s.GetListing(id)
	if listing.ListingStatus != model.StatusSold {
		t.Errorf("expected SOLD, got %s", listing.ListingStatus)
	}
}

func TestAutoFinalize_CLOSED_NoBids(t *testing.T) {
	m, s := setupManagerTest(t)
	Reset()
	defer Reset()

	pastEnd := time.Now().UTC().Add(-1 * time.Minute).Format(time.RFC3339)
	id, _ := s.CreateListing(model.Listing{
		DomainName:      "test.com",
		ListingStatus:   model.StatusOpen,
		StartTime:       pastEnd,
		EndTime:         pastEnd,
		AskingPriceUsd:  5_000_000,
		CurrentPriceUsd: 0, // No bids
		SellerShopperID: "shopper-seller",
	})

	m.checkExpiredListings()

	listing, _ := s.GetListing(id)
	if listing.ListingStatus != model.StatusClosed {
		t.Errorf("expected CLOSED, got %s", listing.ListingStatus)
	}
}

func TestAutoFinalize_CLOSED_ReserveNotMet(t *testing.T) {
	m, s := setupManagerTest(t)
	Reset()
	defer Reset()

	pastEnd := time.Now().UTC().Add(-1 * time.Minute).Format(time.RFC3339)
	id, _ := s.CreateListing(model.Listing{
		DomainName:      "test.com",
		ListingStatus:   model.StatusOpen,
		StartTime:       pastEnd,
		EndTime:         pastEnd,
		AskingPriceUsd:  5_000_000,
		BidsCount:       1,
		ReservePriceUsd: 50_000_000, // Reserve = $50, current = $10 (asking)
		SellerShopperID: "shopper-seller",
	})

	m.checkExpiredListings()

	listing, _ := s.GetListing(id)
	if listing.ListingStatus != model.StatusClosed {
		t.Errorf("expected CLOSED (reserve not met), got %s", listing.ListingStatus)
	}
}

func TestAutoFinalize_Disabled(t *testing.T) {
	m, s := setupManagerTest(t)
	m.Config.SetAutoFinalize(false)
	Reset()
	defer Reset()

	pastEnd := time.Now().UTC().Add(-1 * time.Minute).Format(time.RFC3339)
	id, _ := s.CreateListing(model.Listing{
		DomainName:      "test.com",
		ListingStatus:   model.StatusOpen,
		StartTime:       pastEnd,
		EndTime:         pastEnd,
		AskingPriceUsd:  5_000_000,
		BidsCount:       1,
		SellerShopperID: "shopper-seller",
	})

	m.checkExpiredListings()

	listing, _ := s.GetListing(id)
	if listing.ListingStatus != model.StatusOpen {
		t.Errorf("expected OPEN (autoFinalize disabled), got %s", listing.ListingStatus)
	}
}

func TestTransitionDelay(t *testing.T) {
	m, s := setupManagerTest(t)
	m.Config.SetStatusTransitionDelayMs(5000) // 5 second delay
	Reset()
	defer Reset()

	// Listing ended 2 seconds ago (less than 5s delay)
	recentEnd := time.Now().UTC().Add(-2 * time.Second).Format(time.RFC3339)
	id, _ := s.CreateListing(model.Listing{
		DomainName:      "test.com",
		ListingStatus:   model.StatusOpen,
		StartTime:       recentEnd,
		EndTime:         recentEnd,
		AskingPriceUsd:  5_000_000,
		BidsCount:       1,
		SellerShopperID: "shopper-seller",
	})

	m.checkExpiredListings()

	listing, _ := s.GetListing(id)
	if listing.ListingStatus != model.StatusOpen {
		t.Errorf("expected OPEN (delay not elapsed), got %s", listing.ListingStatus)
	}
}

func TestRunStopsOnCancel(t *testing.T) {
	m, _ := setupManagerTest(t)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		m.Run(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
		// OK — goroutine exited
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after context cancel")
	}
}
