package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"backend/bidding"
	"backend/config"
	"backend/lifecycle"
	"backend/model"
	"backend/store"
)

// ScenarioLoader interface decouples admin handler from scenario package.
type ScenarioLoader interface {
	Load(name string) (map[string]interface{}, error)
}

type AdminHandler struct {
	Store    *store.Store
	Config   *config.Config
	Engine   *bidding.Engine
	Scenario ScenarioLoader
}

func NewAdminHandler(s *store.Store, cfg *config.Config, eng *bidding.Engine, sc ScenarioLoader) *AdminHandler {
	return &AdminHandler{Store: s, Config: cfg, Engine: eng, Scenario: sc}
}

func (h *AdminHandler) buildAdminListingJSON(listing *model.Listing) map[string]interface{} {
	bids, _ := h.Store.GetBidsForListing(listing.ListingID)
	bidHistory := make([]map[string]interface{}, 0, len(bids))
	for _, bid := range bids {
		bidHistory = append(bidHistory, map[string]interface{}{
			"bidId":        bid.BidID,
			"shopperId":    bid.ShopperID,
			"bidAmountUsd": bid.BidAmountUsd,
			"bidType":      bid.BidType,
			"bidStatus":    bid.BidStatus,
			"isHighBid":    bid.IsHighBid,
			"parentBidId":  bid.ParentBidID,
			"createdAt":    bid.CreatedAt,
		})
	}

	return map[string]interface{}{
		"listingId":            listing.ListingID,
		"domainName":           listing.DomainName,
		"listingStatus":        listing.ListingStatus,
		"listingType":          listing.ListingType,
		"auctionTypeId":        listing.AuctionTypeID,
		"startTime":            listing.StartTime,
		"endTime":              listing.EndTime,
		"askingPriceUsd":       listing.AskingPriceUsd,
		"currentPriceUsd":      listing.CurrentPriceUsd,
		"salePriceUsd":         listing.SalePriceUsd,
		"reservePriceUsd":      listing.ReservePriceUsd,
		"nextBidPriceUsd":      listing.NextBidPriceUsd,
		"biddersCount":         listing.BiddersCount,
		"bidsCount":            listing.BidsCount,
		"isReserveMet":         listing.IsReserveMet,
		"isAutoExtended":       listing.IsAutoExtended,
		"sellerShopperId":      listing.SellerShopperID,
		"highestBidderShopper": listing.HighestBidderShopper,
		"autoExtEnabled":       listing.AutoExtEnabled,
		"autoExtWindowSec":     listing.AutoExtWindowSec,
		"autoExtSeconds":       listing.AutoExtSeconds,
		"createdAt":            listing.CreatedAt,
		"bidHistory":           bidHistory,
	}
}

// --- 5.1 CreateListing — POST /admin/listings ---

type createListingRequest struct {
	DomainName       string `json:"domainName"`
	SellerShopperID  string `json:"sellerShopperId"`
	AskingPriceUsd   *int64 `json:"askingPriceUsd"`
	ReservePriceUsd  *int64 `json:"reservePriceUsd"`
	EndTime          string `json:"endTime"`
	StartTime        string `json:"startTime"`
	AuctionTypeID    *int   `json:"auctionTypeId"`
	ListingType      string `json:"listingType"`
	AutoExtEnabled   *bool  `json:"autoExtEnabled"`
	AutoExtWindowSec *int   `json:"autoExtWindowSec"`
	AutoExtSeconds   *int   `json:"autoExtSeconds"`
}

