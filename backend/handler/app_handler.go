package handler

import (
	"encoding/json"
	"fmt"
	"log"
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

type AppHandler struct {
	Store         *store.Store
	Config        *config.Config
	Engine        *bidding.Engine
	AuctionUpstream   string
	FindUpstream  string
}

func NewAppHandler(s *store.Store, cfg *config.Config, eng *bidding.Engine, auctionUpstream, findUpstream string) *AppHandler {
	return &AppHandler{Store: s, Config: cfg, Engine: eng, AuctionUpstream: auctionUpstream, FindUpstream: findUpstream}
}

// GetListing handles GET /v1/aftermarket/domains/listings/{listingId}
func (h *AppHandler) GetListing(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	listingIDStr := vars["listingId"]
	listingID, err := strconv.ParseInt(listingIDStr, 10, 64)
	if err != nil {
		WriteError(w, "INVALID_LISTING_ID", "Invalid listing ID", 400)
		return
	}

	shopper, err := ResolveShopper(r, h.Store)
	if err != nil {
		WriteError(w, "MISSING_SHOPPER", err.Error(), 400)
		return
	}

	listing, err := h.Store.GetListing(listingID)
	if err != nil {
		WriteError(w, "SERVER_ERROR", "Internal server error", 500)
		return
	}
	if listing == nil {
		if proxyToUpstream(h.AuctionUpstream, w, r) {
			return
		}
		WriteError(w, "LISTING_NOT_FOUND", "Listing not found", 404)
		return
	}

	log.Printf("GET listing=%d shopper=%s status=%s", listing.ListingID, shopper.ShopperID, listing.ListingStatus)

	result := h.buildListingJSON(listing, shopper)
	WriteJSON(w, 200, result)
}

// GetBiddingListings handles GET /v1/aftermarket/domains/bidding
func (h *AppHandler) GetBiddingListings(w http.ResponseWriter, r *http.Request) {
	shopper, err := ResolveShopper(r, h.Store)
	if err != nil {
		// Local-test fallback: the real SSO JWT can't always be mapped to an emulator shopper.
		// Use a default buyer so the Bidding tab still populates instead of 400 MISSING_SHOPPER.
		shopper, _ = h.Store.GetOrCreateShopper("shopper-buyer-1")
	}

	allListings, _ := h.Store.ListListings()
	var result []map[string]interface{}
	if shopper != nil {
		for _, listing := range allListings {
			hasBid, _ := h.Store.HasShopperBidOnListing(shopper.ShopperID, listing.ListingID)
			if hasBid {
				result = append(result, h.buildListingJSON(&listing, shopper))
			}
		}
	}

	// Local-test fallback: when the resolved shopper has no bids, surface OPEN auctions so the
	// app's Bidding tab is non-empty and bid flows can be driven end-to-end. Returns local only
	// (no upstream proxy/merge) to stay fast and fully controllable.
	if len(result) == 0 {
		includeBin := h.Config.GetIncludeBin()
		for i := range allListings {
			if allListings[i].ListingStatus != model.StatusOpen {
				continue
			}
			if !includeBin && isBinInventoryType(&allListings[i]) {
				continue
			}
			result = append(result, h.buildListingJSON(&allListings[i], shopper))
		}
	}

	if result == nil {
		result = []map[string]interface{}{}
	}
	WriteJSON(w, 200, map[string]interface{}{
		"lastUpdatedTime": lifecycle.Now().Format(time.RFC3339),
		"viewType":        "SNAPSHOT",
		"listings":        result,
	})
}

// GetWatchlistListings handles GET /v1/aftermarket/domains/watches
// Local-test helper: surfaces OPEN listings so the app's Watchlist tab is non-empty and
// watchlist bid / buy-it-now flows can be driven. Returns local only (no upstream).
func (h *AppHandler) GetWatchlistListings(w http.ResponseWriter, r *http.Request) {
	shopper, err := ResolveShopper(r, h.Store)
	if err != nil {
		shopper, _ = h.Store.GetOrCreateShopper("shopper-buyer-1")
	}

	includeBin := h.Config.GetIncludeBin()
	allListings, _ := h.Store.ListListings()
	result := make([]map[string]interface{}, 0, len(allListings))
	for i := range allListings {
		if allListings[i].ListingStatus != model.StatusOpen {
			continue
		}
		if !includeBin && isBinInventoryType(&allListings[i]) {
			continue
		}
		result = append(result, h.buildListingJSON(&allListings[i], shopper))
	}

	WriteJSON(w, 200, map[string]interface{}{
		"lastUpdatedTime": lifecycle.Now().Format(time.RFC3339),
		"viewType":        "SNAPSHOT",
		"listings":        result,
	})
}

type placeBidRequest struct {
	UsdBidAmount  int64 `json:"usdBidAmount"`
	IsTosAccepted bool  `json:"isTosAccepted"`
}

// PlaceBid handles POST /v1/aftermarket/domains/listings/{listingId}/bids
func (h *AppHandler) PlaceBid(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	listingID, err := strconv.ParseInt(vars["listingId"], 10, 64)
	if err != nil {
		WriteError(w, "LISTING_NOT_FOUND", "Invalid listing ID", 404)
		return
	}

	shopper, err := ResolveShopper(r, h.Store)
	if err != nil {
		WriteError(w, "MISSING_SHOPPER", err.Error(), 400)
		return
	}

	var req placeBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, "SERVER_ERROR", "Invalid request body", 400)
		return
	}

	result, err := h.Engine.PlaceBid(bidding.BidRequest{
		ListingID:     listingID,
		ShopperID:     shopper.ShopperID,
		UsdBidAmount:  req.UsdBidAmount,
		IsTosAccepted: req.IsTosAccepted,
	})

	if err != nil {
		if bidErr, ok := err.(*bidding.BidError); ok {
			WriteError(w, bidErr.Code, bidErr.Message, bidErr.HTTPStatus)
		} else {
			WriteError(w, "SERVER_ERROR", "Internal server error", 500)
		}
		return
	}

	log.Printf("POST /v1/aftermarket/domains/listings/%d/bids shopperId=%s bidAmount=$%.2f → 200 SUCCESS",
		listingID, shopper.ShopperID, float64(req.UsdBidAmount)/1_000_000)

	WriteJSON(w, 200, result)
}

