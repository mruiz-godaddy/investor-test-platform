package scenario

import (
	"fmt"
	"time"

	"backend/bidding"
	"backend/config"
	"backend/lifecycle"
	"backend/model"
	"backend/store"
)

type Loader struct {
	Store  *store.Store
	Config *config.Config
	Engine *bidding.Engine
}

func NewLoader(s *store.Store, cfg *config.Config, eng *bidding.Engine) *Loader {
	return &Loader{Store: s, Config: cfg, Engine: eng}
}

// Load resets the database and clock, then dispatches to the named scenario.
func (l *Loader) Load(name string) (map[string]interface{}, error) {
	l.Store.Reset()
	lifecycle.Reset()

	switch name {
	case "normal_auction":
		return l.normalAuction()
	case "sniper_bid":
		return l.sniperBid()
	case "race_condition":
		return l.raceCondition()
	case "auto_extend_chain":
		return l.autoExtendChain()
	case "delayed_transition":
		return l.delayedTransition()
	case "proxy_outbid":
		return l.proxyOutbid()
	case "proxy_stack":
		return l.proxyStack()
	case "proxy_burn":
		return l.proxyBurn()
	default:
		return nil, fmt.Errorf("unknown scenario: %s", name)
	}
}

func buildResult(name, description string, cfg *config.Config, shoppers []map[string]interface{}, listings []map[string]interface{}, bidsPlaced int) map[string]interface{} {
	return map[string]interface{}{
		"scenario":    name,
		"description": description,
		"config":      cfg.Snapshot(),
		"shoppers":    shoppers,
		"listings":    listings,
		"bidsPlaced":  bidsPlaced,
	}
}

// 9.1 normal_auction — Standard auction ending on time, transitions to SOLD.
func (l *Loader) normalAuction() (map[string]interface{}, error) {
	l.Config.SetAutoFinalize(true)
	l.Config.SetStatusTransitionDelayMs(0)

	l.Store.CreateShopper(model.Shopper{
		ShopperID: "shopper-buyer-2", MemberID: 10003, CustomerID: "cust-buyer-2",
	})

	now := lifecycle.Now()
	listingID, _ := l.Store.CreateListing(model.Listing{
		DomainName:       "credittip.com",
		ListingStatus:    model.StatusOpen,
		ListingType:      "EXPIRY_AUCTIONS",
		AuctionTypeID:    16,
		StartTime:        now.Format(time.RFC3339),
		EndTime:          now.Add(2 * time.Minute).Format(time.RFC3339),
		AskingPriceUsd:   5_000_000,
		SellerShopperID:  "shopper-seller",
		AutoExtEnabled:   false,
		AutoExtWindowSec: 60,
		AutoExtSeconds:   300,
	})

	bidsPlaced := 0
	l.Engine.PlaceBid(bidding.BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer",
		UsdBidAmount: 5_000_000, IsTosAccepted: true,
	})
	bidsPlaced++

	l.Engine.PlaceBid(bidding.BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer-2",
		UsdBidAmount: 10_000_000, IsTosAccepted: true,
	})
	bidsPlaced++

	listing, _ := l.Store.GetListing(listingID)
	return buildResult("normal_auction",
		"Standard auction ending on time, transitions to SOLD",
		l.Config,
		[]map[string]interface{}{
			{"shopperId": "shopper-seller", "memberId": 10001},
			{"shopperId": "shopper-buyer", "memberId": 10002},
			{"shopperId": "shopper-buyer-2", "memberId": 10003},
		},
		[]map[string]interface{}{
			{"listingId": listing.ListingID, "domainName": listing.DomainName, "endTime": listing.EndTime},
		},
		bidsPlaced,
	), nil
}

// 9.2 sniper_bid — Sniper bid triggers auto-extension within last 60s.
func (l *Loader) sniperBid() (map[string]interface{}, error) {
	l.Config.SetAutoFinalize(true)
	l.Config.SetStatusTransitionDelayMs(0)

	l.Store.CreateShopper(model.Shopper{
		ShopperID: "shopper-sniper", MemberID: 10003, CustomerID: "cust-sniper", DisplayName: "Sniper",
	})

	now := lifecycle.Now()
	listingID, _ := l.Store.CreateListing(model.Listing{
		DomainName:       "credittip.com",
		ListingStatus:    model.StatusOpen,
		ListingType:      "EXPIRY_AUCTIONS",
		AuctionTypeID:    16,
		StartTime:        now.Format(time.RFC3339),
		EndTime:          now.Add(90 * time.Second).Format(time.RFC3339),
		AskingPriceUsd:   5_000_000,
		SellerShopperID:  "shopper-seller",
		AutoExtEnabled:   true,
		AutoExtWindowSec: 60,
		AutoExtSeconds:   300,
	})

	l.Engine.PlaceBid(bidding.BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer",
		UsdBidAmount: 5_000_000, IsTosAccepted: true,
	})

	listing, _ := l.Store.GetListing(listingID)
	return buildResult("sniper_bid",
		"Sniper bid triggers auto-extension within last 60s",
		l.Config,
		[]map[string]interface{}{
			{"shopperId": "shopper-seller", "memberId": 10001},
			{"shopperId": "shopper-buyer", "memberId": 10002},
			{"shopperId": "shopper-sniper", "memberId": 10003},
		},
		[]map[string]interface{}{
			{"listingId": listing.ListingID, "domainName": listing.DomainName, "endTime": listing.EndTime},
		},
		1,
	), nil
}

