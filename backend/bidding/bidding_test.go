package bidding

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"backend/config"
	"backend/db"
	"backend/model"
	"backend/store"
)

func setupTest(t *testing.T) (*Engine, *store.Store) {
	t.Helper()
	database := db.New(":memory:")
	database.SeedDefaults()
	t.Cleanup(func() { database.Close() })
	s := store.New(database)
	cfg := config.New()
	return NewEngine(s, cfg), s
}

func createTestListing(t *testing.T, s *store.Store, endTime time.Time) int64 {
	t.Helper()
	id, err := s.CreateListing(model.Listing{
		DomainName:       "test.com",
		ListingStatus:    model.StatusOpen,
		ListingType:      "EXPIRY_AUCTIONS",
		AuctionTypeID:    16,
		StartTime:        time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339),
		EndTime:          endTime.Format(time.RFC3339),
		AskingPriceUsd:   5_000_000, // $5
		SellerShopperID:  "shopper-seller",
		AutoExtEnabled:   true,
		AutoExtWindowSec: 60,
		AutoExtSeconds:   300,
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestGetBidIncrement(t *testing.T) {
	tests := []struct {
		price    int64
		expected int64
	}{
		{0, 5_000_000},
		{5_000_000, 5_000_000},          // $5 → $5 increment
		{499_000_000, 5_000_000},         // $499
		{500_000_000, 10_000_000},        // $500
		{999_000_000, 10_000_000},        // $999
		{1_000_000_000, 25_000_000},      // $1000
		{50_000_000_000, 1_000_000_000},  // $50,000
	}
	for _, tt := range tests {
		got := GetBidIncrement(tt.price)
		if got != tt.expected {
			t.Errorf("GetBidIncrement(%d) = %d, want %d", tt.price, got, tt.expected)
		}
	}
}

func TestPlaceBid_Success(t *testing.T) {
	engine, s := setupTest(t)
	listingID := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))

	result, err := engine.PlaceBid(BidRequest{
		ListingID:     listingID,
		ShopperID:     "shopper-buyer",
		UsdBidAmount:  5_000_000,
		IsTosAccepted: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "SUCCESS" {
		t.Errorf("expected SUCCESS, got %s", result.Status)
	}
	if !result.IsHighestBidder {
		t.Error("expected to be highest bidder")
	}

	// Verify listing updated
	listing, _ := s.GetListing(listingID)
	if listing.CurrentPriceUsd != 5_000_000 {
		t.Errorf("expected currentPrice=5000000, got %d", listing.CurrentPriceUsd)
	}
	if listing.BidsCount != 1 {
		t.Errorf("expected bidsCount=1, got %d", listing.BidsCount)
	}
}

func TestPlaceBid_TosNotAccepted(t *testing.T) {
	engine, s := setupTest(t)
	listingID := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))

	_, err := engine.PlaceBid(BidRequest{
		ListingID:     listingID,
		ShopperID:     "shopper-buyer",
		UsdBidAmount:  5_000_000,
		IsTosAccepted: false,
	})
	bidErr, ok := err.(*BidError)
	if !ok || bidErr.Code != "USER_TOS" {
		t.Errorf("expected TOS error, got %v", err)
	}
}

func TestPlaceBid_BidTooLow(t *testing.T) {
	engine, s := setupTest(t)
	listingID := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))

	_, err := engine.PlaceBid(BidRequest{
		ListingID:     listingID,
		ShopperID:     "shopper-buyer",
		UsdBidAmount:  1_000_000, // $1 < $5 asking
		IsTosAccepted: true,
	})
	bidErr, ok := err.(*BidError)
	if !ok || bidErr.Code != "BID_MIN_NOT_MET" {
		t.Errorf("expected BID_IS_LESS_THAN_STARTING_AMT, got %v", err)
	}
}

func TestPlaceBid_SellerCannotBid(t *testing.T) {
	engine, s := setupTest(t)
	listingID := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))

	_, err := engine.PlaceBid(BidRequest{
		ListingID:     listingID,
		ShopperID:     "shopper-seller",
		UsdBidAmount:  5_000_000,
		IsTosAccepted: true,
	})
	bidErr, ok := err.(*BidError)
	if !ok || bidErr.Code != "BIDDER_IS_SELLER" {
		t.Errorf("expected BIDDER_IS_SELLER, got %v", err)
	}
}

