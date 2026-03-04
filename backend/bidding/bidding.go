package bidding

import (
	"log"
	"time"

	"github.com/google/uuid"

	"backend/lifecycle"
	"backend/model"
	"backend/store"
)

const MuMultiplier = 1_000_000

func GetBidIncrement(currentPriceUsdMicros int64) int64 {
	priceDollars := currentPriceUsdMicros / MuMultiplier
	switch {
	case priceDollars <= 499:
		return 5 * MuMultiplier
	case priceDollars <= 999:
		return 10 * MuMultiplier
	case priceDollars <= 2499:
		return 25 * MuMultiplier
	case priceDollars <= 4999:
		return 50 * MuMultiplier
	case priceDollars <= 9999:
		return 100 * MuMultiplier
	case priceDollars <= 24999:
		return 250 * MuMultiplier
	case priceDollars <= 49999:
		return 500 * MuMultiplier
	default:
		return 1000 * MuMultiplier
	}
}

// BidError represents a bidding validation error with HTTP status and JSON fields.
type BidError struct {
	Code       string // error code for JSON response
	Message    string // human-readable message
	HTTPStatus int    // HTTP status code
}

func (e *BidError) Error() string {
	return e.Message
}

var (
	ErrListingNotFound = &BidError{"LISTING_NOT_FOUND", "Listing not found", 404}
	ErrListingClosed   = &BidError{"LISTING_CLOSED", "Listing is closed or expired", 422}
	ErrBidderIsSeller  = &BidError{"BIDDER_IS_SELLER", "Seller cannot bid on their own listing", 422}
	ErrBidTooLow       = &BidError{"BID_IS_LESS_THAN_STARTING_AMT", "That bid is less than the starting amount", 422}
	ErrTosNotAccepted  = &BidError{"USER_MUST_AGREE_TO_TOS", "You must agree to the terms of service", 422}
	ErrNonBiddableType = &BidError{"NON_BIDDABLE_TYPE", "This listing type does not support bidding", 422}
	ErrServerError     = &BidError{"SERVER_ERROR", "Internal server error", 500}
)

type Engine struct {
	Store *store.Store
}

func NewEngine(s *store.Store) *Engine {
	return &Engine{Store: s}
}

type BidRequest struct {
	ListingID     int64
	ShopperID     string
	UsdBidAmount  int64 // micros
	IsTosAccepted bool
}

type BidResult struct {
	ListingID       int64  `json:"listingId"`
	BidID           string `json:"bidId"`
	BidAmountUsd    int64  `json:"bidAmountUsd"`
	IsHighestBidder bool   `json:"isHighestBidder"`
	Status          string `json:"status"`
}

// PlaceBid validates and places a bid through the full 6-step validation.
func (e *Engine) PlaceBid(req BidRequest) (*BidResult, error) {
	// Step 1: Required fields
	if req.UsdBidAmount <= 0 {
		return nil, ErrBidTooLow
	}
	if !req.IsTosAccepted {
		return nil, ErrTosNotAccepted
	}

	// Step 2: Listing exists
	listing, err := e.Store.GetListing(req.ListingID)
	if err != nil {
		return nil, ErrServerError
	}
	if listing == nil {
		return nil, ErrListingNotFound
	}

	// Step 3: Listing is OPEN
	if listing.ListingStatus != model.StatusOpen {
		return nil, ErrListingClosed
	}

	// Step 4: End time not passed
	endTime, _ := time.Parse(time.RFC3339, listing.EndTime)
	if lifecycle.Now().After(endTime) {
		return nil, ErrListingClosed
	}

	// Step 5: Not the seller
	if req.ShopperID == listing.SellerShopperID {
		return nil, ErrBidderIsSeller
	}

	// Step 6: Meets absolute floor (asking price)
	// The proxy-aware minimum-vs-highest-bid check happens inside placeBidInternal.
	if req.UsdBidAmount < listing.AskingPriceUsd {
		return nil, ErrBidTooLow
	}

	return e.placeBidInternal(listing, req.ShopperID, req.UsdBidAmount)
}

// PlaceSniperBid places a bid bypassing TOS and timing validation.
func (e *Engine) PlaceSniperBid(listingID int64, shopperID string, bidAmountUsd int64) (*BidResult, error) {
	if bidAmountUsd <= 0 {
		return nil, ErrBidTooLow
	}

	listing, err := e.Store.GetListing(listingID)
	if err != nil {
		return nil, ErrServerError
	}
	if listing == nil {
		return nil, ErrListingNotFound
	}
	if listing.ListingStatus != model.StatusOpen {
		return nil, ErrListingClosed
	}

	// Reject bids below the asking price floor
	if bidAmountUsd < listing.AskingPriceUsd {
		return nil, ErrBidTooLow
	}

	return e.placeBidInternal(listing, shopperID, bidAmountUsd)
}