// GetWonListings handles GET /v1/aftermarket/domains/won
func (h *AppHandler) GetWonListings(w http.ResponseWriter, r *http.Request) {
	shopper, err := ResolveShopper(r, h.Store)
	if err != nil {
		WriteError(w, "MISSING_SHOPPER", err.Error(), 400)
		return
	}

	// Won listings are REAL: SOLD auctions the resolved shopper actually won. Assign them at
	// setup via POST /admin/setup { "appShopperId": "<id>" }. No OPEN-listing fallback here.
	listings, _ := h.Store.GetWonListingsForShopper(shopper.ShopperID)
	result := make([]map[string]interface{}, 0, len(listings))
	for _, l := range listings {
		result = append(result, h.buildWonListingJSON(&l))
	}

	WriteJSON(w, 200, map[string]interface{}{
		"lastUpdatedTime": lifecycle.Now().Format(time.RFC3339),
		"viewType":        "SNAPSHOT",
		"listings":        result,
	})
}

// GetLostListings handles GET /v1/aftermarket/domains/didNotWin
func (h *AppHandler) GetLostListings(w http.ResponseWriter, r *http.Request) {
	shopper, err := ResolveShopper(r, h.Store)
	if err != nil {
		WriteError(w, "MISSING_SHOPPER", err.Error(), 400)
		return
	}

	listings, _ := h.Store.GetLostListingsForShopper(shopper.ShopperID)
	var result []map[string]interface{}
	for _, l := range listings {
		result = append(result, h.buildLostListingJSON(&l))
	}

	// No local results → forward entirely to upstream
	if len(result) == 0 && h.AuctionUpstream != "" {
		if proxyToUpstream(h.AuctionUpstream, w, r) {
			return
		}
	}

	// Local results exist → merge with upstream (local wins on dupes)
	if len(result) > 0 && h.AuctionUpstream != "" {
		upstreamItems, err := fetchUpstreamListings(h.AuctionUpstream, "/v1/aftermarket/domains/didNotWin", r)
		if err == nil && upstreamItems != nil {
			result = mergeListings(result, upstreamItems)
		}
	}

	if result == nil {
		result = []map[string]interface{}{}
	}
	WriteJSON(w, 200, map[string]interface{}{
		"lastUpdatedTime": lifecycle.Now().Format(time.RFC3339),
		"viewType":        "SNAPSHOT",
		"listings":        result,
	})
}

