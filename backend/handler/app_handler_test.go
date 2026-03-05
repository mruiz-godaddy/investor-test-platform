package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"backend/bidding"
	"backend/config"
	"backend/db"
	"backend/model"
	"backend/store"
)

func setupAppHandlerTest(t *testing.T) (*AppHandler, *store.Store) {
	t.Helper()
	database := db.New(":memory:")
	database.SeedDefaults()
	t.Cleanup(func() { database.Close() })
	s := store.New(database)
	cfg := config.New()
	eng := bidding.NewEngine(s)
	return NewAppHandler(s, cfg, eng, "", ""), s
}

// setupWithUpstream creates an AppHandler pointed at a fake upstream server.
// The upstream handler receives requests forwarded by the mock.
func setupWithUpstream(t *testing.T, upstreamHandler http.HandlerFunc) (*AppHandler, *store.Store) {
	t.Helper()
	database := db.New(":memory:")
	database.SeedDefaults()
	t.Cleanup(func() { database.Close() })
	s := store.New(database)
	cfg := config.New()
	eng := bidding.NewEngine(s)

	upstream := httptest.NewServer(upstreamHandler)
	t.Cleanup(upstream.Close)

	return NewAppHandler(s, cfg, eng, upstream.URL, upstream.URL), s
}

func TestPlaceBidHandler_Success(t *testing.T) {
	h, s := setupAppHandlerTest(t)

	// Create a listing
	endTime := time.Now().UTC().Add(10 * time.Minute).Format(time.RFC3339)
	id, _ := s.CreateListing(model.Listing{
		DomainName: "test.com", ListingStatus: model.StatusOpen,
		ListingType: "EXPIRY_AUCTIONS", AuctionTypeID: 16,
		StartTime: time.Now().UTC().Format(time.RFC3339), EndTime: endTime,
		AskingPriceUsd: 5_000_000, SellerShopperID: "shopper-seller",
		AutoExtEnabled: true, AutoExtWindowSec: 60, AutoExtSeconds: 300,
	})

	body, _ := json.Marshal(map[string]interface{}{
		"usdBidAmount":  5_000_000,
		"isTosAccepted": true,
	})

	req := httptest.NewRequest("POST", fmt.Sprintf("/v1/aftermarket/domains/listings/%d/bids?shopperId=shopper-buyer", id), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"listingId": fmt.Sprintf("%d", id)})

	rr := httptest.NewRecorder()
	h.PlaceBid(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)

	// listingId must be a number
	if _, ok := resp["listingId"].(float64); !ok {
		t.Error("listingId should be a number")
	}
	if resp["status"] != "SUCCESS" {
		t.Errorf("expected SUCCESS, got %v", resp["status"])
	}
}