func TestPlaceBid_ListingExpired(t *testing.T) {
	engine, s := setupTest(t)
	listingID := createTestListing(t, s, time.Now().UTC().Add(-1*time.Minute))

	_, err := engine.PlaceBid(BidRequest{
		ListingID:     listingID,
		ShopperID:     "shopper-buyer",
		UsdBidAmount:  5_000_000,
		IsTosAccepted: true,
	})
	bidErr, ok := err.(*BidError)
	if !ok || bidErr.Code != "LISTING_NOT_OPEN" {
		t.Errorf("expected LISTING_CLOSED, got %v", err)
	}
}

func TestPlaceBid_FirstBidCreatesProxy(t *testing.T) {
	engine, s := setupTest(t)
	listingID := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))

	// Bid above asking price should create PROXY + AUCTION pair
	result, err := engine.PlaceBid(BidRequest{
		ListingID:     listingID,
		ShopperID:     "shopper-buyer",
		UsdBidAmount:  20_000_000, // $20, asking is $5
		IsTosAccepted: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "SUCCESS" {
		t.Errorf("expected SUCCESS, got %s", result.Status)
	}
	if !result.IsHighestBidder {
		t.Error("expected to be highest bidder")
	}

	// Should have PROXY bid
	proxy, _ := s.GetActiveProxyBid("shopper-buyer", listingID)
	if proxy == nil {
		t.Fatal("expected proxy bid to be created")
	}
	if proxy.BidAmountUsd != 20_000_000 {
		t.Errorf("expected proxy at $20, got %d", proxy.BidAmountUsd)
	}

	// Listing currentPrice should be asking price ($5), not the bid amount
	listing, _ := s.GetListing(listingID)
	if listing.CurrentPriceUsd != 5_000_000 {
		t.Errorf("expected currentPrice=$5, got %d", listing.CurrentPriceUsd)
	}
	if listing.BidsCount != 1 {
		t.Errorf("expected bidsCount=1, got %d", listing.BidsCount)
	}
}

func TestPlaceBid_ExactAskingNoProxy(t *testing.T) {
	engine, s := setupTest(t)
	listingID := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))

	// Bid at exactly asking price — no proxy created
	result, err := engine.PlaceBid(BidRequest{
		ListingID:     listingID,
		ShopperID:     "shopper-buyer",
		UsdBidAmount:  5_000_000, // exact asking
		IsTosAccepted: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsHighestBidder {
		t.Error("expected highest bidder")
	}

	proxy, _ := s.GetActiveProxyBid("shopper-buyer", listingID)
	if proxy != nil {
		t.Error("expected NO proxy for exact asking price bid")
	}
}

func TestPlaceBid_ProxyAutoOutbids(t *testing.T) {
	engine, s := setupTest(t)
	listingID := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))

	s.CreateShopper(model.Shopper{ShopperID: "buyer-b", MemberID: 10003, CustomerID: "cust-b"})

	// buyer places $20 proxy
	_, err := engine.PlaceBid(BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer",
		UsdBidAmount: 20_000_000, IsTosAccepted: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// buyer-b bids $10 — should be auto-outbid by proxy
	result, err := engine.PlaceBid(BidRequest{
		ListingID: listingID, ShopperID: "buyer-b",
		UsdBidAmount: 10_000_000, IsTosAccepted: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsHighestBidder {
		t.Error("expected buyer-b NOT to be highest bidder (proxy outbids)")
	}

	// Verify listing state
	listing, _ := s.GetListing(listingID)
	if listing.HighestBidderShopper != "shopper-buyer" {
		t.Errorf("expected shopper-buyer as highest bidder, got %s", listing.HighestBidderShopper)
	}
	// Proxy should auto-bid at min(20, 10+increment). increment for $10 is $5, so child at $15
	if listing.CurrentPriceUsd != 15_000_000 {
		t.Errorf("expected currentPrice=$15, got %d", listing.CurrentPriceUsd)
	}
	// 3 AUCTION bids: initial $5, buyer-b $10, proxy child $15
	if listing.BidsCount != 3 {
		t.Errorf("expected bidsCount=3, got %d", listing.BidsCount)
	}
}

func TestPlaceBid_ProxyBurned(t *testing.T) {
	engine, s := setupTest(t)
	listingID := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))

	s.CreateShopper(model.Shopper{ShopperID: "buyer-b", MemberID: 10003, CustomerID: "cust-b"})

	// buyer places $12 proxy
	_, err := engine.PlaceBid(BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer",
		UsdBidAmount: 12_000_000, IsTosAccepted: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// buyer-b bids $15 — exceeds $12 proxy, burns it
	result, err := engine.PlaceBid(BidRequest{
		ListingID: listingID, ShopperID: "buyer-b",
		UsdBidAmount: 15_000_000, IsTosAccepted: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsHighestBidder {
		t.Error("expected buyer-b to be highest bidder after burning proxy")
	}

	listing, _ := s.GetListing(listingID)
	if listing.HighestBidderShopper != "buyer-b" {
		t.Errorf("expected buyer-b as highest, got %s", listing.HighestBidderShopper)
	}
	if listing.CurrentPriceUsd != 15_000_000 {
		t.Errorf("expected currentPrice=$15, got %d", listing.CurrentPriceUsd)
	}

	// Original proxy should be cancelled
	proxy, _ := s.GetActiveProxyBid("shopper-buyer", listingID)
	if proxy != nil {
		t.Error("expected proxy to be cancelled (burned)")
	}
}

func TestPlaceBid_ProxyStack(t *testing.T) {
	engine, s := setupTest(t)
	listingID := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))

	// buyer places $20 proxy
	_, err := engine.PlaceBid(BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer",
		UsdBidAmount: 20_000_000, IsTosAccepted: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Same buyer raises to $50 — stack proxy
	result, err := engine.PlaceBid(BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer",
		UsdBidAmount: 50_000_000, IsTosAccepted: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsHighestBidder {
		t.Error("expected highest bidder after proxy stack")
	}

	// New proxy at $50
	proxy, _ := s.GetActiveProxyBid("shopper-buyer", listingID)
	if proxy == nil {
		t.Fatal("expected active proxy")
	}
	if proxy.BidAmountUsd != 50_000_000 {
		t.Errorf("expected proxy at $50, got %d", proxy.BidAmountUsd)
	}

	// bidsCount should still be 1 (only the initial AUCTION, proxy stack doesn't add AUCTION)
	listing, _ := s.GetListing(listingID)
	if listing.BidsCount != 1 {
		t.Errorf("expected bidsCount=1 (no new AUCTION from stack), got %d", listing.BidsCount)
	}
}

func TestPlaceBid_SameAmountAsProxy(t *testing.T) {
	engine, s := setupTest(t)
	listingID := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))

	// buyer places $20 proxy
	engine.PlaceBid(BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer",
		UsdBidAmount: 20_000_000, IsTosAccepted: true,
	})

	// Same buyer bids $20 again — should fail (same amount as existing proxy)
	_, err := engine.PlaceBid(BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer",
		UsdBidAmount: 20_000_000, IsTosAccepted: true,
	})
	bidErr, ok := err.(*BidError)
	if !ok || bidErr.Code != "BID_MIN_NOT_MET" {
		t.Errorf("expected BID_IS_LESS_THAN_STARTING_AMT, got %v", err)
	}
}