func (h *AdminHandler) CreateListing(w http.ResponseWriter, r *http.Request) {
	var req createListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	if req.DomainName == "" || req.SellerShopperID == "" {
		WriteJSON(w, 400, map[string]string{"error": "domainName and sellerShopperId are required"})
		return
	}

	now := lifecycle.Now()

	// Apply defaults
	askingPrice := int64(5_000_000)
	if req.AskingPriceUsd != nil {
		askingPrice = *req.AskingPriceUsd
	}
	reservePrice := int64(0)
	if req.ReservePriceUsd != nil {
		reservePrice = *req.ReservePriceUsd
	}
	startTime := now.Format(time.RFC3339)
	if req.StartTime != "" {
		startTime = req.StartTime
	}
	endTime := now.Add(5 * time.Minute).Format(time.RFC3339)
	if req.EndTime != "" {
		endTime = req.EndTime
	}
	auctionTypeID := 16
	if req.AuctionTypeID != nil {
		auctionTypeID = *req.AuctionTypeID
	}
	listingType := "EXPIRY_AUCTIONS"
	if req.ListingType != "" {
		listingType = req.ListingType
	}
	autoExtEnabled := true
	if req.AutoExtEnabled != nil {
		autoExtEnabled = *req.AutoExtEnabled
	}
	autoExtWindowSec := 60
	if req.AutoExtWindowSec != nil {
		autoExtWindowSec = *req.AutoExtWindowSec
	}
	autoExtSeconds := 300
	if req.AutoExtSeconds != nil {
		autoExtSeconds = *req.AutoExtSeconds
	}

	listing := model.Listing{
		DomainName:       req.DomainName,
		ListingStatus:    model.StatusOpen,
		ListingType:      listingType,
		AuctionTypeID:    auctionTypeID,
		StartTime:        startTime,
		EndTime:          endTime,
		AskingPriceUsd:   askingPrice,
		ReservePriceUsd:  reservePrice,
		SellerShopperID:  req.SellerShopperID,
		AutoExtEnabled:   autoExtEnabled,
		AutoExtWindowSec: autoExtWindowSec,
		AutoExtSeconds:   autoExtSeconds,
	}

	id, err := h.Store.CreateListing(listing)
	if err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	log.Printf("ADMIN created listing=%d domain=%s endTime=%s", id, req.DomainName, endTime)

	WriteJSON(w, 201, map[string]interface{}{
		"listingId":     id,
		"domainName":    req.DomainName,
		"endTime":       endTime,
		"listingStatus": "OPEN",
	})
}

// --- 5.2 ListListings — GET /admin/listings ---

func (h *AdminHandler) ListListings(w http.ResponseWriter, r *http.Request) {
	listings, err := h.Store.ListListings()
	if err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	result := make([]map[string]interface{}, 0, len(listings))
	for _, listing := range listings {
		result = append(result, h.buildAdminListingJSON(&listing))
	}
	WriteJSON(w, 200, result)
}

// --- 5.2a GetListing — GET /admin/listings/{id} ---

func (h *AdminHandler) GetListing(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		WriteJSON(w, 400, map[string]string{"error": "invalid listing ID"})
		return
	}

	listing, err := h.Store.GetListing(id)
	if err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if listing == nil {
		WriteJSON(w, 404, map[string]string{"error": "listing not found"})
		return
	}

	WriteJSON(w, 200, h.buildAdminListingJSON(listing))
}

// --- 5.3 UpdateStatus — PUT /admin/listings/{id}/status ---