// 9.3 race_condition — Reproduces CreditTip.com race condition.
func (l *Loader) raceCondition() (map[string]interface{}, error) {
	l.Config.SetAutoFinalize(false)
	l.Config.SetStatusTransitionDelayMs(0)

	l.Store.CreateShopper(model.Shopper{
		ShopperID: "shopper-sniper", MemberID: 10003, CustomerID: "cust-sniper", DisplayName: "Sniper",
	})

	now := lifecycle.Now()
	listingID, _ := l.Store.CreateListing(model.Listing{
		DomainName:       "credittip.com",
		ListingStatus:    model.StatusOpen,
		ListingType:      "EXPIRY_AUCTIONS",
		AuctionTypeID:    16,
		StartTime:        now.Format(time.RFC3339),
		EndTime:          now.Add(30 * time.Second).Format(time.RFC3339),
		AskingPriceUsd:   5_000_000,
		SellerShopperID:  "shopper-seller",
		AutoExtEnabled:   false,
		AutoExtWindowSec: 60,
		AutoExtSeconds:   300,
	})

	l.Engine.PlaceBid(bidding.BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer",
		UsdBidAmount: 5_000_000, IsTosAccepted: true,
	})

	listing, _ := l.Store.GetListing(listingID)
	return buildResult("race_condition",
		"Reproduces CreditTip.com race condition — autoFinalize disabled, sniper can bid after endTime",
		l.Config,
		[]map[string]interface{}{
			{"shopperId": "shopper-seller", "memberId": 10001},
			{"shopperId": "shopper-buyer", "memberId": 10002},
			{"shopperId": "shopper-sniper", "memberId": 10003},
		},
		[]map[string]interface{}{
			{"listingId": listing.ListingID, "domainName": listing.DomainName, "endTime": listing.EndTime},
		},
		1,
	), nil
}

// 9.4 auto_extend_chain — Multiple snipers trigger chained auto-extensions.
func (l *Loader) autoExtendChain() (map[string]interface{}, error) {
	l.Config.SetAutoFinalize(true)
	l.Config.SetStatusTransitionDelayMs(2000)

	for _, s := range []model.Shopper{
		{ShopperID: "shopper-sniper-a", MemberID: 10003, CustomerID: "cust-sniper-a", DisplayName: "Sniper A"},
		{ShopperID: "shopper-sniper-b", MemberID: 10004, CustomerID: "cust-sniper-b", DisplayName: "Sniper B"},
		{ShopperID: "shopper-sniper-c", MemberID: 10005, CustomerID: "cust-sniper-c", DisplayName: "Sniper C"},
	} {
		l.Store.CreateShopper(s)
	}

	now := lifecycle.Now()
	listingID, _ := l.Store.CreateListing(model.Listing{
		DomainName:       "credittip.com",
		ListingStatus:    model.StatusOpen,
		ListingType:      "EXPIRY_AUCTIONS",
		AuctionTypeID:    16,
		StartTime:        now.Format(time.RFC3339),
		EndTime:          now.Add(90 * time.Second).Format(time.RFC3339),
		AskingPriceUsd:   5_000_000,
		SellerShopperID:  "shopper-seller",
		AutoExtEnabled:   true,
		AutoExtWindowSec: 60,
		AutoExtSeconds:   120,
	})

	l.Engine.PlaceBid(bidding.BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer",
		UsdBidAmount: 5_000_000, IsTosAccepted: true,
	})

	listing, _ := l.Store.GetListing(listingID)
	return buildResult("auto_extend_chain",
		"Multiple snipers trigger chained auto-extensions (120s each)",
		l.Config,
		[]map[string]interface{}{
			{"shopperId": "shopper-seller", "memberId": 10001},
			{"shopperId": "shopper-buyer", "memberId": 10002},
			{"shopperId": "shopper-sniper-a", "memberId": 10003},
			{"shopperId": "shopper-sniper-b", "memberId": 10004},
			{"shopperId": "shopper-sniper-c", "memberId": 10005},
		},
		[]map[string]interface{}{
			{"listingId": listing.ListingID, "domainName": listing.DomainName, "endTime": listing.EndTime},
		},
		1,
	), nil
}