// SearchListings handles GET /v4/aftermarket/find/auction/recommend
// With query param: searches by domain name. Without: returns radar-visible listings.
func (h *AppHandler) SearchListings(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")

	// Radar (the app's recommend feed) sends no query; search sends a query.
	isRadar := query == ""

	includeBin := h.Config.GetIncludeBin()
	var result []map[string]interface{}
	if query != "" {
		listings, _ := h.Store.SearchListingsByDomain(query)
		for i := range listings {
			if !includeBin && isBinInventoryType(&listings[i]) {
				continue
			}
			result = append(result, h.buildSearchResultJSON(&listings[i]))
		}
	} else {
		listings, _ := h.Store.GetRadarListings()
		for i := range listings {
			if !includeBin && isBinInventoryType(&listings[i]) {
				continue
			}
			result = append(result, h.buildSearchResultJSON(&listings[i]))
		}
	}

	// Radar feed is fully controlled by the posted radar domains: serve ONLY those.
	// Merging/proxying the full upstream recommend feed (~500 items, multi-second) makes the
	// app's radar request time out (DomainRadarFindRequest onFailure -1), so radar never shows
	// the posted domains. Keep it small and fast; post radar domains via PUT /admin/listings/{id}/radar.
	if isRadar {
		if result == nil {
			result = []map[string]interface{}{}
		}
		WriteJSON(w, 200, map[string]interface{}{"results": result})
		return
	}

	// --- Search path (query present): fall back to / merge with the real Find upstream ---
	// If no local results, forward entirely to find-upstream.
	if len(result) == 0 && h.FindUpstream != "" {
		if proxyToUpstream(h.FindUpstream, w, r) {
			return
		}
	}

	// Local matches exist → return ONLY those (no upstream merge). Merging the full ~500-item
	// upstream feed is large/slow and can make the app's search request time out, hiding the
	// emulator's controllable domains. Keep search fast and deterministic.
	if result == nil {
		result = []map[string]interface{}{}
	}
	WriteJSON(w, 200, map[string]interface{}{
		"results": result,
	})
}

// GetMemberAuthorized handles GET /v1/aftermarket/domains/member/authorized
// Always returns ACTIVE membership so the app allows bidding.
func (h *AppHandler) GetMemberAuthorized(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, 200, map[string]interface{}{
		"id":               0,
		"membershipStatus": "ACTIVE",
	})
}

type addToCartItem struct {
	DomainName   string `json:"domainName"`
	ListingID    int64  `json:"listingId"`
	RequestPrice int64  `json:"requestPrice"`
	AcceptTos    bool   `json:"acceptTos"`
	ItcCode      string `json:"itcCode"`
}

// parseItcCode splits "dna_invapp_<area>_android_<inventory>" into (area, inventory).
// Handles the base form with no surface, both collapsed ("dna_invapp_android_…") and
// empty ("dna_invapp__android_…"). Returns empty strings for segments it cannot find.
func parseItcCode(itc string) (area, inventory string) {
	const mid = "android_"
	rest := strings.TrimPrefix(itc, "dna_invapp_")
	idx := strings.Index(rest, mid)
	if idx < 0 {
		return "", ""
	}
	inventory = rest[idx+len(mid):]
	area = strings.TrimSuffix(rest[:idx], "_")
	return area, inventory
}

// AddToCart handles POST /v1/aftermarket/domains/cart. The app sends a JSON array
// of cart items plus an X-Itc-Code header. We capture the itc code per item so the
// dashboard can verify which itc string the app sent for each inventory type.
func (h *AppHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	headerItc := r.Header.Get("X-Itc-Code")

	var items []addToCartItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		WriteError(w, "SERVER_ERROR", "Invalid request body", 400)
		return
	}

	valid := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		itc := item.ItcCode
		if itc == "" {
			itc = headerItc
		}
		area, inventory := parseItcCode(itc)

		inventoryType := 0
		if listing, _ := h.Store.GetListing(item.ListingID); listing != nil {
			inventoryType = listing.AuctionTypeID
		}

		if _, err := h.Store.CreateCartEvent(model.CartEvent{
			DomainName:    item.DomainName,
			ListingID:     item.ListingID,
			InventoryType: inventoryType,
			ItcCode:       itc,
			ItcInventory:  inventory,
			Area:          area,
			RequestPrice:  item.RequestPrice,
		}); err != nil {
			WriteError(w, "SERVER_ERROR", "Failed to record cart event", 500)
			return
		}

		log.Printf("CART add domain=%s listing=%d itc=%s (inventory=%s)", item.DomainName, item.ListingID, itc, inventory)
		valid = append(valid, map[string]interface{}{"domainName": item.DomainName})
	}

	WriteJSON(w, 200, map[string]interface{}{
		"validItems":  valid,
		"failedItems": []interface{}{},
	})
}