func (h *AdminHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		WriteJSON(w, 400, map[string]string{"error": "invalid listing ID"})
		return
	}

	var req struct {
		ListingStatus string `json:"listingStatus"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	listing, err := h.Store.GetListing(id)
	if err != nil || listing == nil {
		WriteJSON(w, 404, map[string]string{"error": "listing not found"})
		return
	}

	salePrice := listing.CurrentPriceUsd
	if err := h.Store.UpdateListingStatus(id, req.ListingStatus, salePrice); err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	log.Printf("ADMIN listing=%d status %s→%s", id, listing.ListingStatus, req.ListingStatus)

	updated, _ := h.Store.GetListing(id)
	WriteJSON(w, 200, h.buildAdminListingJSON(updated))
}

// --- 5.4 UpdateEndTime — PUT /admin/listings/{id}/endtime ---

func (h *AdminHandler) UpdateEndTime(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		WriteJSON(w, 400, map[string]string{"error": "invalid listing ID"})
		return
	}

	var req struct {
		EndTime    string `json:"endTime"`
		AddSeconds *int   `json:"addSeconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	listing, err := h.Store.GetListing(id)
	if err != nil || listing == nil {
		WriteJSON(w, 404, map[string]string{"error": "listing not found"})
		return
	}

	newEndTime := req.EndTime
	if req.AddSeconds != nil {
		currentEnd, err := time.Parse(time.RFC3339, listing.EndTime)
		if err != nil {
			WriteJSON(w, 500, map[string]string{"error": "failed to parse current endTime"})
			return
		}
		newEndTime = currentEnd.Add(time.Duration(*req.AddSeconds) * time.Second).Format(time.RFC3339)
	}

	if newEndTime == "" {
		WriteJSON(w, 400, map[string]string{"error": "endTime or addSeconds required"})
		return
	}

	if err := h.Store.UpdateListingEndTime(id, newEndTime); err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	log.Printf("ADMIN listing=%d endTime→%s", id, newEndTime)

	updated, _ := h.Store.GetListing(id)
	WriteJSON(w, 200, h.buildAdminListingJSON(updated))
}

// --- 5.5 SniperBid — POST /admin/listings/{id}/sniper-bid ---

func (h *AdminHandler) SniperBid(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		WriteJSON(w, 400, map[string]string{"error": "invalid listing ID"})
		return
	}

	var req struct {
		ShopperID    string `json:"shopperId"`
		BidAmountUsd int64  `json:"bidAmountUsd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	// Auto-create shopper if needed
	if _, err := h.Store.GetOrCreateShopper(req.ShopperID); err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	result, err := h.Engine.PlaceSniperBid(id, req.ShopperID, req.BidAmountUsd)
	if err != nil {
		if bidErr, ok := err.(*bidding.BidError); ok {
			WriteJSON(w, bidErr.HTTPStatus, map[string]string{"error": bidErr.Message})
		} else {
			WriteJSON(w, 500, map[string]string{"error": err.Error()})
		}
		return
	}

	log.Printf("ADMIN sniper-bid listing=%d shopper=%s amount=$%.2f",
		id, req.ShopperID, float64(req.BidAmountUsd)/1_000_000)

	WriteJSON(w, 200, result)
}

// --- 5.6 CreateShopper — POST /admin/shoppers ---

func (h *AdminHandler) CreateShopper(w http.ResponseWriter, r *http.Request) {
	var shopper model.Shopper
	if err := json.NewDecoder(r.Body).Decode(&shopper); err != nil {
		WriteJSON(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.Store.CreateShopper(shopper); err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	log.Printf("ADMIN created shopper=%s memberId=%d", shopper.ShopperID, shopper.MemberID)
	WriteJSON(w, 201, shopper)
}

// --- 5.6a ListShoppers — GET /admin/shoppers ---

func (h *AdminHandler) ListShoppers(w http.ResponseWriter, r *http.Request) {
	shoppers, err := h.Store.ListShoppers()
	if err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if shoppers == nil {
		shoppers = []model.Shopper{}
	}
	WriteJSON(w, 200, shoppers)
}

// --- 5.7 Reset — POST /admin/reset ---

func (h *AdminHandler) Reset(w http.ResponseWriter, r *http.Request) {
	h.Store.Reset()
	lifecycle.Reset()
	log.Printf("ADMIN reset — DB dropped and re-seeded, clock reset")
	WriteJSON(w, 200, map[string]string{"status": "reset"})
}

// --- 5.8 LoadScenario — POST /admin/scenarios/{name} ---

func (h *AdminHandler) LoadScenario(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if h.Scenario == nil {
		WriteJSON(w, 404, map[string]string{"error": "scenario loader not configured"})
		return
	}

	result, err := h.Scenario.Load(name)
	if err != nil {
		WriteJSON(w, 404, map[string]string{"error": fmt.Sprintf("scenario %q not found: %v", name, err)})
		return
	}

	log.Printf("ADMIN loaded scenario=%s", name)
	WriteJSON(w, 200, result)
}

// --- 5.9 UpdateConfig — PUT /admin/config ---

func (h *AdminHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AutoFinalize            *bool `json:"autoFinalize"`
		StatusTransitionDelayMs *int  `json:"statusTransitionDelayMs"`
		FinalizerIntervalMs     *int  `json:"finalizerIntervalMs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	h.Config.Update(req.AutoFinalize, req.StatusTransitionDelayMs, req.FinalizerIntervalMs)
	log.Printf("ADMIN config updated")
	WriteJSON(w, 200, h.Config.Snapshot())
}

