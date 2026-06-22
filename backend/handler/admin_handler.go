package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
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
		"nextBidPriceUsd":      listing.NextBidPriceUsd,
		"biddersCount":         listing.BiddersCount,
		"bidsCount":            listing.BidsCount,
		"isAutoExtended":       listing.IsAutoExtended,
		"sellerShopperId":      listing.SellerShopperID,
		"highestBidderShopper": listing.HighestBidderShopper,
		"autoExtEnabled":       listing.AutoExtEnabled,
		"autoExtWindowSec":     listing.AutoExtWindowSec,
		"autoExtSeconds":       listing.AutoExtSeconds,
		"radarVisible":         listing.RadarVisible,
		"createdAt":            listing.CreatedAt,
		"bidHistory":           bidHistory,
	}
}

// --- 5.1 CreateListing — POST /admin/listings ---

type createListingRequest struct {
	DomainName       string `json:"domainName"`
	SellerShopperID  string `json:"sellerShopperId"`
	AskingPriceUsd   *int64 `json:"askingPriceUsd"`
	EndTime          string `json:"endTime"`
	StartTime        string `json:"startTime"`
	AuctionTypeID    *int   `json:"auctionTypeId"`
	ListingType      string `json:"listingType"`
	AutoExtEnabled   *bool  `json:"autoExtEnabled"`
	AutoExtWindowSec *int   `json:"autoExtWindowSec"`
	AutoExtSeconds   *int   `json:"autoExtSeconds"`
	RadarVisible     *bool  `json:"radarVisible"`
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
	autoExtWindowSec := h.Config.GetAutoExtWindowSec()
	if req.AutoExtWindowSec != nil {
		autoExtWindowSec = *req.AutoExtWindowSec
	}
	autoExtSeconds := h.Config.GetAutoExtSeconds()
	if req.AutoExtSeconds != nil {
		autoExtSeconds = *req.AutoExtSeconds
	}
	radarVisible := false
	if req.RadarVisible != nil {
		radarVisible = *req.RadarVisible
	}

	listing := model.Listing{
		DomainName:       req.DomainName,
		ListingStatus:    model.StatusOpen,
		ListingType:      listingType,
		AuctionTypeID:    auctionTypeID,
		StartTime:        startTime,
		EndTime:          endTime,
		AskingPriceUsd:   askingPrice,
		SellerShopperID:  req.SellerShopperID,
		AutoExtEnabled:   autoExtEnabled,
		AutoExtWindowSec: autoExtWindowSec,
		AutoExtSeconds:   autoExtSeconds,
		RadarVisible:     radarVisible,
	}

	id, err := h.Store.CreateListing(listing)
	if err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	// Seed an initial bid on radar-visible listings so they appear in the app's radar tab
	if radarVisible {
		h.Store.GetOrCreateShopper("shopper-buyer-1")
		h.Engine.PlaceSniperBid(id, "shopper-buyer-1", askingPrice)
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

// --- 5.4b UpdateRadarVisible — PUT /admin/listings/{id}/radar ---

func (h *AdminHandler) UpdateRadarVisible(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		WriteJSON(w, 400, map[string]string{"error": "invalid listing ID"})
		return
	}

	var req struct {
		RadarVisible bool `json:"radarVisible"`
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

	if err := h.Store.UpdateListingRadarVisible(id, req.RadarVisible); err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	log.Printf("ADMIN listing=%d radarVisible→%v", id, req.RadarVisible)

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

// --- 5.6b GetShopper — GET /admin/shoppers/{id} ---

func (h *AdminHandler) GetShopper(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shopperID := vars["id"]

	shopper, err := h.Store.GetShopper(shopperID)
	if err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if shopper == nil {
		WriteJSON(w, 404, map[string]string{"error": "shopper not found"})
		return
	}

	bids, err := h.Store.GetBidsForShopper(shopperID)
	if err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	bidHistory := make([]map[string]interface{}, 0, len(bids))
	for _, bid := range bids {
		listing, _ := h.Store.GetListing(bid.ListingID)
		entry := map[string]interface{}{
			"bidId":        bid.BidID,
			"listingId":    bid.ListingID,
			"shopperId":    bid.ShopperID,
			"bidAmountUsd": bid.BidAmountUsd,
			"bidType":      bid.BidType,
			"bidStatus":    bid.BidStatus,
			"isHighBid":    bid.IsHighBid,
			"parentBidId":  bid.ParentBidID,
			"createdAt":    bid.CreatedAt,
		}
		if listing != nil {
			entry["domainName"] = listing.DomainName
			entry["listingStatus"] = listing.ListingStatus
			entry["highestBidderShopper"] = listing.HighestBidderShopper
		}
		bidHistory = append(bidHistory, entry)
	}

	WriteJSON(w, 200, map[string]interface{}{
		"shopperId":   shopper.ShopperID,
		"memberId":    shopper.MemberID,
		"customerId":  shopper.CustomerID,
		"displayName": shopper.DisplayName,
		"bidHistory":  bidHistory,
	})
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
		AutoExtWindowSec        *int  `json:"autoExtWindowSec"`
		AutoExtSeconds          *int  `json:"autoExtSeconds"`
		IncludeBin              *bool `json:"includeBin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, 400, map[string]string{"error": "invalid request body"})
		return
	}

	h.Config.Update(req.AutoFinalize, req.StatusTransitionDelayMs, req.FinalizerIntervalMs, req.AutoExtWindowSec, req.AutoExtSeconds, req.IncludeBin)
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
	// Optional body:
	//   { "durationMinutes": <int>,   // duration for ALL generated auctions (default 5, must be >= 1)
	//     "appShopperId": "<id>" }    // when set, a few FINISHED auctions are won/lost by this
	//                                  // shopper so the app's Won/Lost tabs show real data.
	durationMin := 5
	appShopperID := ""
	if r.Body != nil {
		var body struct {
			DurationMinutes *int   `json:"durationMinutes"`
			AppShopperID    string `json:"appShopperId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			if body.DurationMinutes != nil {
				if *body.DurationMinutes < 1 {
					WriteJSON(w, 400, map[string]string{"error": "durationMinutes must be >= 1"})
					return
				}
				durationMin = *body.DurationMinutes
			}
			appShopperID = strings.TrimSpace(body.AppShopperID)
		}
	}

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

		// All generated auctions share the configured duration (default 5 min).
		endTime := now.Add(time.Duration(durationMin) * time.Minute).Format(time.RFC3339)

		askingPrice := pickRandom(setupAskingPrices)
		autoExtEnabled := rand.Float64() > 0.3
		autoExtWindowSec := 60
		autoExtSeconds := []int{120, 300, 600}[rand.Intn(3)]
		radarVisible := rand.Float64() > 0.6

		listing := model.Listing{
			DomainName:       domainName,
			ListingStatus:    model.StatusOpen,
			ListingType:      "EXPIRY_AUCTIONS",
			AuctionTypeID:    16,
			StartTime:        now.Format(time.RFC3339),
			EndTime:          endTime,
			AskingPriceUsd:   askingPrice,
			SellerShopperID:  pickRandom(sellers),
			AutoExtEnabled:   autoExtEnabled,
			AutoExtWindowSec: autoExtWindowSec,
			AutoExtSeconds:   autoExtSeconds,
			RadarVisible:     radarVisible,
		}

		id, err := h.Store.CreateListing(listing)
		if err != nil {
			WriteJSON(w, 500, map[string]string{"error": fmt.Sprintf("create listing %s: %v", domainName, err)})
			return
		}

		// Seed an initial bid on radar-visible listings so they appear in the app's radar tab
		if radarVisible {
			buyers := []string{"shopper-buyer-1", "shopper-buyer-2", "shopper-buyer-3", "shopper-buyer-4", "shopper-buyer-5"}
			buyer := buyers[rand.Intn(len(buyers))]
			h.Engine.PlaceSniperBid(id, buyer, askingPrice)
		}

		createdListings = append(createdListings, map[string]interface{}{
			"listingId":  id,
			"domainName": domainName,
			"endTime":    endTime,
			"letter":     string(letter),
		})
	}

	// Optional: assign a few REAL finished auctions to the app user's shopper so the Won/Lost
	// tabs show genuine data (won = SOLD + this shopper is highest bidder; lost = SOLD + this
	// shopper bid but someone else won).
	wonCount, lostCount := 0, 0
	if appShopperID != "" {
		h.Store.GetOrCreateShopper(appShopperID)
		startPast := now.Add(-2 * time.Hour).Format(time.RFC3339)
		endPast := now.Add(-1 * time.Hour).Format(time.RFC3339)
		mkFinished := func(domain string, winningBid int64) (int64, bool) {
			id, err := h.Store.CreateListing(model.Listing{
				DomainName: domain, ListingStatus: model.StatusOpen, ListingType: "EXPIRY_AUCTIONS",
				AuctionTypeID: 16, StartTime: startPast, EndTime: endPast,
				AskingPriceUsd: 5_000_000, SellerShopperID: "shopper-seller-1",
			})
			if err != nil {
				return 0, false
			}
			return id, true
		}
		// WON: app shopper is the highest (only) bidder, then the auction is SOLD.
		for n := 1; n <= 3; n++ {
			if id, ok := mkFinished(fmt.Sprintf("won-domain-%d.com", n), 5_000_000); ok {
				if _, err := h.Engine.PlaceSniperBid(id, appShopperID, 5_000_000); err == nil {
					h.Store.UpdateListingStatus(id, "SOLD", 5_000_000)
					wonCount++
				}
			}
		}
		// LOST: app shopper bids, another buyer outbids, then the auction is SOLD.
		for n := 1; n <= 2; n++ {
			if id, ok := mkFinished(fmt.Sprintf("lost-domain-%d.com", n), 10_000_000); ok {
				h.Engine.PlaceSniperBid(id, appShopperID, 5_000_000)
				if _, err := h.Engine.PlaceSniperBid(id, "shopper-buyer-2", 10_000_000); err == nil {
					h.Store.UpdateListingStatus(id, "SOLD", 10_000_000)
					lostCount++
				}
			}
		}
		log.Printf("ADMIN setup — assigned %d won + %d lost to appShopper=%s", wonCount, lostCount, appShopperID)
	}

	log.Printf("ADMIN setup — 8 shoppers (3 sellers + 5 buyers) + 26 listings created")
	WriteJSON(w, 200, map[string]interface{}{
		"status":       "ready",
		"shoppers":     len(shoppers),
		"listings":     len(createdListings),
		"appShopperId": appShopperID,
		"won":          wonCount,
		"lost":         lostCount,
		"details":      createdListings,
	})
}