func (h *AppHandler) buildWonListingJSON(l *model.Listing) map[string]interface{} {
	salePrice := []map[string]interface{}{}
	if l.SalePriceUsd != nil && *l.SalePriceUsd > 0 {
		salePrice = priceArray(*l.SalePriceUsd)
	}

	return map[string]interface{}{
		"listingId":         l.ListingID,
		"domainName":        l.DomainName,
		"listingStatus":     l.ListingStatus,
		"listingType":       l.ListingType,
		"auctionTypeId":     l.AuctionTypeID,
		"endTime":           toFractionalISO(l.EndTime),
		"askingPrice":       priceArray(l.AskingPriceUsd),
		"salePrice":         salePrice,
		"bidsOrOffersCount": l.BidsCount,
	}
}

func (h *AppHandler) buildLostListingJSON(l *model.Listing) map[string]interface{} {
	salePrice := []map[string]interface{}{}
	if l.SalePriceUsd != nil && *l.SalePriceUsd > 0 {
		salePrice = priceArray(*l.SalePriceUsd)
	}

	return map[string]interface{}{
		"listingId":         l.ListingID,
		"domainName":        l.DomainName,
		"listingType":       l.ListingType,
		"endTime":           toFractionalISO(l.EndTime),
		"askingPrice":       priceArray(l.AskingPriceUsd),
		"salePrice":         salePrice,
		"bidsOrOffersCount": l.BidsCount,
	}
}

func (h *AppHandler) buildSearchResultJSON(l *model.Listing) map[string]interface{} {
	askingPrice := float64(l.AskingPriceUsd) / 1_000_000
	currentBid := float64(l.CurrentPriceUsd) / 1_000_000

	// BIN/closeout/OCO-BIN listings expose a Buy Now price; the app keys the
	// buy-now action off these fields (auction_type already selects the cell type).
	hasBuyNow := buyNowInventoryTypes[l.AuctionTypeID]
	buyNowAmount := 0.0
	buyNowDisplay := ""
	if hasBuyNow {
		buyNowAmount = askingPrice
		buyNowDisplay = formatUSD(askingPrice)
	}

	return map[string]interface{}{
		// IDs
		"auction_id": l.ListingID,
		"fqdn":       l.DomainName,
		"fqdn_from_feed": l.DomainName,

		// Type & status (integers expected by Android Gson models)
		"auction_type":   l.AuctionTypeID,
		"auction_status": listingStatusToInt(l.ListingStatus),
		"active":         l.ListingStatus == model.StatusOpen,

		// Prices (doubles)
		"auction_price":     askingPrice,
		"auction_price_usd": askingPrice,
		"current_bid_price":     currentBid,
		"current_bid_price_usd": currentBid,
		"start_bid_amount":     askingPrice,
		"start_bid_amount_usd": askingPrice,
		"buy_it_now_amount":     buyNowAmount,
		"buy_it_now_amount_usd": buyNowAmount,
		"reserved_price_amount":     0,
		"reserved_price_amount_usd": 0,
		"valuation_price":     0,
		"valuation_price_usd": 0,

		// Display strings (required by Android — Gson sets missing Strings to null → NPE)
		"auction_price_display":         formatUSD(askingPrice),
		"auction_price_display_usd":     formatUSD(askingPrice),
		"current_bid_price_display":     formatUSD(currentBid),
		"current_bid_price_display_usd": formatUSD(currentBid),
		"start_bid_amount_display":      formatUSD(askingPrice),
		"start_bid_amount_display_usd":  formatUSD(askingPrice),
		"buy_it_now_amount_display":     buyNowDisplay,
		"buy_it_now_amount_display_usd": buyNowDisplay,
		"reserved_price_amount_display":     "",
		"reserved_price_amount_display_usd": "",
		"valuation_price_display":     "",
		"valuation_price_display_usd": "",

		// Times
		"end_time":           l.EndTime,
		"auction_end_time":   formatSpaceTime(l.EndTime),
		"auction_start_time": l.StartTime,

		// Counts & flags
		"bids":              l.BidsCount,
		"monthly_traffic":   0,
		"reserved_price_flag":  false,
		"buy_it_now_flag":      hasBuyNow,
		"bid_accepted_flag":    false,
		"is_website_included":  false,
		"feature_listing_flag": false,
		"include_in_search_flag":                true,
		"display_result_in_category_list_flag":  true,
		"sub_category_feature_listing_flag":     false,
		"add_i_category_listing_flag":           false,
		"on_sale_percent": 0,
		"appraised_value": 0,

		// Misc
		"data_source":      "mock",
		"item_description": "",
		"vendor_id":        0,
		"isidn":            false,
		"auction_adult":    false,
	}
}