// --- 5.9a GetConfig — GET /admin/config ---

func (h *AdminHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, 200, h.Config.Snapshot())
}

// --- 5.10 UpdateTime — PUT /admin/time ---

func (h *AdminHandler) UpdateTime(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OffsetSeconds *int   `json:"offsetSeconds"`
		FreezeAt      string `json:"freezeAt"`
		Reset         *bool  `json:"reset"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Reset != nil && *req.Reset {
		lifecycle.Reset()
		log.Printf("ADMIN time reset to realtime")
	} else if req.FreezeAt != "" {
		t, err := time.Parse(time.RFC3339, req.FreezeAt)
		if err != nil {
			WriteJSON(w, 400, map[string]string{"error": "invalid freezeAt format, use RFC3339"})
			return
		}
		lifecycle.Freeze(t)
		log.Printf("ADMIN time frozen at %s", req.FreezeAt)
	} else if req.OffsetSeconds != nil {
		lifecycle.SetOffset(*req.OffsetSeconds)
		log.Printf("ADMIN time offset by %d seconds", *req.OffsetSeconds)
	} else {
		WriteJSON(w, 400, map[string]string{"error": "provide offsetSeconds, freezeAt, or reset"})
		return
	}

	WriteJSON(w, 200, timeResponse())
}

// --- 5.10a GetTime — GET /admin/time ---

func (h *AdminHandler) GetTime(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, 200, timeResponse())
}

// --- 5.11 ExportDB — GET /admin/export ---

func (h *AdminHandler) ExportDB(w http.ResponseWriter, r *http.Request) {
	shoppers, err := h.Store.ListShoppers()
	if err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if shoppers == nil {
		shoppers = []model.Shopper{}
	}

	listings, err := h.Store.ListListings()
	if err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if listings == nil {
		listings = []model.Listing{}
	}

	bids, err := h.Store.ListAllBids()
	if err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if bids == nil {
		bids = []model.Bid{}
	}

	export := map[string]interface{}{
		"version":   1,
		"exportedAt": lifecycle.Now().Format(time.RFC3339),
		"shoppers":  shoppers,
		"listings":  listings,
		"bids":      bids,
	}

	log.Printf("ADMIN export — %d shoppers, %d listings, %d bids", len(shoppers), len(listings), len(bids))
	WriteJSON(w, 200, export)
}

// --- 5.12 ImportDB — POST /admin/import ---

func (h *AdminHandler) ImportDB(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Shoppers []model.Shopper `json:"shoppers"`
		Listings []model.Listing `json:"listings"`
		Bids     []model.Bid     `json:"bids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	// Wipe all tables and recreate without seeding
	h.Store.WipeAll()
	lifecycle.Reset()

	// Import shoppers
	for _, sh := range req.Shoppers {
		if err := h.Store.ImportShopper(sh); err != nil {
			WriteJSON(w, 500, map[string]string{"error": fmt.Sprintf("import shopper %s: %v", sh.ShopperID, err)})
			return
		}
	}

	// Import listings
	for _, l := range req.Listings {
		if err := h.Store.ImportListing(l); err != nil {
			WriteJSON(w, 500, map[string]string{"error": fmt.Sprintf("import listing %d: %v", l.ListingID, err)})
			return
		}
	}

	// Import bids
	for _, b := range req.Bids {
		if err := h.Store.ImportBid(b); err != nil {
			WriteJSON(w, 500, map[string]string{"error": fmt.Sprintf("import bid %s: %v", b.BidID, err)})
			return
		}
	}

	log.Printf("ADMIN import — %d shoppers, %d listings, %d bids", len(req.Shoppers), len(req.Listings), len(req.Bids))
	WriteJSON(w, 200, map[string]interface{}{
		"status":   "imported",
		"shoppers": len(req.Shoppers),
		"listings": len(req.Listings),
		"bids":     len(req.Bids),
	})
}

