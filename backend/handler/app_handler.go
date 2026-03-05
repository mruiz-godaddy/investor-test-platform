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
		WriteError(w, "MISSING_SHOPPER", err.Error(), 400)
		return
	}

	allListings, _ := h.Store.ListListings()
	var result []map[string]interface{}
	for _, listing := range allListings {
		hasBid, _ := h.Store.HasShopperBidOnListing(shopper.ShopperID, listing.ListingID)
		if hasBid {
			result = append(result, h.buildListingJSON(&listing, shopper))
		}
	}

	// No local results → forward entirely to upstream
	if len(result) == 0 && h.AuctionUpstream != "" {
		if proxyToUpstream(h.AuctionUpstream, w, r) {
			return
		}
	}

	// Local results exist → merge with upstream (local wins on dupes)
	if len(result) > 0 && h.AuctionUpstream != "" {
		upstreamItems, err := fetchUpstreamListings(h.AuctionUpstream, "/v1/aftermarket/domains/bidding", r)
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

	listings, _ := h.Store.GetWonListingsForShopper(shopper.ShopperID)
	var result []map[string]interface{}
	for _, l := range listings {
		result = append(result, h.buildWonListingJSON(&l))
	}

	// No local results → forward entirely to upstream
	if len(result) == 0 && h.AuctionUpstream != "" {
		if proxyToUpstream(h.AuctionUpstream, w, r) {
			return
		}
	}

	// Local results exist → merge with upstream (local wins on dupes)
	if len(result) > 0 && h.AuctionUpstream != "" {
		upstreamItems, err := fetchUpstreamListings(h.AuctionUpstream, "/v1/aftermarket/domains/won", r)
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
func (h *AppHandler) SearchListings(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")

	// Only search local DB when there's an actual query
	var result []map[string]interface{}
	if query != "" {
		listings, _ := h.Store.SearchListingsByDomain(query)
		for _, l := range listings {
			result = append(result, h.buildSearchResultJSON(&l))
		}
	}

	// If no local results, try forwarding entirely to find-upstream
	if len(result) == 0 && h.FindUpstream != "" {
		if proxyToUpstream(h.FindUpstream, w, r) {
			return
		}
	}

	// If we have local results, merge with upstream (local wins on dupes)
	if len(result) > 0 && h.FindUpstream != "" {
		upstreamItems, err := fetchUpstreamSearchResults(h.FindUpstream, r)
		if err == nil && upstreamItems != nil {
			result = mergeListings(result, upstreamItems)
		}
	}

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
		"buy_it_now_amount":     0,
		"buy_it_now_amount_usd": 0,
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
		"buy_it_now_amount_display":     "",
		"buy_it_now_amount_display_usd": "",
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
		"buy_it_now_flag":      false,
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
		"priceType":             "",
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