// toFractionalISO ensures an ISO timestamp has microsecond fractional seconds
// (e.g. "2026-03-04T03:01:37Z" → "2026-03-04T03:01:37.000000Z").
// BidAdapter on Android parses endTime with pattern "yyyy-MM-dd'T'HH:mm:ss.SSSSSS'Z'".
func toFractionalISO(iso string) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso
	}
	return t.UTC().Format("2006-01-02T15:04:05.000000Z")
}

// formatUSD formats a dollar amount as "$X.XX" or "" if zero.
func formatUSD(amount float64) string {
	if amount <= 0 {
		return ""
	}
	return fmt.Sprintf("$%.2f", amount)
}

// formatSpaceTime converts ISO "2006-01-02T15:04:05Z" to "2006-01-02 15:04:05" (Find API format).
func formatSpaceTime(iso string) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso
	}
	return t.Format("2006-01-02 15:04:05")
}

// listingStatusToInt maps internal status strings to the numeric codes the Find API uses.
func listingStatusToInt(status string) int {
	switch status {
	case model.StatusOpen:
		return 4
	case model.StatusClosed:
		return 5
	case model.StatusSold:
		return 6
	default:
		return 4
	}
}

// priceTypeForListing maps a listing to the app's PriceType. BIN listings (ListingType
// "BUY_IT_NOW"/"BIN") render as buy-it-now; everything else is a biddable AUCTION. Without
// this the app defaulted empty priceType to a "$0 Buy Now" row that exposes no bid action.
func priceTypeForListing(l *model.Listing) string {
	if strings.EqualFold(l.ListingType, "BUY_IT_NOW") || strings.EqualFold(l.ListingType, "BIN") {
		return "BUY_IT_NOW"
	}
	return "AUCTION"
}

// binInventoryTypes are the auction_type codes the app renders as BIN / closeout /
// OCO rows (i.e. non-biddable add-to-cart inventory). These map to the itc codes
// expirycloseout / m2mbin / m2moco / m2mocobin in AuctionRequestAddToCart.
var binInventoryTypes = map[int]bool{
	20: true, // GoDaddy closeout    → expirycloseout
	39: true, // partner closeout    → expirycloseout
	11: true, // public buy-it-now   → m2mbin
	10: true, // public OCO + buy now → m2mocobin
	9:  true, // public OCO (offer)  → m2moco
}

// buyNowInventoryTypes are the BIN types that expose an actual Buy Now action
// (everything in binInventoryTypes except the offer-only OCO types).
var buyNowInventoryTypes = map[int]bool{
	20: true, 39: true, 11: true, 10: true,
}

// isBinInventoryType reports whether a listing is BIN/closeout/OCO inventory
// (used to gate app-facing visibility behind the includeBin config toggle).
func isBinInventoryType(l *model.Listing) bool {
	return binInventoryTypes[l.AuctionTypeID]
}