// --- 5.13 WipeDB — POST /admin/wipe ---

func (h *AdminHandler) WipeDB(w http.ResponseWriter, r *http.Request) {
	h.Store.WipeAll()
	lifecycle.Reset()
	log.Printf("ADMIN wipe — DB dropped and recreated (no seed data)")
	WriteJSON(w, 200, map[string]string{"status": "wiped"})
}

// --- 5.14 SetupSystem — POST /admin/setup ---

var setupTLDs = []string{".com", ".net", ".org", ".co", ".info", ".tv", ".us", ".cc", ".io", ".biz"}

var setupWords = map[byte][]string{
	'A': {"alpha", "apex", "atlas", "aqua", "arrow"},
	'B': {"bravo", "blaze", "byte", "bolt", "bloom"},
	'C': {"cyber", "core", "cloud", "crest", "craft"},
	'D': {"delta", "dash", "drift", "drive", "dawn"},
	'E': {"echo", "edge", "elite", "ember", "epic"},
	'F': {"flash", "flux", "forge", "frost", "fuse"},
	'G': {"grid", "glow", "gate", "grain", "glint"},
	'H': {"helix", "haze", "hive", "haven", "hyper"},
	'I': {"ion", "iris", "iron", "ivory", "ignite"},
	'J': {"jade", "jet", "jolt", "jump", "juno"},
	'K': {"kite", "knot", "keen", "karma", "krypton"},
	'L': {"lux", "link", "lance", "lyric", "lunar"},
	'M': {"mesa", "mint", "mist", "mocha", "matrix"},
	'N': {"nova", "nexus", "node", "neon", "nimbus"},
	'O': {"orbit", "onyx", "omega", "opal", "oxide"},
	'P': {"pulse", "peak", "prism", "pixel", "pine"},
	'Q': {"quest", "quartz", "qubit", "quad", "quill"},
	'R': {"ridge", "reef", "rune", "rush", "raven"},
	'S': {"spark", "slate", "solar", "swift", "surge"},
	'T': {"terra", "tidal", "trace", "turbo", "torch"},
	'U': {"ultra", "unity", "urban", "umbra", "uplink"},
	'V': {"vex", "volt", "vivid", "vapor", "vortex"},
	'W': {"wave", "wren", "wind", "warp", "wilde"},
	'X': {"xenon", "xray", "xero", "xylo", "xalt"},
	'Y': {"yonder", "yield", "yarrow", "yeti", "yukon"},
	'Z': {"zen", "zephyr", "zinc", "zone", "zero"},
}

var setupAskingPrices = []int64{
	5_000_000, 10_000_000, 15_000_000, 20_000_000, 25_000_000, 50_000_000, 75_000_000, 100_000_000,
}

func pickRandom[T any](items []T) T {
	return items[rand.Intn(len(items))]
}