func TestPlaceBidHandler_BidTooLow(t *testing.T) {
	h, s := setupAppHandlerTest(t)

	endTime := time.Now().UTC().Add(10 * time.Minute).Format(time.RFC3339)
	id, _ := s.CreateListing(model.Listing{
		DomainName: "test.com", ListingStatus: model.StatusOpen,
		ListingType: "EXPIRY_AUCTIONS", AuctionTypeID: 16,
		StartTime: time.Now().UTC().Format(time.RFC3339), EndTime: endTime,
		AskingPriceUsd: 5_000_000, SellerShopperID: "shopper-seller",
		AutoExtEnabled: true, AutoExtWindowSec: 60, AutoExtSeconds: 300,
	})

	body, _ := json.Marshal(map[string]interface{}{
		"usdBidAmount":  1_000_000, // $1 < $5 asking
		"isTosAccepted": true,
	})

	req := httptest.NewRequest("POST", "/test?shopperId=shopper-buyer", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"listingId": fmt.Sprintf("%d", id)})

	rr := httptest.NewRecorder()
	h.PlaceBid(rr, req)

	if rr.Code != 422 {
		t.Fatalf("expected 422, got %d", rr.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["code"] != "BID_MIN_NOT_MET" {
		t.Errorf("expected BID_MIN_NOT_MET, got %v", resp["code"])
	}
	// Verify error is NOT nested under "error" key
	if _, hasError := resp["error"]; hasError {
		t.Error("error should NOT be nested under 'error' key")
	}
}

func TestPlaceBidHandler_MissingShopper(t *testing.T) {
	h, s := setupAppHandlerTest(t)

	endTime := time.Now().UTC().Add(10 * time.Minute).Format(time.RFC3339)
	id, _ := s.CreateListing(model.Listing{
		DomainName: "test.com", ListingStatus: model.StatusOpen,
		ListingType: "EXPIRY_AUCTIONS", AuctionTypeID: 16,
		StartTime: time.Now().UTC().Format(time.RFC3339), EndTime: endTime,
		AskingPriceUsd: 5_000_000, SellerShopperID: "shopper-seller",
		AutoExtEnabled: true, AutoExtWindowSec: 60, AutoExtSeconds: 300,
	})

	body, _ := json.Marshal(map[string]interface{}{
		"usdBidAmount": 5_000_000, "isTosAccepted": true,
	})

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body)) // No shopperId
	req = mux.SetURLVars(req, map[string]string{"listingId": fmt.Sprintf("%d", id)})

	rr := httptest.NewRecorder()
	h.PlaceBid(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// --- Forwarding / upstream proxy tests ---

func TestSearchListings_ForwardsToUpstreamWhenNoLocalResults(t *testing.T) {
	var gotPath string
	upstreamCalled := false

	h, _ := setupWithUpstream(t, func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		gotPath = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"auction_id": 999, "fqdn": "upstream.com"},
			},
		})
	})

	// Search for a domain that doesn't exist locally
	req := httptest.NewRequest("GET", "/v4/aftermarket/find/auction/recommend?query=nonexistent.xyz&paginationSize=100", nil)
	rr := httptest.NewRecorder()
	h.SearchListings(rr, req)

	if !upstreamCalled {
		t.Fatal("expected upstream to be called when no local results")
	}
	if gotPath != "/v4/aftermarket/find/auction/recommend?query=nonexistent.xyz&paginationSize=100" {
		t.Errorf("upstream received wrong path: %s", gotPath)
	}
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	// Verify upstream response was proxied through
	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	results, ok := resp["results"].([]interface{})
	if !ok || len(results) != 1 {
		t.Fatalf("expected 1 upstream result, got %v", resp["results"])
	}
}

func TestSearchListings_EmptyQueryForwardsToUpstream(t *testing.T) {
	upstreamCalled := false

	h, _ := setupWithUpstream(t, func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{
				{"auction_id": 1, "fqdn": "radar-result.com"},
			},
		})
	})

	// Radar-style request: no query, just filters
	req := httptest.NewRequest("GET", "/v4/aftermarket/find/auction/recommend?typeIncludeList=14,16,38&paginationSize=500", nil)
	rr := httptest.NewRecorder()
	h.SearchListings(rr, req)

	if !upstreamCalled {
		t.Fatal("expected upstream to be called for empty-query (Radar) request")
	}
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestSearchListings_LocalResultsNotForwarded(t *testing.T) {
	upstreamCalled := false

	h, s := setupWithUpstream(t, func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": []map[string]interface{}{},
		})
	})

	// Create an OPEN listing that matches search
	endTime := time.Now().UTC().Add(10 * time.Minute).Format(time.RFC3339)
	s.CreateListing(model.Listing{
		DomainName: "hive95.cc", ListingStatus: model.StatusOpen,
		ListingType: "EXPIRY_AUCTIONS", AuctionTypeID: 16,
		StartTime: time.Now().UTC().Format(time.RFC3339), EndTime: endTime,
		AskingPriceUsd: 5_000_000, SellerShopperID: "shopper-seller-1",
	})

	req := httptest.NewRequest("GET", "/v4/aftermarket/find/auction/recommend?query=hive95", nil)
	rr := httptest.NewRecorder()
	h.SearchListings(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	results := resp["results"].([]interface{})
	if len(results) < 1 {
		t.Fatal("expected at least 1 local result")
	}
	first := results[0].(map[string]interface{})
	if first["fqdn"] != "hive95.cc" {
		t.Errorf("expected fqdn=hive95.cc, got %v", first["fqdn"])
	}

	// Upstream should have been called for merging (not proxying)
	if !upstreamCalled {
		t.Error("expected upstream to be called for merging when local results exist")
	}
}

