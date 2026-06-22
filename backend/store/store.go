package store

import (
	"database/sql"
	"time"

	"backend/db"
	"backend/model"
)

type Store struct {
	DB *db.DB
}

func New(database *db.DB) *Store {
	return &Store{DB: database}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i != 0
}

const listingColumns = `listing_id, domain_name, listing_status, listing_type, auction_type_id,
	start_time, end_time, asking_price_usd, current_price_usd, sale_price_usd,
	next_bid_price_usd, bidders_count, bids_count, is_auto_extended,
	seller_shopper_id, highest_bidder_shopper, auto_ext_window_sec, auto_ext_seconds,
	auto_ext_enabled, radar_visible, created_at`

// scanListing scans a listing row into a model.Listing.
func scanListing(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.Listing, error) {
	var l model.Listing
	var isAutoExtended, autoExtEnabled, radarVisible int
	var salePrice sql.NullInt64
	var highestBidder sql.NullString

	err := scanner.Scan(
		&l.ListingID,
		&l.DomainName,
		&l.ListingStatus,
		&l.ListingType,
		&l.AuctionTypeID,
		&l.StartTime,
		&l.EndTime,
		&l.AskingPriceUsd,
		&l.CurrentPriceUsd,
		&salePrice,
		&l.NextBidPriceUsd,
		&l.BiddersCount,
		&l.BidsCount,
		&isAutoExtended,
		&l.SellerShopperID,
		&highestBidder,
		&l.AutoExtWindowSec,
		&l.AutoExtSeconds,
		&autoExtEnabled,
		&radarVisible,
		&l.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	l.IsAutoExtended = intToBool(isAutoExtended)
	l.AutoExtEnabled = intToBool(autoExtEnabled)
	l.RadarVisible = intToBool(radarVisible)

	if salePrice.Valid {
		v := salePrice.Int64
		l.SalePriceUsd = &v
	}
	if highestBidder.Valid {
		l.HighestBidderShopper = highestBidder.String
	}

	return &l, nil
}

// scanBid scans a bid row into a model.Bid.
func scanBid(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.Bid, error) {
	var b model.Bid
	var isHighBid int

	err := scanner.Scan(
		&b.BidID,
		&b.ListingID,
		&b.ShopperID,
		&b.BidAmountUsd,
		&b.BidType,
		&b.BidStatus,
		&isHighBid,
		&b.ParentBidID,
		&b.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	b.IsHighBid = intToBool(isHighBid)
	return &b, nil
}

// --- Shopper Operations ---

func (s *Store) CreateShopper(sh model.Shopper) error {
	_, err := s.DB.Conn.Exec(
		`INSERT INTO shoppers (shopper_id, member_id, customer_id, display_name) VALUES (?, ?, ?, ?)`,
		sh.ShopperID, sh.MemberID, sh.CustomerID, sh.DisplayName,
	)
	return err
}

func (s *Store) GetShopper(shopperID string) (*model.Shopper, error) {
	row := s.DB.Conn.QueryRow(
		`SELECT shopper_id, member_id, customer_id, display_name FROM shoppers WHERE shopper_id = ?`,
		shopperID,
	)
	var sh model.Shopper
	err := row.Scan(&sh.ShopperID, &sh.MemberID, &sh.CustomerID, &sh.DisplayName)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sh, nil
}

func (s *Store) ListShoppers() ([]model.Shopper, error) {
	rows, err := s.DB.Conn.Query(
		`SELECT shopper_id, member_id, customer_id, display_name FROM shoppers ORDER BY member_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shoppers []model.Shopper
	for rows.Next() {
		var sh model.Shopper
		if err := rows.Scan(&sh.ShopperID, &sh.MemberID, &sh.CustomerID, &sh.DisplayName); err != nil {
			return nil, err
		}
		shoppers = append(shoppers, sh)
	}
	return shoppers, rows.Err()
}

func (s *Store) GetOrCreateShopper(shopperID string) (*model.Shopper, error) {
	sh, err := s.GetShopper(shopperID)
	if err != nil {
		return nil, err
	}
	if sh != nil {
		return sh, nil
	}

	var nextMemberID int64
	err = s.DB.Conn.QueryRow(`SELECT COALESCE(MAX(member_id), 10000) + 1 FROM shoppers`).Scan(&nextMemberID)
	if err != nil {
		return nil, err
	}

	newShopper := model.Shopper{
		ShopperID:   shopperID,
		MemberID:    nextMemberID,
		CustomerID:  "cust-" + shopperID,
		DisplayName: "",
	}
	if err := s.CreateShopper(newShopper); err != nil {
		return nil, err
	}
	return &newShopper, nil
}

// --- Listing Operations ---

func (s *Store) CreateListing(l model.Listing) (int64, error) {
	result, err := s.DB.Conn.Exec(
		`INSERT INTO listings (
			domain_name, listing_status, listing_type, auction_type_id,
			start_time, end_time, asking_price_usd, current_price_usd,
			sale_price_usd, next_bid_price_usd,
			bidders_count, bids_count, is_auto_extended,
			seller_shopper_id, highest_bidder_shopper,
			auto_ext_window_sec, auto_ext_seconds, auto_ext_enabled, radar_visible
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		l.DomainName, l.ListingStatus, l.ListingType, l.AuctionTypeID,
		l.StartTime, l.EndTime, l.AskingPriceUsd, l.AskingPriceUsd,
		l.SalePriceUsd, l.AskingPriceUsd,
		l.BiddersCount, l.BidsCount, boolToInt(l.IsAutoExtended),
		l.SellerShopperID, sql.NullString{String: l.HighestBidderShopper, Valid: l.HighestBidderShopper != ""},
		l.AutoExtWindowSec, l.AutoExtSeconds, boolToInt(l.AutoExtEnabled), boolToInt(l.RadarVisible),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) GetListing(listingID int64) (*model.Listing, error) {
	row := s.DB.Conn.QueryRow(`SELECT `+listingColumns+` FROM listings WHERE listing_id = ?`, listingID)
	l, err := scanListing(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (s *Store) ListListings() ([]model.Listing, error) {
	rows, err := s.DB.Conn.Query(`SELECT ` + listingColumns + ` FROM listings ORDER BY listing_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []model.Listing
	for rows.Next() {
		l, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		listings = append(listings, *l)
	}
	return listings, rows.Err()
}

func (s *Store) UpdateListingStatus(listingID int64, status string, salePriceUsd int64) error {
	_, err := s.DB.Conn.Exec(
		`UPDATE listings SET listing_status = ?, sale_price_usd = ? WHERE listing_id = ?`,
		status, salePriceUsd, listingID,
	)
	return err
}

func (s *Store) UpdateListingEndTime(listingID int64, endTime string) error {
	_, err := s.DB.Conn.Exec(
		`UPDATE listings SET end_time = ? WHERE listing_id = ?`,
		endTime, listingID,
	)
	return err
}

func (s *Store) UpdateListingAfterBid(listingID int64, currentPrice, nextBidPrice int64, highestBidder string, biddersCount, bidsCount int) error {
	_, err := s.DB.Conn.Exec(
		`UPDATE listings SET current_price_usd = ?, next_bid_price_usd = ?, highest_bidder_shopper = ?, bidders_count = ?, bids_count = ? WHERE listing_id = ?`,
		currentPrice, nextBidPrice, highestBidder, biddersCount, bidsCount, listingID,
	)
	return err
}

func (s *Store) SetAutoExtended(listingID int64, newEndTime string) error {
	_, err := s.DB.Conn.Exec(
		`UPDATE listings SET end_time = ?, is_auto_extended = 1 WHERE listing_id = ?`,
		newEndTime, listingID,
	)
	return err
}

func (s *Store) GetOpenListingsPastEndTime(now time.Time) ([]model.Listing, error) {
	rows, err := s.DB.Conn.Query(
		`SELECT `+listingColumns+` FROM listings WHERE listing_status = 'OPEN' AND end_time <= ?`,
		now.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []model.Listing
	for rows.Next() {
		l, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		listings = append(listings, *l)
	}
	return listings, rows.Err()
}

// --- Bid Operations ---

func (s *Store) CreateBid(b model.Bid) error {
	_, err := s.DB.Conn.Exec(
		`INSERT INTO bids (bid_id, listing_id, shopper_id, bid_amount_usd, bid_type, bid_status, is_high_bid, parent_bid_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		b.BidID, b.ListingID, b.ShopperID, b.BidAmountUsd, b.BidType, b.BidStatus, boolToInt(b.IsHighBid), b.ParentBidID,
	)
	return err
}

func (s *Store) ClearHighBid(listingID int64) error {
	_, err := s.DB.Conn.Exec(
		`UPDATE bids SET is_high_bid = 0 WHERE listing_id = ? AND is_high_bid = 1`,
		listingID,
	)
	return err
}

func (s *Store) GetBidsForListing(listingID int64) ([]model.Bid, error) {
	rows, err := s.DB.Conn.Query(
		`SELECT * FROM bids WHERE listing_id = ? ORDER BY created_at DESC`,
		listingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []model.Bid
	for rows.Next() {
		b, err := scanBid(rows)
		if err != nil {
			return nil, err
		}
		bids = append(bids, *b)
	}
	return bids, rows.Err()
}

func (s *Store) GetActiveBidsForListing(listingID int64) ([]model.Bid, error) {
	rows, err := s.DB.Conn.Query(
		`SELECT * FROM bids WHERE listing_id = ? AND bid_type = 'AUCTION' AND bid_status = 'ACTIVE' ORDER BY created_at DESC`,
		listingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []model.Bid
	for rows.Next() {
		b, err := scanBid(rows)
		if err != nil {
			return nil, err
		}
		bids = append(bids, *b)
	}
	return bids, rows.Err()
}

func (s *Store) GetDistinctBidderCount(listingID int64) (int, error) {
	var count int
	err := s.DB.Conn.QueryRow(
		`SELECT COUNT(DISTINCT shopper_id) FROM bids WHERE listing_id = ? AND bid_status = 'ACTIVE' AND bid_type = 'AUCTION'`,
		listingID,
	).Scan(&count)
	return count, err
}

func (s *Store) HasShopperBidOnListing(shopperID string, listingID int64) (bool, error) {
	var count int
	err := s.DB.Conn.QueryRow(
		`SELECT COUNT(*) FROM bids WHERE listing_id = ? AND shopper_id = ? AND bid_status = 'ACTIVE'`,
		listingID, shopperID,
	).Scan(&count)
	return count > 0, err
}

func (s *Store) GetActiveProxyBid(shopperID string, listingID int64) (*model.Bid, error) {
	row := s.DB.Conn.QueryRow(
		`SELECT * FROM bids WHERE listing_id = ? AND shopper_id = ? AND bid_type = 'PROXY' AND bid_status = 'ACTIVE' ORDER BY bid_amount_usd DESC LIMIT 1`,
		listingID, shopperID,
	)
	b, err := scanBid(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Store) CancelBid(bidID string) error {
	_, err := s.DB.Conn.Exec(
		`UPDATE bids SET bid_status = 'CANCELLED' WHERE bid_id = ?`,
		bidID,
	)
	return err
}

func (s *Store) SetHighBid(bidID string) error {
	_, err := s.DB.Conn.Exec(
		`UPDATE bids SET is_high_bid = 1 WHERE bid_id = ?`,
		bidID,
	)
	return err
}

// GetHighestAuctionBid returns the highest ACTIVE AUCTION bid for a listing.
// Tie-break: earliest created_at wins (matches real auc-bidding behavior).
func (s *Store) GetHighestAuctionBid(listingID int64) (*model.Bid, error) {
	row := s.DB.Conn.QueryRow(
		`SELECT * FROM bids WHERE listing_id = ? AND bid_type = 'AUCTION' AND bid_status = 'ACTIVE' ORDER BY bid_amount_usd DESC, created_at ASC LIMIT 1`,
		listingID,
	)
	b, err := scanBid(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GetAllActiveProxies returns one highest active PROXY bid per shopperID for a listing.
func (s *Store) GetAllActiveProxies(listingID int64) (map[string]*model.Bid, error) {
	rows, err := s.DB.Conn.Query(
		`SELECT * FROM bids WHERE listing_id = ? AND bid_type = 'PROXY' AND bid_status = 'ACTIVE' ORDER BY bid_amount_usd DESC`,
		listingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]*model.Bid)
	for rows.Next() {
		b, err := scanBid(rows)
		if err != nil {
			return nil, err
		}
		// Keep only the highest proxy per shopper (first seen since ordered DESC)
		if _, found := result[b.ShopperID]; !found {
			result[b.ShopperID] = b
		}
	}
	return result, rows.Err()
}

// GetBidsForShopper returns all bids placed by a given shopper.
func (s *Store) GetBidsForShopper(shopperID string) ([]model.Bid, error) {
	rows, err := s.DB.Conn.Query(
		`SELECT * FROM bids WHERE shopper_id = ? ORDER BY created_at DESC`,
		shopperID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []model.Bid
	for rows.Next() {
		b, err := scanBid(rows)
		if err != nil {
			return nil, err
		}
		bids = append(bids, *b)
	}
	return bids, rows.Err()
}

// ListAllBids returns all bids in the database.
func (s *Store) ListAllBids() ([]model.Bid, error) {
	rows, err := s.DB.Conn.Query(`SELECT * FROM bids ORDER BY listing_id, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []model.Bid
	for rows.Next() {
		b, err := scanBid(rows)
		if err != nil {
			return nil, err
		}
		bids = append(bids, *b)
	}
	return bids, rows.Err()
}

// --- Cart Event Operations ---

func (s *Store) CreateCartEvent(e model.CartEvent) (int64, error) {
	result, err := s.DB.Conn.Exec(
		`INSERT INTO cart_events (domain_name, listing_id, inventory_type, itc_code, itc_inventory, area, request_price)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.DomainName, e.ListingID, e.InventoryType, e.ItcCode, e.ItcInventory, e.Area, e.RequestPrice,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) ListCartEvents() ([]model.CartEvent, error) {
	rows, err := s.DB.Conn.Query(
		`SELECT event_id, domain_name, listing_id, inventory_type, itc_code, itc_inventory, area, request_price, created_at
		 FROM cart_events ORDER BY event_id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.CartEvent
	for rows.Next() {
		var e model.CartEvent
		if err := rows.Scan(
			&e.EventID, &e.DomainName, &e.ListingID, &e.InventoryType,
			&e.ItcCode, &e.ItcInventory, &e.Area, &e.RequestPrice, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (s *Store) ClearCartEvents() error {
	_, err := s.DB.Conn.Exec(`DELETE FROM cart_events`)
	return err
}

// ImportShopper inserts a shopper with exact field values (for DB import).
func (s *Store) ImportShopper(sh model.Shopper) error {
	_, err := s.DB.Conn.Exec(
		`INSERT OR REPLACE INTO shoppers (shopper_id, member_id, customer_id, display_name) VALUES (?, ?, ?, ?)`,
		sh.ShopperID, sh.MemberID, sh.CustomerID, sh.DisplayName,
	)
	return err
}

// ImportListing inserts a listing with exact field values including listing_id (for DB import).
func (s *Store) ImportListing(l model.Listing) error {
	_, err := s.DB.Conn.Exec(
		`INSERT OR REPLACE INTO listings (
			listing_id, domain_name, listing_status, listing_type, auction_type_id,
			start_time, end_time, asking_price_usd, current_price_usd,
			sale_price_usd, next_bid_price_usd,
			bidders_count, bids_count, is_auto_extended,
			seller_shopper_id, highest_bidder_shopper,
			auto_ext_window_sec, auto_ext_seconds, auto_ext_enabled, radar_visible, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		l.ListingID, l.DomainName, l.ListingStatus, l.ListingType, l.AuctionTypeID,
		l.StartTime, l.EndTime, l.AskingPriceUsd, l.CurrentPriceUsd,
		l.SalePriceUsd, l.NextBidPriceUsd,
		l.BiddersCount, l.BidsCount, boolToInt(l.IsAutoExtended),
		l.SellerShopperID, sql.NullString{String: l.HighestBidderShopper, Valid: l.HighestBidderShopper != ""},
		l.AutoExtWindowSec, l.AutoExtSeconds, boolToInt(l.AutoExtEnabled), boolToInt(l.RadarVisible), l.CreatedAt,
	)
	return err
}

// ImportBid inserts a bid with exact field values (for DB import).
func (s *Store) ImportBid(b model.Bid) error {
	_, err := s.DB.Conn.Exec(
		`INSERT OR REPLACE INTO bids (bid_id, listing_id, shopper_id, bid_amount_usd, bid_type, bid_status, is_high_bid, parent_bid_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.BidID, b.ListingID, b.ShopperID, b.BidAmountUsd, b.BidType, b.BidStatus, boolToInt(b.IsHighBid), b.ParentBidID, b.CreatedAt,
	)
	return err
}

// GetWonListingsForShopper returns listings where status=SOLD, shopper is the highest bidder, and shopper has bids.
func (s *Store) GetWonListingsForShopper(shopperID string) ([]model.Listing, error) {
	rows, err := s.DB.Conn.Query(
		`SELECT l.listing_id, l.domain_name, l.listing_status, l.listing_type, l.auction_type_id,
			l.start_time, l.end_time, l.asking_price_usd, l.current_price_usd, l.sale_price_usd,
			l.next_bid_price_usd, l.bidders_count, l.bids_count, l.is_auto_extended,
			l.seller_shopper_id, l.highest_bidder_shopper, l.auto_ext_window_sec, l.auto_ext_seconds,
			l.auto_ext_enabled, l.radar_visible, l.created_at
		 FROM listings l
		 WHERE l.listing_status = 'SOLD'
		   AND l.highest_bidder_shopper = ?
		   AND EXISTS (SELECT 1 FROM bids b WHERE b.listing_id = l.listing_id AND b.shopper_id = ? AND b.bid_status = 'ACTIVE')
		 ORDER BY l.listing_id`,
		shopperID, shopperID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []model.Listing
	for rows.Next() {
		l, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		listings = append(listings, *l)
	}
	return listings, rows.Err()
}

// GetLostListingsForShopper returns listings where status IN (SOLD,CLOSED), shopper has bids, but shopper is NOT the highest bidder.
func (s *Store) GetLostListingsForShopper(shopperID string) ([]model.Listing, error) {
	rows, err := s.DB.Conn.Query(
		`SELECT l.listing_id, l.domain_name, l.listing_status, l.listing_type, l.auction_type_id,
			l.start_time, l.end_time, l.asking_price_usd, l.current_price_usd, l.sale_price_usd,
			l.next_bid_price_usd, l.bidders_count, l.bids_count, l.is_auto_extended,
			l.seller_shopper_id, l.highest_bidder_shopper, l.auto_ext_window_sec, l.auto_ext_seconds,
			l.auto_ext_enabled, l.radar_visible, l.created_at
		 FROM listings l
		 WHERE l.listing_status IN ('SOLD', 'CLOSED')
		   AND (l.highest_bidder_shopper IS NULL OR l.highest_bidder_shopper != ?)
		   AND EXISTS (SELECT 1 FROM bids b WHERE b.listing_id = l.listing_id AND b.shopper_id = ? AND b.bid_status = 'ACTIVE')
		 ORDER BY l.listing_id`,
		shopperID, shopperID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []model.Listing
	for rows.Next() {
		l, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		listings = append(listings, *l)
	}
	return listings, rows.Err()
}

// SearchListingsByDomain returns OPEN listings where domain_name matches query (case-insensitive LIKE).
func (s *Store) SearchListingsByDomain(query string) ([]model.Listing, error) {
	rows, err := s.DB.Conn.Query(
		`SELECT `+listingColumns+` FROM listings WHERE listing_status = 'OPEN' AND LOWER(domain_name) LIKE LOWER(?) ORDER BY listing_id`,
		"%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []model.Listing
	for rows.Next() {
		l, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		listings = append(listings, *l)
	}
	return listings, rows.Err()
}

// GetRadarListings returns OPEN listings where radar_visible = 1.
func (s *Store) GetRadarListings() ([]model.Listing, error) {
	rows, err := s.DB.Conn.Query(
		`SELECT `+listingColumns+` FROM listings WHERE listing_status = 'OPEN' AND radar_visible = 1 ORDER BY listing_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []model.Listing
	for rows.Next() {
		l, err := scanListing(rows)
		if err != nil {
			return nil, err
		}
		listings = append(listings, *l)
	}
	return listings, rows.Err()
}

func (s *Store) UpdateListingRadarVisible(listingID int64, visible bool) error {
	_, err := s.DB.Conn.Exec(
		`UPDATE listings SET radar_visible = ? WHERE listing_id = ?`,
		boolToInt(visible), listingID,
	)
	return err
}

// WipeAll drops all tables and recreates them without seeding.
func (s *Store) WipeAll() {
	s.DB.DropAll()
}

// --- Utility Operations ---

func (s *Store) SeedDefaults() {
	s.DB.SeedDefaults()
}

func (s *Store) Reset() {
	s.DB.DropAll()
	s.DB.SeedDefaults()
}
