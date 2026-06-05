package model

const (
	StatusOpen   = "OPEN"
	StatusSold   = "SOLD"
	StatusClosed = "CLOSED"

	BidTypeAuction = "AUCTION"
	BidTypeProxy   = "PROXY"

	BidStatusActive    = "ACTIVE"
	BidStatusCancelled = "CANCELLED"
)

type Shopper struct {
	ShopperID   string `json:"shopperId"`
	MemberID    int64  `json:"memberId"`
	CustomerID  string `json:"customerId"`
	DisplayName string `json:"displayName"`
}

type Listing struct {
	ListingID            int64  `json:"listingId"`
	DomainName           string `json:"domainName"`
	ListingStatus        string `json:"listingStatus"`
	ListingType          string `json:"listingType"`
	AuctionTypeID        int    `json:"auctionTypeId"`
	StartTime            string `json:"startTime"`
	EndTime              string `json:"endTime"`
	AskingPriceUsd       int64  `json:"askingPriceUsd"`
	CurrentPriceUsd      int64  `json:"currentPriceUsd"`
	SalePriceUsd         *int64 `json:"salePriceUsd"`
	NextBidPriceUsd      int64  `json:"nextBidPriceUsd"`
	BiddersCount         int    `json:"biddersCount"`
	BidsCount            int    `json:"bidsCount"`
	IsAutoExtended       bool   `json:"isAutoExtended"`
	SellerShopperID      string `json:"sellerShopperId"`
	HighestBidderShopper string `json:"highestBidderShopper"`
	AutoExtWindowSec     int    `json:"autoExtWindowSec"`
	AutoExtSeconds       int    `json:"autoExtSeconds"`
	AutoExtEnabled       bool   `json:"autoExtEnabled"`
	RadarVisible         bool   `json:"radarVisible"`
	CreatedAt            string `json:"createdAt"`
}

type Bid struct {
	BidID        string `json:"bidId"`
	ListingID    int64  `json:"listingId"`
	ShopperID    string `json:"shopperId"`
	BidAmountUsd int64  `json:"bidAmountUsd"`
	BidType      string `json:"bidType"`
	BidStatus    string `json:"bidStatus"`
	IsHighBid    bool   `json:"isHighBid"`
	ParentBidID  string `json:"parentBidId"`
	CreatedAt    string `json:"createdAt"`
}
