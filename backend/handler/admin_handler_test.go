package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"backend/bidding"
	"backend/config"
	"backend/db"
	"backend/lifecycle"
	"backend/store"
)

func setupAdminTest(t *testing.T) *AdminHandler {
	t.Helper()
	database := db.New(":memory:")
	database.SeedDefaults()
	t.Cleanup(func() { database.Close() })
	s := store.New(database)
	cfg := config.New()
	eng := bidding.NewEngine(s)
	return NewAdminHandler(s, cfg, eng, nil) // nil scenario loader
}

func TestCreateListing(t *testing.T) {
	h := setupAdminTest(t)
	body, _ := json.Marshal(map[string]interface{}{
		"domainName":      "test.com",
		"sellerShopperId": "shopper-seller",
	})
	req := httptest.NewRequest("POST", "/admin/listings", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.CreateListing(rr, req)

	if rr.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["listingStatus"] != "OPEN" {
		t.Errorf("expected OPEN, got %v", resp["listingStatus"])
	}
}

func TestUpdateStatus(t *testing.T) {
	h := setupAdminTest(t)

	// First create a listing
	body, _ := json.Marshal(map[string]interface{}{
		"domainName": "test.com", "sellerShopperId": "shopper-seller",
	})
	req := httptest.NewRequest("POST", "/admin/listings", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.CreateListing(rr, req)
	var createResp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &createResp)
	listingID := fmt.Sprintf("%.0f", createResp["listingId"].(float64))

	// Force to SOLD
	statusBody, _ := json.Marshal(map[string]string{"listingStatus": "SOLD"})
	req2 := httptest.NewRequest("PUT", "/admin/listings/"+listingID+"/status", bytes.NewReader(statusBody))
	req2 = mux.SetURLVars(req2, map[string]string{"id": listingID})
	rr2 := httptest.NewRecorder()
	h.UpdateStatus(rr2, req2)

	if rr2.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr2.Code, rr2.Body.String())
	}
	var statusResp map[string]interface{}
	json.Unmarshal(rr2.Body.Bytes(), &statusResp)
	if statusResp["listingStatus"] != "SOLD" {
		t.Errorf("expected SOLD, got %v", statusResp["listingStatus"])
	}
}

func TestUpdateConfig(t *testing.T) {
	h := setupAdminTest(t)
	body, _ := json.Marshal(map[string]interface{}{
		"autoFinalize": false, "statusTransitionDelayMs": 5000,
	})
	req := httptest.NewRequest("PUT", "/admin/config", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.UpdateConfig(rr, req)

	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["autoFinalize"] != false {
		t.Errorf("expected autoFinalize=false, got %v", resp["autoFinalize"])
	}
	if resp["statusTransitionDelayMs"].(float64) != 5000 {
		t.Errorf("expected delay=5000, got %v", resp["statusTransitionDelayMs"])
	}
}

func TestUpdateTimeOffset(t *testing.T) {
	h := setupAdminTest(t)
	defer lifecycle.Reset()

	body, _ := json.Marshal(map[string]interface{}{"offsetSeconds": 3600})
	req := httptest.NewRequest("PUT", "/admin/time", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.UpdateTime(rr, req)

	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["mode"] != "offset" {
		t.Errorf("expected mode=offset, got %v", resp["mode"])
	}
}

func TestAdminReset(t *testing.T) {
	h := setupAdminTest(t)
	req := httptest.NewRequest("POST", "/admin/reset", nil)
	rr := httptest.NewRecorder()
	h.Reset(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["status"] != "reset" {
		t.Errorf("expected status=reset, got %v", resp["status"])
	}
}

func TestListShoppers(t *testing.T) {
	h := setupAdminTest(t)
	req := httptest.NewRequest("GET", "/admin/shoppers", nil)
	rr := httptest.NewRecorder()
	h.ListShoppers(rr, req)

	var resp []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if len(resp) != 2 {
		t.Errorf("expected 2 default shoppers, got %d", len(resp))
	}
}
