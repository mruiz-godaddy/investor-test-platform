package scenario

import (
	"testing"

	"backend/bidding"
	"backend/config"
	"backend/db"
	"backend/lifecycle"
	"backend/store"
)

func setupScenarioTest(t *testing.T) *Loader {
	t.Helper()
	database := db.New(":memory:")
	database.SeedDefaults()
	t.Cleanup(func() { database.Close() })
	s := store.New(database)
	cfg := config.New()
	eng := bidding.NewEngine(s)
	lifecycle.Reset()
	return NewLoader(s, cfg, eng)
}

func TestNormalAuction(t *testing.T) {
	l := setupScenarioTest(t)
	result, err := l.Load("normal_auction")
	if err != nil {
		t.Fatal(err)
	}
	if result["scenario"] != "normal_auction" {
		t.Errorf("expected normal_auction, got %v", result["scenario"])
	}
	if result["bidsPlaced"].(int) != 2 {
		t.Errorf("expected 2 bids, got %v", result["bidsPlaced"])
	}
}

func TestRaceCondition(t *testing.T) {
	l := setupScenarioTest(t)
	result, err := l.Load("race_condition")
	if err != nil {
		t.Fatal(err)
	}
	if result["scenario"] != "race_condition" {
		t.Errorf("expected race_condition, got %v", result["scenario"])
	}
	// Verify autoFinalize is disabled
	cfgSnap := result["config"].(config.ConfigSnapshot)
	if cfgSnap.AutoFinalize {
		t.Error("race_condition should have autoFinalize=false")
	}
	if result["bidsPlaced"].(int) != 1 {
		t.Errorf("expected 1 bid, got %v", result["bidsPlaced"])
	}
}

func TestUnknownScenario(t *testing.T) {
	l := setupScenarioTest(t)
	_, err := l.Load("nonexistent")
	if err == nil {
		t.Error("expected error for unknown scenario")
	}
}

func TestProxyOutbid(t *testing.T) {
	l := setupScenarioTest(t)
	result, err := l.Load("proxy_outbid")
	if err != nil {
		t.Fatal(err)
	}
	if result["scenario"] != "proxy_outbid" {
		t.Errorf("expected proxy_outbid, got %v", result["scenario"])
	}

	listings := result["listings"].([]map[string]interface{})
	listingID := listings[0]["listingId"].(int64)
	listing, _ := l.Store.GetListing(listingID)

	// buyer-A's $20 proxy should auto-outbid buyer-B's $10 bid
	// Auto-bid at min(20, 10+5) = $15
	if listing.CurrentPriceUsd != 15_000_000 {
		t.Errorf("expected currentPrice=$15M, got %d", listing.CurrentPriceUsd)
	}
	if listing.HighestBidderShopper != "shopper-buyer-a" {
		t.Errorf("expected buyer-a as highest, got %s", listing.HighestBidderShopper)
	}
	if listing.BidsCount != 3 {
		t.Errorf("expected bidsCount=3, got %d", listing.BidsCount)
	}
}

func TestProxyStack(t *testing.T) {
	l := setupScenarioTest(t)
	result, err := l.Load("proxy_stack")
	if err != nil {
		t.Fatal(err)
	}

	listings := result["listings"].([]map[string]interface{})
	listingID := listings[0]["listingId"].(int64)
	listing, _ := l.Store.GetListing(listingID)

	// Current price should still be $5 (no new AUCTION from stack)
	if listing.CurrentPriceUsd != 5_000_000 {
		t.Errorf("expected currentPrice=$5M, got %d", listing.CurrentPriceUsd)
	}
	// Only 1 AUCTION bid
	if listing.BidsCount != 1 {
		t.Errorf("expected bidsCount=1, got %d", listing.BidsCount)
	}

	// Proxy should be at $50
	proxy, _ := l.Store.GetActiveProxyBid("shopper-buyer-a", listingID)
	if proxy == nil {
		t.Fatal("expected active proxy")
	}
	if proxy.BidAmountUsd != 50_000_000 {
		t.Errorf("expected proxy at $50M, got %d", proxy.BidAmountUsd)
	}
}

func TestProxyBurn(t *testing.T) {
	l := setupScenarioTest(t)
	result, err := l.Load("proxy_burn")
	if err != nil {
		t.Fatal(err)
	}

	listings := result["listings"].([]map[string]interface{})
	listingID := listings[0]["listingId"].(int64)
	listing, _ := l.Store.GetListing(listingID)

	if listing.CurrentPriceUsd != 15_000_000 {
		t.Errorf("expected currentPrice=$15M, got %d", listing.CurrentPriceUsd)
	}
	if listing.HighestBidderShopper != "shopper-buyer-b" {
		t.Errorf("expected buyer-b as highest, got %s", listing.HighestBidderShopper)
	}

	// Original proxy should be burned (cancelled)
	proxy, _ := l.Store.GetActiveProxyBid("shopper-buyer-a", listingID)
	if proxy != nil {
		t.Error("expected buyer-a proxy to be burned")
	}
}

func TestAllScenarios(t *testing.T) {
	names := []string{
		"normal_auction", "sniper_bid", "race_condition",
		"auto_extend_chain", "delayed_transition",
		"proxy_outbid", "proxy_stack", "proxy_burn",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			l := setupScenarioTest(t)
			result, err := l.Load(name)
			if err != nil {
				t.Fatalf("Load(%s) failed: %v", name, err)
			}
			if result["scenario"] != name {
				t.Errorf("expected scenario=%s, got %v", name, result["scenario"])
			}
			// Every scenario should have at least 1 listing
			listings := result["listings"].([]map[string]interface{})
			if len(listings) == 0 {
				t.Error("expected at least 1 listing")
			}
		})
	}
}