func (h *AppHandler) buildListingJSON(listing *model.Listing, shopper *model.Shopper) map[string]interface{} {
	// Fetch bids for bid history
	bids, _ := h.Store.GetActiveBidsForListing(listing.ListingID)

	// Build bid history array
	bidHistory := make([]map[string]interface{}, 0, len(bids))
	for _, bid := range bids {
		bidShopper, _ := h.Store.GetShopper(bid.ShopperID)
		memberID := int64(0)
		if bidShopper != nil {
			memberID = bidShopper.MemberID
		}
		bidHistory = append(bidHistory, map[string]interface{}{
			"bidAmount":         priceArray(bid.BidAmountUsd),
			"bidDate":           toFractionalISO(bid.CreatedAt),
			"bidExpirationDate": toFractionalISO(listing.EndTime),
			"bidder":            memberID,
			"comment":           "",
		})
	}

	// Compute memberBiddingStatus
	memberBiddingStatus := "NOT_BIDDING"
	if shopper != nil {
		hasBid, _ := h.Store.HasShopperBidOnListing(shopper.ShopperID, listing.ListingID)
		if hasBid {
			if listing.HighestBidderShopper == shopper.ShopperID {
				memberBiddingStatus = "HIGHEST_BIDDER"
			} else {
				memberBiddingStatus = "OUTBID"
			}
		}
	}

	// Compute proxyBidPrice
	proxyBidPrice := []map[string]interface{}{}
	if shopper != nil {
		proxyBid, _ := h.Store.GetActiveProxyBid(shopper.ShopperID, listing.ListingID)
		if proxyBid != nil {
			proxyBidPrice = priceArray(proxyBid.BidAmountUsd)
		}
	}

	// Get seller info for memberId
	seller, _ := h.Store.GetShopper(listing.SellerShopperID)
	sellerMemberID := int64(0)
	if seller != nil {
		sellerMemberID = seller.MemberID
	}

	// Split domain into SLD and TLD
	sld, tld := splitDomain(listing.DomainName)

	return map[string]interface{}{
		"listingId":             listing.ListingID,
		"domainName":            listing.DomainName,
		"listingStatus":         listing.ListingStatus,
		"listingType":           listing.ListingType,
		"auctionTypeId":         listing.AuctionTypeID,
		"startTime":             toFractionalISO(listing.StartTime),
		"endTime":               toFractionalISO(listing.EndTime),
		"lastUpdatedTime":       lifecycle.Now().Format(time.RFC3339),
		"listDate":              toFractionalISO(listing.StartTime),
		"description":           "",
		"domainId":              0,
		"domainCreateDate":      "",
		"sld":                   sld,
		"tld":                   tld,
		"age":                   0,
		"traffic":               0,
		"visits":                0,
		"watchers":              0,
		"priceType":             priceTypeForListing(listing),
		"eventName":             "",
		"expireDate":            "",
		"estimatedTransferTime": "",
		"shopperId":             listing.SellerShopperID,
		"memberId":              sellerMemberID,
		"isPinned":              false,
		"isWatching":            false,
		"isAutoExtended":        listing.IsAutoExtended,
		"renewalPfid":           0,
		"estimatedValueRank":    0,
		"memberBiddingStatus":   memberBiddingStatus,
		"memberOfferStatus":     "",
		"biddersCount":          listing.BiddersCount,
		"bidsOrOffersCount":     listing.BidsCount,
		"askingPrice":           priceArray(listing.AskingPriceUsd),
		"currentPrice":          priceArray(listing.CurrentPriceUsd),
		"nextBidPrice":          nextBidPriceArray(listing.NextBidPriceUsd),
		"minBidOrOfferPrice":    priceArray(listing.AskingPriceUsd),
		"buyItNowPrice":         []interface{}{},
		"estimatedValue":        []interface{}{},
		"estimatedValueRange":   map[string]interface{}{"min": []interface{}{}, "max": []interface{}{}},
		"proxyBidPrice":         proxyBidPrice,
		"reservedPrice":         []interface{}{},
		"renewalPrice":          []interface{}{},
		"transferPrice":         []interface{}{},
		"parkingRevenue":        []interface{}{},
		"bidHistory":            bidHistory,
		"offerHistory":          []interface{}{},
		"categories":            []interface{}{},
	}
}

// priceArray returns [{cost: micros, currency: "USD"}] or [] if micros <= 0
func priceArray(micros int64) []map[string]interface{} {
	if micros <= 0 {
		return []map[string]interface{}{}
	}
	return []map[string]interface{}{
		{"cost": micros, "currency": "USD"},
	}
}

// nextBidPriceArray returns [{cost: float64(micros), currency: "USD"}]
// nextBidPrice uses Float in the Android model (NextBidPrice.kt)
func nextBidPriceArray(micros int64) []map[string]interface{} {
	if micros <= 0 {
		return []map[string]interface{}{}
	}
	return []map[string]interface{}{
		{"cost": float64(micros), "currency": "USD"},
	}
}

func splitDomain(domain string) (string, string) {
	parts := strings.SplitN(domain, ".", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return domain, ""
}