// 9.6 delayed_transition — Listing stays OPEN for 10s after endTime.
func (l *Loader) delayedTransition() (map[string]interface{}, error) {
	l.Config.SetAutoFinalize(true)
	l.Config.SetStatusTransitionDelayMs(10000)

	now := lifecycle.Now()
	listingID, _ := l.Store.CreateListing(model.Listing{
		DomainName:       "example.com",
		ListingStatus:    model.StatusOpen,
		ListingType:      "EXPIRY_AUCTIONS",
		AuctionTypeID:    16,
		StartTime:        now.Format(time.RFC3339),
		EndTime:          now.Add(30 * time.Second).Format(time.RFC3339),
		AskingPriceUsd:   5_000_000,
		SellerShopperID:  "shopper-seller",
		AutoExtEnabled:   false,
	})

	l.Engine.PlaceBid(bidding.BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer",
		UsdBidAmount: 10_000_000, IsTosAccepted: true,
	})

	listing, _ := l.Store.GetListing(listingID)
	return buildResult("delayed_transition",
		"Listing stays OPEN for 10s after endTime, then transitions to SOLD",
		l.Config,
		[]map[string]interface{}{
			{"shopperId": "shopper-seller", "memberId": 10001},
			{"shopperId": "shopper-buyer", "memberId": 10002},
		},
		[]map[string]interface{}{
			{"listingId": listing.ListingID, "domainName": listing.DomainName, "endTime": listing.EndTime},
		},
		1,
	), nil
}

// proxy_outbid — Proxy holder auto-outbids a competitor.
func (l *Loader) proxyOutbid() (map[string]interface{}, error) {
	l.Config.SetAutoFinalize(true)
	l.Config.SetStatusTransitionDelayMs(0)

	for _, s := range []model.Shopper{
		{ShopperID: "shopper-buyer-a", MemberID: 10003, CustomerID: "cust-buyer-a"},
		{ShopperID: "shopper-buyer-b", MemberID: 10004, CustomerID: "cust-buyer-b"},
	} {
		l.Store.CreateShopper(s)
	}

	now := lifecycle.Now()
	listingID, _ := l.Store.CreateListing(model.Listing{
		DomainName:      "proxy-test.com",
		ListingStatus:   model.StatusOpen,
		ListingType:     "EXPIRY_AUCTIONS",
		AuctionTypeID:   16,
		StartTime:       now.Format(time.RFC3339),
		EndTime:         now.Add(5 * time.Minute).Format(time.RFC3339),
		AskingPriceUsd:  5_000_000, // $5
		SellerShopperID: "shopper-seller",
		AutoExtEnabled:  false,
	})

	bidsPlaced := 0

	// buyer-A bids $20 → creates PROXY($20) + AUCTION($5)
	l.Engine.PlaceBid(bidding.BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer-a",
		UsdBidAmount: 20_000_000, IsTosAccepted: true,
	})
	bidsPlaced++

	// buyer-B bids $10 → proxy auto-creates AUCTION for buyer-A, buyer-A stays high
	l.Engine.PlaceBid(bidding.BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer-b",
		UsdBidAmount: 10_000_000, IsTosAccepted: true,
	})
	bidsPlaced++

	listing, _ := l.Store.GetListing(listingID)
	return buildResult("proxy_outbid",
		"Proxy holder auto-outbids a competitor — buyer-A's $20 proxy auto-outbids buyer-B's $10 bid",
		l.Config,
		[]map[string]interface{}{
			{"shopperId": "shopper-seller", "memberId": 10001},
			{"shopperId": "shopper-buyer-a", "memberId": 10003},
			{"shopperId": "shopper-buyer-b", "memberId": 10004},
		},
		[]map[string]interface{}{
			{"listingId": listing.ListingID, "domainName": listing.DomainName, "endTime": listing.EndTime,
				"currentPriceUsd": listing.CurrentPriceUsd, "highestBidder": listing.HighestBidderShopper,
				"bidsCount": listing.BidsCount},
		},
		bidsPlaced,
	), nil
}