// placeBidInternal contains the shared placement logic with proxy bidding support.
// This mirrors the algorithm in auc-bidding/dao/dao.go PlaceBid.
func (e *Engine) placeBidInternal(listing *model.Listing, shopperID string, bidAmountUsd int64) (*BidResult, error) {
	newBidID := uuid.New().String()

	// Phase 1: Gather state
	proxyMap, err := e.Store.GetAllActiveProxies(listing.ListingID)
	if err != nil {
		return nil, ErrServerError
	}
	highestAuction, err := e.Store.GetHighestAuctionBid(listing.ListingID)
	if err != nil {
		return nil, ErrServerError
	}

	var highestAuctionAmt int64
	if highestAuction != nil {
		highestAuctionAmt = highestAuction.BidAmountUsd
	}

	// Phase 2: Validate against existing AUCTION bids
	if highestAuction != nil {
		// Incoming bid must be strictly greater than the highest AUCTION bid
		if bidAmountUsd <= highestAuctionAmt {
			return nil, ErrBidTooLow
		}
		// Must beat highest AUCTION bid by at least one increment
		bidInc := GetBidIncrement(highestAuctionAmt)
		minimumBid := highestAuctionAmt + bidInc
		if bidAmountUsd < minimumBid {
			return nil, ErrBidTooLow
		}
	}

	// If shopper already has a proxy >= incoming bid → BID_MIN_NOT_MET
	if myProxy, found := proxyMap[shopperID]; found {
		if myProxy.BidAmountUsd > bidAmountUsd {
			return nil, ErrBidTooLow
		}
	}

	// Phase 3: Proxy resolution — determine bidsToPlace
	type bidToPlace struct {
		bid       model.Bid
		burnProxy bool // if true, cancel the source proxy after placing
	}
	var bidsToPlace []bidToPlace
	isHighestBidder := false

	// Find the relevant proxy: highest proxy >= highestAuctionAmt (from any shopper)
	var relevantProxy *model.Bid
	for _, proxy := range proxyMap {
		if proxy.BidAmountUsd >= highestAuctionAmt {
			if relevantProxy == nil || proxy.BidAmountUsd > relevantProxy.BidAmountUsd {
				relevantProxy = proxy
			}
		}
	}

	if relevantProxy != nil && relevantProxy.BidAmountUsd >= bidAmountUsd {
		// CASE A: Existing proxy >= incoming bid (proxy auto-outbids)

		if relevantProxy.ShopperID == shopperID && relevantProxy.BidAmountUsd == bidAmountUsd {
			// Same customer, same amount → reject
			return nil, ErrBidTooLow
		}

		// Create child AUCTION bid on behalf of proxy holder
		childBidID := uuid.New().String()
		bidInc := GetBidIncrement(bidAmountUsd)
		childAmt := relevantProxy.BidAmountUsd
		if relevantProxy.BidAmountUsd > bidAmountUsd {
			childAmt = min64(relevantProxy.BidAmountUsd, bidAmountUsd+bidInc)
		}

		burnProxy := relevantProxy.BidAmountUsd <= bidAmountUsd+bidInc

		if relevantProxy.ShopperID != shopperID {
			// Different customer: place incoming bid + child bid from proxy
			bidsToPlace = append(bidsToPlace,
				bidToPlace{bid: model.Bid{
					BidID: newBidID, ListingID: listing.ListingID,
					ShopperID: shopperID, BidAmountUsd: bidAmountUsd,
					BidType: model.BidTypeAuction, BidStatus: model.BidStatusActive,
				}},
				bidToPlace{bid: model.Bid{
					BidID: childBidID, ListingID: listing.ListingID,
					ShopperID: relevantProxy.ShopperID, BidAmountUsd: childAmt,
					BidType: model.BidTypeAuction, BidStatus: model.BidStatusActive,
					ParentBidID: relevantProxy.BidID,
				}, burnProxy: burnProxy},
			)
			// Proxy holder stays high
			isHighestBidder = false
		}
		// If same customer with higher proxy, this is already rejected above or is a no-op

	} else if relevantProxy != nil && bidAmountUsd > relevantProxy.BidAmountUsd && bidAmountUsd > highestAuctionAmt {
		// CASE B: Incoming bid > proxy (incoming wins)

		if relevantProxy.ShopperID == shopperID && bidAmountUsd > relevantProxy.BidAmountUsd {
			// Same customer: stack proxy (cancel old, create new PROXY)
			newProxyBid := model.Bid{
				BidID: uuid.New().String(), ListingID: listing.ListingID,
				ShopperID: shopperID, BidAmountUsd: bidAmountUsd,
				BidType: model.BidTypeProxy, BidStatus: model.BidStatusActive,
			}
			bidsToPlace = append(bidsToPlace,
				bidToPlace{bid: newProxyBid, burnProxy: false},
			)
			// Cancel old proxy
			if err := e.Store.CancelBid(relevantProxy.BidID); err != nil {
				return nil, ErrServerError
			}
			isHighestBidder = true

		} else if relevantProxy.BidAmountUsd > highestAuctionAmt {
			// Different customer: burn proxy, create child bid from proxy
			childBidID := uuid.New().String()
			childBid := model.Bid{
				BidID: childBidID, ListingID: listing.ListingID,
				ShopperID: relevantProxy.ShopperID, BidAmountUsd: relevantProxy.BidAmountUsd,
				BidType: model.BidTypeAuction, BidStatus: model.BidStatusActive,
				ParentBidID: relevantProxy.BidID,
			}

			bidInc := GetBidIncrement(relevantProxy.BidAmountUsd)

			if bidAmountUsd > relevantProxy.BidAmountUsd+bidInc {
				// Incoming bid is large enough for a new proxy
				newProxyBid := model.Bid{
					BidID: uuid.New().String(), ListingID: listing.ListingID,
					ShopperID: shopperID, BidAmountUsd: bidAmountUsd,
					BidType: model.BidTypeProxy, BidStatus: model.BidStatusActive,
				}
				newAuctionBid := model.Bid{
					BidID: newBidID, ListingID: listing.ListingID,
					ShopperID: shopperID, BidAmountUsd: relevantProxy.BidAmountUsd + bidInc,
					BidType: model.BidTypeAuction, BidStatus: model.BidStatusActive,
					ParentBidID: newProxyBid.BidID,
				}
				bidsToPlace = append(bidsToPlace,
					bidToPlace{bid: newProxyBid},
					bidToPlace{bid: childBid, burnProxy: true},
					bidToPlace{bid: newAuctionBid},
				)
			} else {
				// Incoming bid just barely beats proxy
				bidsToPlace = append(bidsToPlace,
					bidToPlace{bid: model.Bid{
						BidID: newBidID, ListingID: listing.ListingID,
						ShopperID: shopperID, BidAmountUsd: bidAmountUsd,
						BidType: model.BidTypeAuction, BidStatus: model.BidStatusActive,
					}},
					bidToPlace{bid: childBid, burnProxy: true},
				)
			}
			isHighestBidder = true
		}
	}

	if len(bidsToPlace) == 0 {
		// CASE C: No relevant proxy (first bid or no active proxies above highest auction)
		currentPrice := listing.AskingPriceUsd
		if highestAuction != nil && highestAuctionAmt >= currentPrice {
			bidInc := GetBidIncrement(highestAuctionAmt)
			currentPrice = highestAuctionAmt + bidInc
		}

		if bidAmountUsd > currentPrice {
			// Create PROXY + AUCTION pair
			proxyBid := model.Bid{
				BidID: uuid.New().String(), ListingID: listing.ListingID,
				ShopperID: shopperID, BidAmountUsd: bidAmountUsd,
				BidType: model.BidTypeProxy, BidStatus: model.BidStatusActive,
			}

			// Check if same customer already holds highest auction — just stack proxy
			if highestAuction != nil && highestAuction.ShopperID == shopperID {
				bidsToPlace = append(bidsToPlace, bidToPlace{bid: proxyBid})
			} else {
				auctionAmt := currentPrice
				// If bid >= reserve and reserve > 0 and no existing bid at reserve
				if listing.ReservePriceUsd > 0 && bidAmountUsd >= listing.ReservePriceUsd &&
					(highestAuction == nil || highestAuctionAmt < listing.ReservePriceUsd) {
					auctionAmt = listing.ReservePriceUsd
				}
				auctionBid := model.Bid{
					BidID: newBidID, ListingID: listing.ListingID,
					ShopperID: shopperID, BidAmountUsd: auctionAmt,
					BidType: model.BidTypeAuction, BidStatus: model.BidStatusActive,
					ParentBidID: proxyBid.BidID,
				}
				bidsToPlace = append(bidsToPlace,
					bidToPlace{bid: proxyBid},
					bidToPlace{bid: auctionBid},
				)
			}
			isHighestBidder = true

		} else if bidAmountUsd == currentPrice {
			// Exact asking price — AUCTION only, no proxy
			if highestAuction != nil && highestAuction.ShopperID == shopperID {
				// Same customer at current price — just create proxy
				proxyBid := model.Bid{
					BidID: uuid.New().String(), ListingID: listing.ListingID,
					ShopperID: shopperID, BidAmountUsd: bidAmountUsd,
					BidType: model.BidTypeProxy, BidStatus: model.BidStatusActive,
				}
				bidsToPlace = append(bidsToPlace, bidToPlace{bid: proxyBid})
			} else {
				bidsToPlace = append(bidsToPlace, bidToPlace{bid: model.Bid{
					BidID: newBidID, ListingID: listing.ListingID,
					ShopperID: shopperID, BidAmountUsd: bidAmountUsd,
					BidType: model.BidTypeAuction, BidStatus: model.BidStatusActive,
				}})
			}
			isHighestBidder = true

		} else {
			return nil, ErrBidTooLow
		}
	}

	// Phase 4: Persist all bids, update listing
	if err := e.Store.ClearHighBid(listing.ListingID); err != nil {
		return nil, ErrServerError
	}

	var lastAuctionBid *model.Bid
	newAuctionCount := 0
	for i := range bidsToPlace {
		btp := &bidsToPlace[i]
		if btp.burnProxy {
			if err := e.Store.CancelBid(btp.bid.ParentBidID); err != nil {
				return nil, ErrServerError
			}
		}
		if err := e.Store.CreateBid(btp.bid); err != nil {
			return nil, ErrServerError
		}
		if btp.bid.BidType == model.BidTypeAuction {
			newAuctionCount++
			b := btp.bid
			lastAuctionBid = &b
		}
	}

	// Set high bid on the final AUCTION bid
	if lastAuctionBid != nil {
		if err := e.Store.SetHighBid(lastAuctionBid.BidID); err != nil {
			return nil, ErrServerError
		}
	}

	// Determine final state from the actual highest AUCTION bid
	finalHighest, err := e.Store.GetHighestAuctionBid(listing.ListingID)
	if err != nil {
		return nil, ErrServerError
	}

	var finalPrice int64
	var highestBidderShopper string
	if finalHighest != nil {
		finalPrice = finalHighest.BidAmountUsd
		highestBidderShopper = finalHighest.ShopperID
	}

	biddersCount, err := e.Store.GetDistinctBidderCount(listing.ListingID)
	if err != nil {
		return nil, ErrServerError
	}

	nextBidPrice := finalPrice + GetBidIncrement(finalPrice)
	isReserveMet := listing.ReservePriceUsd > 0 && finalPrice >= listing.ReservePriceUsd

	if err := e.Store.UpdateListingAfterBid(
		listing.ListingID,
		finalPrice,
		nextBidPrice,
		highestBidderShopper,
		biddersCount,
		listing.BidsCount+newAuctionCount,
		isReserveMet,
	); err != nil {
		return nil, ErrServerError
	}

	// Phase 5: Auto-extension check
	now := lifecycle.Now()
	endTime, _ := time.Parse(time.RFC3339, listing.EndTime)
	if shouldAutoExtend(listing, now) {
		newEndTime := endTime.Add(time.Duration(listing.AutoExtSeconds) * time.Second)
		if err := e.Store.SetAutoExtended(listing.ListingID, newEndTime.Format(time.RFC3339)); err != nil {
			log.Printf("Failed to auto-extend listing %d: %v", listing.ListingID, err)
		} else {
			log.Printf("Auto-extended listing %d by %d seconds, new endTime: %s",
				listing.ListingID, listing.AutoExtSeconds, newEndTime.Format(time.RFC3339))
		}
	}

	// Phase 6: Return result
	return &BidResult{
		ListingID:       listing.ListingID,
		BidID:           newBidID,
		BidAmountUsd:    bidAmountUsd,
		IsHighestBidder: isHighestBidder,
		Status:          "SUCCESS",
	}, nil
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func shouldAutoExtend(listing *model.Listing, now time.Time) bool {
	if !listing.AutoExtEnabled {
		return false
	}
	endTime, err := time.Parse(time.RFC3339, listing.EndTime)
	if err != nil {
		return false
	}
	windowStart := endTime.Add(-time.Duration(listing.AutoExtWindowSec) * time.Second)
	return now.After(windowStart) && now.Before(endTime)
}
