package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	return NewAppHandler(s, cfg, eng), s
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
		StartTime: time.Now().UTC().Format(time.RFC3339), EndTime: endTime,
		AskingPriceUsd: 5_000_000, SellerShopperID: "shopper-seller",
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
	if resp["code"] != "BID_IS_LESS_THAN_STARTING_AMT" {
		t.Errorf("expected BID_IS_LESS_THAN_STARTING_AMT, got %v", resp["code"])
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
		StartTime: time.Now().UTC().Format(time.RFC3339), EndTime: endTime,
		AskingPriceUsd: 5_000_000, SellerShopperID: "shopper-seller",
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
