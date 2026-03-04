package store

import (
	"testing"
	"time"

	"backend/db"
	"backend/model"
)

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	database := db.New(":memory:")
	database.SeedDefaults()
	t.Cleanup(func() { database.Close() })
	return New(database)
}

func TestCreateAndGetShopper(t *testing.T) {
	s := setupTestStore(t)
	err := s.CreateShopper(model.Shopper{
		ShopperID: "test-1", MemberID: 99999, CustomerID: "cust-test-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	shopper, err := s.GetShopper("test-1")
	if err != nil {
		t.Fatal(err)
	}
	if shopper == nil || shopper.MemberID != 99999 {
		t.Errorf("unexpected shopper: %+v", shopper)
	}
}

func TestCreateAndGetListing(t *testing.T) {
	s := setupTestStore(t)
	now := time.Now().UTC().Format(time.RFC3339)
	id, err := s.CreateListing(model.Listing{
		DomainName:       "test.com",
		ListingStatus:    model.StatusOpen,
		ListingType:      "EXPIRY_AUCTIONS",
		AuctionTypeID:    16,
		StartTime:        now,
		EndTime:          now,
		AskingPriceUsd:   5000000,
		SellerShopperID:  "shopper-seller",
		AutoExtEnabled:   true,
		AutoExtWindowSec: 60,
		AutoExtSeconds:   300,
	})
	if err != nil {
		t.Fatal(err)
	}
	listing, err := s.GetListing(id)
	if err != nil {
		t.Fatal(err)
	}
	if listing == nil || listing.DomainName != "test.com" {
		t.Errorf("unexpected listing: %+v", listing)
	}
	if !listing.AutoExtEnabled {
		t.Error("expected autoExtEnabled=true")
	}
}

func TestCreateBidAndQuery(t *testing.T) {
	s := setupTestStore(t)
	now := time.Now().UTC().Format(time.RFC3339)
	id, _ := s.CreateListing(model.Listing{
		DomainName: "test.com", ListingStatus: model.StatusOpen,
		StartTime: now, EndTime: now, AskingPriceUsd: 5000000,
		SellerShopperID: "shopper-seller", AutoExtEnabled: true,
	})
	_ = s.CreateBid(model.Bid{
		BidID: "bid-1", ListingID: id, ShopperID: "shopper-buyer",
		BidAmountUsd: 6000000, BidType: model.BidTypeAuction,
		BidStatus: model.BidStatusActive, IsHighBid: true,
	})
	bids, err := s.GetActiveBidsForListing(id)
	if err != nil {
		t.Fatal(err)
	}
	if len(bids) != 1 {
		t.Errorf("expected 1 bid, got %d", len(bids))
	}
	hasBid, _ := s.HasShopperBidOnListing("shopper-buyer", id)
	if !hasBid {
		t.Error("expected shopper to have bid")
	}
}

func TestGetOpenListingsPastEndTime(t *testing.T) {
	s := setupTestStore(t)
	pastTime := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)
	futureTime := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	s.CreateListing(model.Listing{
		DomainName: "past.com", ListingStatus: model.StatusOpen,
		StartTime: pastTime, EndTime: pastTime, AskingPriceUsd: 5000000,
		SellerShopperID: "shopper-seller",
	})
	s.CreateListing(model.Listing{
		DomainName: "future.com", ListingStatus: model.StatusOpen,
		StartTime: pastTime, EndTime: futureTime, AskingPriceUsd: 5000000,
		SellerShopperID: "shopper-seller",
	})
	expired, err := s.GetOpenListingsPastEndTime(time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if len(expired) != 1 || expired[0].DomainName != "past.com" {
		t.Errorf("expected 1 expired listing (past.com), got %d", len(expired))
	}
}

func createTestListingForStore(t *testing.T, s *Store) int64 {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339)
	id, err := s.CreateListing(model.Listing{
		DomainName: "test.com", ListingStatus: model.StatusOpen,
		StartTime: now, EndTime: now, AskingPriceUsd: 5000000,
		SellerShopperID: "shopper-seller", AutoExtEnabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestCancelBid(t *testing.T) {
	s := setupTestStore(t)
	id := createTestListingForStore(t, s)
	s.CreateBid(model.Bid{
		BidID: "bid-cancel", ListingID: id, ShopperID: "shopper-buyer",
		BidAmountUsd: 6000000, BidType: model.BidTypeAuction,
		BidStatus: model.BidStatusActive, IsHighBid: true,
	})

	err := s.CancelBid("bid-cancel")
	if err != nil {
		t.Fatal(err)
	}

	// Verify bid is cancelled — should not appear in active bids
	bids, _ := s.GetActiveBidsForListing(id)
	if len(bids) != 0 {
		t.Errorf("expected 0 active bids after cancel, got %d", len(bids))
	}

	// Should still appear in all bids
	allBids, _ := s.GetBidsForListing(id)
	if len(allBids) != 1 || allBids[0].BidStatus != model.BidStatusCancelled {
		t.Errorf("expected 1 cancelled bid, got %+v", allBids)
	}
}

func TestSetHighBid(t *testing.T) {
	s := setupTestStore(t)
	id := createTestListingForStore(t, s)
	s.CreateBid(model.Bid{
		BidID: "bid-1", ListingID: id, ShopperID: "shopper-buyer",
		BidAmountUsd: 5000000, BidType: model.BidTypeAuction,
		BidStatus: model.BidStatusActive, IsHighBid: false,
	})

	err := s.SetHighBid("bid-1")
	if err != nil {
		t.Fatal(err)
	}

	bids, _ := s.GetBidsForListing(id)
	if len(bids) != 1 || !bids[0].IsHighBid {
		t.Errorf("expected bid to be high bid, got %+v", bids)
	}
}

func TestGetHighestAuctionBid(t *testing.T) {
	s := setupTestStore(t)
	id := createTestListingForStore(t, s)

	// No bids — should return nil
	highest, err := s.GetHighestAuctionBid(id)
	if err != nil {
		t.Fatal(err)
	}
	if highest != nil {
		t.Error("expected nil when no bids")
	}

	// Add two AUCTION bids and a PROXY bid
	s.CreateBid(model.Bid{
		BidID: "bid-a", ListingID: id, ShopperID: "shopper-buyer",
		BidAmountUsd: 5000000, BidType: model.BidTypeAuction,
		BidStatus: model.BidStatusActive,
	})
	s.CreateBid(model.Bid{
		BidID: "bid-b", ListingID: id, ShopperID: "shopper-seller",
		BidAmountUsd: 10000000, BidType: model.BidTypeAuction,
		BidStatus: model.BidStatusActive,
	})
	s.CreateBid(model.Bid{
		BidID: "proxy-1", ListingID: id, ShopperID: "shopper-buyer",
		BidAmountUsd: 50000000, BidType: model.BidTypeProxy,
		BidStatus: model.BidStatusActive,
	})

	highest, err = s.GetHighestAuctionBid(id)
	if err != nil {
		t.Fatal(err)
	}
	if highest == nil || highest.BidID != "bid-b" {
		t.Errorf("expected bid-b as highest auction, got %+v", highest)
	}
	if highest.BidAmountUsd != 10000000 {
		t.Errorf("expected amount 10000000, got %d", highest.BidAmountUsd)
	}
}

func TestGetAllActiveProxies(t *testing.T) {
	s := setupTestStore(t)
	id := createTestListingForStore(t, s)

	s.CreateShopper(model.Shopper{ShopperID: "s-a", MemberID: 20001, CustomerID: "c-a"})
	s.CreateShopper(model.Shopper{ShopperID: "s-b", MemberID: 20002, CustomerID: "c-b"})

	// Two proxies for s-a (should keep highest), one for s-b
	s.CreateBid(model.Bid{
		BidID: "proxy-a1", ListingID: id, ShopperID: "s-a",
		BidAmountUsd: 10000000, BidType: model.BidTypeProxy,
		BidStatus: model.BidStatusActive,
	})
	s.CreateBid(model.Bid{
		BidID: "proxy-a2", ListingID: id, ShopperID: "s-a",
		BidAmountUsd: 20000000, BidType: model.BidTypeProxy,
		BidStatus: model.BidStatusActive,
	})
	s.CreateBid(model.Bid{
		BidID: "proxy-b1", ListingID: id, ShopperID: "s-b",
		BidAmountUsd: 15000000, BidType: model.BidTypeProxy,
		BidStatus: model.BidStatusActive,
	})
	// Cancelled proxy should not appear
	s.CreateBid(model.Bid{
		BidID: "proxy-b2", ListingID: id, ShopperID: "s-b",
		BidAmountUsd: 50000000, BidType: model.BidTypeProxy,
		BidStatus: model.BidStatusCancelled,
	})

	proxies, err := s.GetAllActiveProxies(id)
	if err != nil {
		t.Fatal(err)
	}
	if len(proxies) != 2 {
		t.Fatalf("expected 2 shoppers with proxies, got %d", len(proxies))
	}
	if proxies["s-a"].BidAmountUsd != 20000000 {
		t.Errorf("expected s-a proxy at 20M, got %d", proxies["s-a"].BidAmountUsd)
	}
	if proxies["s-b"].BidAmountUsd != 15000000 {
		t.Errorf("expected s-b proxy at 15M, got %d", proxies["s-b"].BidAmountUsd)
	}
}

func TestGetDistinctBidderCount_OnlyAuction(t *testing.T) {
	s := setupTestStore(t)
	id := createTestListingForStore(t, s)

	// AUCTION bid from buyer
	s.CreateBid(model.Bid{
		BidID: "bid-1", ListingID: id, ShopperID: "shopper-buyer",
		BidAmountUsd: 5000000, BidType: model.BidTypeAuction,
		BidStatus: model.BidStatusActive,
	})
	// PROXY bid from same buyer — should NOT count as a distinct bidder
	s.CreateBid(model.Bid{
		BidID: "proxy-1", ListingID: id, ShopperID: "shopper-buyer",
		BidAmountUsd: 10000000, BidType: model.BidTypeProxy,
		BidStatus: model.BidStatusActive,
	})

	count, err := s.GetDistinctBidderCount(id)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("expected 1 distinct bidder (AUCTION only), got %d", count)
	}
}

func TestReset(t *testing.T) {
	s := setupTestStore(t)
	now := time.Now().UTC().Format(time.RFC3339)
	s.CreateListing(model.Listing{
		DomainName: "test.com", ListingStatus: model.StatusOpen,
		StartTime: now, EndTime: now, AskingPriceUsd: 5000000,
		SellerShopperID: "shopper-seller",
	})
	s.Reset()
	listings, _ := s.ListListings()
	if len(listings) != 0 {
		t.Errorf("expected 0 listings after reset, got %d", len(listings))
	}
	shoppers, _ := s.ListShoppers()
	if len(shoppers) != 2 {
		t.Errorf("expected 2 default shoppers after reset, got %d", len(shoppers))
	}
}