func (h *AdminHandler) SetupSystem(w http.ResponseWriter, r *http.Request) {
	// Clean slate — wipe without seeding legacy defaults
	h.Store.WipeAll()
	lifecycle.Reset()

	// 1. Create 8 shoppers: 3 sellers + 5 buyers
	shoppers := []model.Shopper{
		{ShopperID: "shopper-seller-1", MemberID: 10001, CustomerID: "cust-seller-1", DisplayName: "Seller 1"},
		{ShopperID: "shopper-seller-2", MemberID: 10002, CustomerID: "cust-seller-2", DisplayName: "Seller 2"},
		{ShopperID: "shopper-seller-3", MemberID: 10003, CustomerID: "cust-seller-3", DisplayName: "Seller 3"},
		{ShopperID: "shopper-buyer-1", MemberID: 10004, CustomerID: "cust-buyer-1", DisplayName: "Buyer 1"},
		{ShopperID: "shopper-buyer-2", MemberID: 10005, CustomerID: "cust-buyer-2", DisplayName: "Buyer 2"},
		{ShopperID: "shopper-buyer-3", MemberID: 10006, CustomerID: "cust-buyer-3", DisplayName: "Buyer 3"},
		{ShopperID: "shopper-buyer-4", MemberID: 10007, CustomerID: "cust-buyer-4", DisplayName: "Buyer 4"},
		{ShopperID: "shopper-buyer-5", MemberID: 10008, CustomerID: "cust-buyer-5", DisplayName: "Buyer 5"},
	}
	for _, sh := range shoppers {
		if err := h.Store.CreateShopper(sh); err != nil {
			WriteJSON(w, 500, map[string]string{"error": fmt.Sprintf("create shopper %s: %v", sh.ShopperID, err)})
			return
		}
	}

	// 2. Create 26 A-Z listings with staggered end times (5 min for A ... 30 min for Z)
	now := lifecycle.Now()
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	sellers := []string{"shopper-seller-1", "shopper-seller-2", "shopper-seller-3"}
	createdListings := make([]map[string]interface{}, 0, 26)

	for i := 0; i < 26; i++ {
		letter := letters[i]
		words := setupWords[letter]
		word := pickRandom(words)
		suffix := rand.Intn(90) + 10
		tld := pickRandom(setupTLDs)
		domainName := fmt.Sprintf("%s%d%s", word, suffix, tld)

		durationMin := 5 + i // A=5min, B=6min, ..., Z=30min
		endTime := now.Add(time.Duration(durationMin) * time.Minute).Format(time.RFC3339)

		askingPrice := pickRandom(setupAskingPrices)
		hasReserve := rand.Float64() > 0.7
		reservePrice := int64(0)
		if hasReserve {
			reservePrice = askingPrice * int64(rand.Intn(4)+2)
		}
		autoExtEnabled := rand.Float64() > 0.3
		autoExtWindowSec := 60
		autoExtSeconds := []int{120, 300, 600}[rand.Intn(3)]

		listing := model.Listing{
			DomainName:       domainName,
			ListingStatus:    model.StatusOpen,
			ListingType:      "EXPIRY_AUCTIONS",
			AuctionTypeID:    16,
			StartTime:        now.Format(time.RFC3339),
			EndTime:          endTime,
			AskingPriceUsd:   askingPrice,
			ReservePriceUsd:  reservePrice,
			SellerShopperID:  sellers[i%len(sellers)],
			AutoExtEnabled:   autoExtEnabled,
			AutoExtWindowSec: autoExtWindowSec,
			AutoExtSeconds:   autoExtSeconds,
		}

		id, err := h.Store.CreateListing(listing)
		if err != nil {
			WriteJSON(w, 500, map[string]string{"error": fmt.Sprintf("create listing %s: %v", domainName, err)})
			return
		}

		createdListings = append(createdListings, map[string]interface{}{
			"listingId":  id,
			"domainName": domainName,
			"endTime":    endTime,
			"letter":     string(letter),
		})
	}

	log.Printf("ADMIN setup — 8 shoppers (3 sellers + 5 buyers) + 26 listings created")
	WriteJSON(w, 200, map[string]interface{}{
		"status":   "ready",
		"shoppers": len(shoppers),
		"listings": len(createdListings),
		"details":  createdListings,
	})
}

func timeResponse() map[string]interface{} {
	return map[string]interface{}{
		"serverTime": lifecycle.Now().Format(time.RFC3339),
		"mode":       lifecycle.Mode(),
	}
}
