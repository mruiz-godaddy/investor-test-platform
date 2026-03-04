package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"

	"backend/db"
	"backend/store"
)

func TestExtractShopperFromJWT(t *testing.T) {
	// Build a fake JWT with shopperId claim
	payload := map[string]interface{}{"shopperId": "903813"}
	payloadBytes, _ := json.Marshal(payload)
	encoded := base64.RawURLEncoding.EncodeToString(payloadBytes)
	fakeJWT := "Bearer header." + encoded + ".signature"

	shopperID, err := extractShopperFromJWT(fakeJWT)
	if err != nil {
		t.Fatal(err)
	}
	if shopperID != "903813" {
		t.Errorf("expected 903813, got %s", shopperID)
	}
}

func TestResolveShopperFromQueryParam(t *testing.T) {
	database := db.New(":memory:")
	database.SeedDefaults()
	defer database.Close()
	s := store.New(database)

	req, _ := http.NewRequest("GET", "/test?shopperId=shopper-buyer", nil)
	shopper, err := ResolveShopper(req, s)
	if err != nil {
		t.Fatal(err)
	}
	if shopper.ShopperID != "shopper-buyer" {
		t.Errorf("expected shopper-buyer, got %s", shopper.ShopperID)
	}
}

func TestResolveShopperAutoCreate(t *testing.T) {
	database := db.New(":memory:")
	database.SeedDefaults()
	defer database.Close()
	s := store.New(database)

	req, _ := http.NewRequest("GET", "/test?shopperId=new-shopper-123", nil)
	shopper, err := ResolveShopper(req, s)
	if err != nil {
		t.Fatal(err)
	}
	if shopper.ShopperID != "new-shopper-123" {
		t.Errorf("expected new-shopper-123, got %s", shopper.ShopperID)
	}
	if shopper.MemberID == 0 {
		t.Error("expected auto-generated member ID")
	}
}
