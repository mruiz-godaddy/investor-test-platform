package handler

import (
	"encoding/json"
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
	Store  *store.Store
	Config *config.Config
	Engine *bidding.Engine
}

func NewAppHandler(s *store.Store, cfg *config.Config, eng *bidding.Engine) *AppHandler {
	return &AppHandler{Store: s, Config: cfg, Engine: eng}
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
			"bidDate":           bid.CreatedAt,
			"bidExpirationDate": listing.EndTime,
			"bidder":            memberID,
			"comment":           "",
		})
	}

	// Compute memberBiddingStatus
	memberBiddingStatus := ""
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
		"startTime":             listing.StartTime,
		"endTime":               listing.EndTime,
		"lastUpdatedTime":       lifecycle.Now().Format(time.RFC3339),
		"listDate":              listing.StartTime,
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
		"isReserveMet":          listing.IsReserveMet,
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