func TestPlaceBid_BidCountOnlyCountsAuction(t *testing.T) {
	engine, s := setupTest(t)
	listingID := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))

	// Place bid above asking → creates PROXY + AUCTION
	engine.PlaceBid(BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer",
		UsdBidAmount: 20_000_000, IsTosAccepted: true,
	})

	listing, _ := s.GetListing(listingID)
	// Only 1 AUCTION bid should be counted (not the PROXY)
	if listing.BidsCount != 1 {
		t.Errorf("expected bidsCount=1 (AUCTION only), got %d", listing.BidsCount)
	}
}

func TestPlaceBid_IsHighestBidderFalseWhenOutbid(t *testing.T) {
	engine, s := setupTest(t)
	listingID := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))

	s.CreateShopper(model.Shopper{ShopperID: "buyer-b", MemberID: 10003, CustomerID: "cust-b"})

	// buyer places $20 proxy
	engine.PlaceBid(BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer",
		UsdBidAmount: 20_000_000, IsTosAccepted: true,
	})

	// buyer-b bids $10 (minimum valid: $5 + $5 increment) — proxy auto-outbids
	result, err := engine.PlaceBid(BidRequest{
		ListingID: listingID, ShopperID: "buyer-b",
		UsdBidAmount: 10_000_000, IsTosAccepted: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsHighestBidder {
		t.Error("expected IsHighestBidder=false when proxy auto-outbids")
	}
}

func TestConcurrent_SameListingBidsAreSerialized(t *testing.T) {
	engine, s := setupTest(t)
	listingID := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))

	// Place initial bid so there's a baseline
	_, err := engine.PlaceBid(BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer",
		UsdBidAmount: 5_000_000, IsTosAccepted: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Fire 10 concurrent bids from different shoppers on the same listing.
	// With the per-listing lock, exactly one should succeed per price level;
	// without the lock, multiple goroutines could read the same state and
	// produce duplicate bids at the same amount.
	const numBidders = 10
	shopperIDs := make([]string, numBidders)
	for i := 0; i < numBidders; i++ {
		shopperIDs[i] = fmt.Sprintf("concurrent-buyer-%d", i)
		s.CreateShopper(model.Shopper{
			ShopperID:  shopperIDs[i],
			MemberID:   int64(20000 + i),
			CustomerID: fmt.Sprintf("cust-concurrent-%d", i),
		})
	}

	var wg sync.WaitGroup
	results := make([]*BidResult, numBidders)
	errors := make([]error, numBidders)

	for i := 0; i < numBidders; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// All bid $10; the first serialized bid wins, the rest get BID_MIN_NOT_MET
			// because after the first $10 bid, the next minimum is $10 + $5 = $15.
			results[idx], errors[idx] = engine.PlaceBid(BidRequest{
				ListingID:    listingID,
				ShopperID:    shopperIDs[idx],
				UsdBidAmount: 10_000_000,
				IsTosAccepted: true,
			})
		}(i)
	}
	wg.Wait()

	successCount := 0
	bidTooLowCount := 0
	for i := 0; i < numBidders; i++ {
		if errors[i] == nil {
			successCount++
		} else if bidErr, ok := errors[i].(*BidError); ok && bidErr.Code == "BID_MIN_NOT_MET" {
			bidTooLowCount++
		}
	}

	if successCount != 1 {
		t.Errorf("expected exactly 1 successful bid at $10, got %d", successCount)
	}
	if successCount+bidTooLowCount != numBidders {
		t.Errorf("expected all bids to either succeed or get BID_MIN_NOT_MET, got %d success + %d too-low out of %d",
			successCount, bidTooLowCount, numBidders)
	}
}