func TestSearchListings_NoUpstreamReturnsEmpty(t *testing.T) {
	h, _ := setupAppHandlerTest(t) // no upstream configured

	req := httptest.NewRequest("GET", "/v4/aftermarket/find/auction/recommend?query=nonexistent.xyz", nil)
	rr := httptest.NewRecorder()
	h.SearchListings(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	results := resp["results"].([]interface{})
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

func TestGetListing_ForwardsToUpstreamWhenNotFound(t *testing.T) {
	upstreamCalled := false

	h, _ := setupWithUpstream(t, func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"listingId":  99999,
			"domainName": "upstream-listing.com",
		})
	})

	req := httptest.NewRequest("GET", "/v1/aftermarket/domains/listings/99999?shopperId=shopper-buyer-1", nil)
	req = mux.SetURLVars(req, map[string]string{"listingId": "99999"})
	rr := httptest.NewRecorder()
	h.GetListing(rr, req)

	if !upstreamCalled {
		t.Fatal("expected upstream to be called when listing not found locally")
	}
	if rr.Code != 200 {
		t.Fatalf("expected 200 from upstream, got %d", rr.Code)
	}
}

func TestGetListing_NoUpstreamReturns404(t *testing.T) {
	h, _ := setupAppHandlerTest(t) // no upstream

	req := httptest.NewRequest("GET", "/v1/aftermarket/domains/listings/99999?shopperId=shopper-buyer-1", nil)
	req = mux.SetURLVars(req, map[string]string{"listingId": "99999"})
	rr := httptest.NewRecorder()
	h.GetListing(rr, req)

	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestGetBiddingListings_ForwardsWhenNoLocalBids(t *testing.T) {
	upstreamCalled := false

	h, _ := setupWithUpstream(t, func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"listings": []map[string]interface{}{
				{"listingId": 555, "domainName": "upstream-bid.com"},
			},
			"viewType": "SNAPSHOT",
		})
	})

	// Shopper has no bids on any local listing
	req := httptest.NewRequest("GET", "/v1/aftermarket/domains/bidding?shopperId=shopper-buyer-1", nil)
	rr := httptest.NewRecorder()
	h.GetBiddingListings(rr, req)

	if !upstreamCalled {
		t.Fatal("expected upstream to be called when no local bidding listings")
	}
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetWonListings_ForwardsWhenNoLocalWins(t *testing.T) {
	upstreamCalled := false

	h, _ := setupWithUpstream(t, func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"listings": []map[string]interface{}{},
			"viewType": "SNAPSHOT",
		})
	})

	req := httptest.NewRequest("GET", "/v1/aftermarket/domains/won?shopperId=shopper-buyer-1", nil)
	rr := httptest.NewRecorder()
	h.GetWonListings(rr, req)

	if !upstreamCalled {
		t.Fatal("expected upstream to be called when no local won listings")
	}
}

func TestGetLostListings_ForwardsWhenNoLocalLosses(t *testing.T) {
	upstreamCalled := false

	h, _ := setupWithUpstream(t, func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"listings": []map[string]interface{}{},
			"viewType": "SNAPSHOT",
		})
	})

	req := httptest.NewRequest("GET", "/v1/aftermarket/domains/didNotWin?shopperId=shopper-buyer-1", nil)
	rr := httptest.NewRecorder()
	h.GetLostListings(rr, req)

	if !upstreamCalled {
		t.Fatal("expected upstream to be called when no local lost listings")
	}
}