// proxy_stack — Same customer raises their proxy.
func (l *Loader) proxyStack() (map[string]interface{}, error) {
	l.Config.SetAutoFinalize(true)
	l.Config.SetStatusTransitionDelayMs(0)

	l.Store.CreateShopper(model.Shopper{
		ShopperID: "shopper-buyer-a", MemberID: 10003, CustomerID: "cust-buyer-a",
	})

	now := lifecycle.Now()
	listingID, _ := l.Store.CreateListing(model.Listing{
		DomainName:      "proxy-stack.com",
		ListingStatus:   model.StatusOpen,
		ListingType:     "EXPIRY_AUCTIONS",
		AuctionTypeID:   16,
		StartTime:       now.Format(time.RFC3339),
		EndTime:         now.Add(5 * time.Minute).Format(time.RFC3339),
		AskingPriceUsd:  5_000_000, // $5
		SellerShopperID: "shopper-seller",
		AutoExtEnabled:  false,
	})

	bidsPlaced := 0

	// buyer-A bids $20 → PROXY($20) + AUCTION($5)
	l.Engine.PlaceBid(bidding.BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer-a",
		UsdBidAmount: 20_000_000, IsTosAccepted: true,
	})
	bidsPlaced++

	// buyer-A bids $50 → old proxy cancelled, new PROXY($50), no new AUCTION
	l.Engine.PlaceBid(bidding.BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer-a",
		UsdBidAmount: 50_000_000, IsTosAccepted: true,
	})
	bidsPlaced++

	listing, _ := l.Store.GetListing(listingID)
	return buildResult("proxy_stack",
		"Same customer raises their proxy — old proxy cancelled, new PROXY($50), bidsCount stays 1",
		l.Config,
		[]map[string]interface{}{
			{"shopperId": "shopper-seller", "memberId": 10001},
			{"shopperId": "shopper-buyer-a", "memberId": 10003},
		},
		[]map[string]interface{}{
			{"listingId": listing.ListingID, "domainName": listing.DomainName, "endTime": listing.EndTime,
				"currentPriceUsd": listing.CurrentPriceUsd, "bidsCount": listing.BidsCount},
		},
		bidsPlaced,
	), nil
}

// proxy_burn — Incoming bid exceeds and burns proxy.
func (l *Loader) proxyBurn() (map[string]interface{}, error) {
	l.Config.SetAutoFinalize(true)
	l.Config.SetStatusTransitionDelayMs(0)

	for _, s := range []model.Shopper{
		{ShopperID: "shopper-buyer-a", MemberID: 10003, CustomerID: "cust-buyer-a"},
		{ShopperID: "shopper-buyer-b", MemberID: 10004, CustomerID: "cust-buyer-b"},
	} {
		l.Store.CreateShopper(s)
	}

	now := lifecycle.Now()
	listingID, _ := l.Store.CreateListing(model.Listing{
		DomainName:      "proxy-burn.com",
		ListingStatus:   model.StatusOpen,
		ListingType:     "EXPIRY_AUCTIONS",
		AuctionTypeID:   16,
		StartTime:       now.Format(time.RFC3339),
		EndTime:         now.Add(5 * time.Minute).Format(time.RFC3339),
		AskingPriceUsd:  5_000_000, // $5
		SellerShopperID: "shopper-seller",
		AutoExtEnabled:  false,
	})

	bidsPlaced := 0

	// buyer-A bids $12 → PROXY($12) + AUCTION($5)
	l.Engine.PlaceBid(bidding.BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer-a",
		UsdBidAmount: 12_000_000, IsTosAccepted: true,
	})
	bidsPlaced++

	// buyer-B bids $15 → proxy burned, childBid at $12, buyer-B wins at $15
	l.Engine.PlaceBid(bidding.BidRequest{
		ListingID: listingID, ShopperID: "shopper-buyer-b",
		UsdBidAmount: 15_000_000, IsTosAccepted: true,
	})
	bidsPlaced++

	listing, _ := l.Store.GetListing(listingID)
	return buildResult("proxy_burn",
		"Incoming bid exceeds and burns proxy — buyer-B's $15 beats buyer-A's $12 proxy",
		l.Config,
		[]map[string]interface{}{
			{"shopperId": "shopper-seller", "memberId": 10001},
			{"shopperId": "shopper-buyer-a", "memberId": 10003},
			{"shopperId": "shopper-buyer-b", "memberId": 10004},
		},
		[]map[string]interface{}{
			{"listingId": listing.ListingID, "domainName": listing.DomainName, "endTime": listing.EndTime,
				"currentPriceUsd": listing.CurrentPriceUsd, "highestBidder": listing.HighestBidderShopper,
				"bidsCount": listing.BidsCount},
		},
		bidsPlaced,
	), nil
}