// --- 5.15 GenerateBinListings — POST /admin/listings/bin ---

// binTypeToListingType maps an auction_type code to the app's listingType string.
func binTypeToListingType(auctionTypeID int) string {
	switch auctionTypeID {
	case 20, 39:
		return "CLOSEOUT_DOMAINS"
	case 11:
		return "BUY_IT_NOW"
	case 9, 10:
		return "MEMBER_LISTINGS"
	default:
		return "BUY_IT_NOW"
	}
}

// GenerateBinListings appends BIN/closeout/OCO listings on top of whatever already
// exists (no wipe), mirroring SetupSystem's per-type domain generation. Each created
// listing carries the inventory type that drives a distinct itc code in the app.
func (h *AdminHandler) GenerateBinListings(w http.ResponseWriter, r *http.Request) {
	countPerType := 1
	durationMin := 60
	types := []int{20, 39, 11, 10, 9}

	if r.Body != nil {
		var body struct {
			CountPerType    *int  `json:"countPerType"`
			DurationMinutes *int  `json:"durationMinutes"`
			Types           []int `json:"types"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			if body.CountPerType != nil && *body.CountPerType >= 1 {
				countPerType = *body.CountPerType
			}
			if body.DurationMinutes != nil && *body.DurationMinutes >= 1 {
				durationMin = *body.DurationMinutes
			}
			if len(body.Types) > 0 {
				types = body.Types
			}
		}
	}

	// Ensure a seller exists (seed defaults may only have "shopper-seller").
	seller := "shopper-seller-1"
	h.Store.GetOrCreateShopper(seller)

	now := lifecycle.Now()
	startTime := now.Format(time.RFC3339)
	endTime := now.Add(time.Duration(durationMin) * time.Minute).Format(time.RFC3339)
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	created := make([]map[string]interface{}, 0, len(types)*countPerType)

	for _, auctionTypeID := range types {
		listingType := binTypeToListingType(auctionTypeID)
		for n := 0; n < countPerType; n++ {
			letter := letters[rand.Intn(len(letters))]
			word := pickRandom(setupWords[letter])
			suffix := rand.Intn(900) + 100
			tld := pickRandom(setupTLDs)
			domainName := fmt.Sprintf("%s%d%s", word, suffix, tld)
			askingPrice := pickRandom(setupAskingPrices)

			listing := model.Listing{
				DomainName:      domainName,
				ListingStatus:   model.StatusOpen,
				ListingType:     listingType,
				AuctionTypeID:   auctionTypeID,
				StartTime:       startTime,
				EndTime:         endTime,
				AskingPriceUsd:  askingPrice,
				SellerShopperID: seller,
				AutoExtEnabled:  false,
				RadarVisible:    true,
			}

			id, err := h.Store.CreateListing(listing)
			if err != nil {
				WriteJSON(w, 500, map[string]string{"error": fmt.Sprintf("create BIN listing %s: %v", domainName, err)})
				return
			}

			created = append(created, map[string]interface{}{
				"listingId":     id,
				"domainName":    domainName,
				"auctionTypeId": auctionTypeID,
				"listingType":   listingType,
				"endTime":       endTime,
			})
		}
	}

	log.Printf("ADMIN generate-bin — %d BIN/closeout/OCO listings appended", len(created))
	WriteJSON(w, 201, map[string]interface{}{
		"status":   "created",
		"listings": len(created),
		"types":    types,
		"details":  created,
	})
}

// --- 5.16 ListCartEvents — GET /admin/cart-events ---

func (h *AdminHandler) ListCartEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.Store.ListCartEvents()
	if err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if events == nil {
		events = []model.CartEvent{}
	}
	WriteJSON(w, 200, events)
}

// --- 5.17 ClearCartEvents — DELETE /admin/cart-events ---

func (h *AdminHandler) ClearCartEvents(w http.ResponseWriter, r *http.Request) {
	if err := h.Store.ClearCartEvents(); err != nil {
		WriteJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	log.Printf("ADMIN cart-events cleared")
	WriteJSON(w, 200, map[string]string{"status": "cleared"})
}

func timeResponse() map[string]interface{} {
	return map[string]interface{}{
		"serverTime": lifecycle.Now().Format(time.RFC3339),
		"mode":       lifecycle.Mode(),
	}
}