func TestConcurrent_DifferentListingsAreIndependent(t *testing.T) {
	engine, s := setupTest(t)
	listingA := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))
	listingB := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))

	// Place concurrent bids on two different listings — both should succeed
	// without blocking each other.
	var wg sync.WaitGroup
	var resultA, resultB *BidResult
	var errA, errB error

	wg.Add(2)
	go func() {
		defer wg.Done()
		resultA, errA = engine.PlaceBid(BidRequest{
			ListingID: listingA, ShopperID: "shopper-buyer",
			UsdBidAmount: 5_000_000, IsTosAccepted: true,
		})
	}()
	go func() {
		defer wg.Done()
		resultB, errB = engine.PlaceBid(BidRequest{
			ListingID: listingB, ShopperID: "shopper-buyer",
			UsdBidAmount: 5_000_000, IsTosAccepted: true,
		})
	}()
	wg.Wait()

	if errA != nil {
		t.Errorf("listing A bid failed: %v", errA)
	}
	if errB != nil {
		t.Errorf("listing B bid failed: %v", errB)
	}
	if resultA != nil && resultA.Status != "SUCCESS" {
		t.Errorf("listing A expected SUCCESS, got %s", resultA.Status)
	}
	if resultB != nil && resultB.Status != "SUCCESS" {
		t.Errorf("listing B expected SUCCESS, got %s", resultB.Status)
	}
}

func TestConcurrent_SniperBidsSerialized(t *testing.T) {
	engine, s := setupTest(t)
	listingID := createTestListing(t, s, time.Now().UTC().Add(10*time.Minute))

	// Place initial bid
	_, err := engine.PlaceSniperBid(listingID, "shopper-buyer", 5_000_000)
	if err != nil {
		t.Fatal(err)
	}

	// Fire concurrent sniper bids at the same price — only one should succeed
	const numBidders = 5
	shopperIDs := make([]string, numBidders)
	for i := 0; i < numBidders; i++ {
		shopperIDs[i] = fmt.Sprintf("sniper-%d", i)
		s.CreateShopper(model.Shopper{
			ShopperID:  shopperIDs[i],
			MemberID:   int64(30000 + i),
			CustomerID: fmt.Sprintf("cust-sniper-%d", i),
		})
	}

	var wg sync.WaitGroup
	errors := make([]error, numBidders)

	for i := 0; i < numBidders; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, errors[idx] = engine.PlaceSniperBid(listingID, shopperIDs[idx], 10_000_000)
		}(i)
	}
	wg.Wait()

	successCount := 0
	for i := 0; i < numBidders; i++ {
		if errors[i] == nil {
			successCount++
		}
	}

	if successCount != 1 {
		t.Errorf("expected exactly 1 successful sniper bid at $10, got %d", successCount)
	}
}

func TestGetLock_ReturnsSameMutexForSameListing(t *testing.T) {
	engine, _ := setupTest(t)

	lock1 := engine.getLock(123)
	lock2 := engine.getLock(123)
	lock3 := engine.getLock(456)

	if lock1 != lock2 {
		t.Error("expected same mutex for same listing ID")
	}
	if lock1 == lock3 {
		t.Error("expected different mutex for different listing ID")
	}
}

