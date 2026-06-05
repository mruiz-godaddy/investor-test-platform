package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

type DB struct {
	Conn *sql.DB
}

func New(dbPath string) *DB {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open sqlite: %v", err)
	}

	// For :memory: databases, each connection gets its own independent DB.
	// Pin to a single connection so all queries share the same in-memory DB.
	conn.SetMaxOpenConns(1)

	_, err = conn.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		log.Fatalf("failed to set WAL mode: %v", err)
	}

	_, err = conn.Exec("PRAGMA foreign_keys=ON")
	if err != nil {
		log.Fatalf("failed to enable foreign keys: %v", err)
	}

	d := &DB{Conn: conn}
	d.createTables()
	return d
}

func (d *DB) createTables() {
	_, err := d.Conn.Exec(`
CREATE TABLE IF NOT EXISTS shoppers (
    shopper_id    TEXT PRIMARY KEY,
    member_id     INTEGER NOT NULL UNIQUE,
    customer_id   TEXT NOT NULL UNIQUE,
    display_name  TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS listings (
    listing_id              INTEGER PRIMARY KEY AUTOINCREMENT,
    domain_name             TEXT    NOT NULL,
    listing_status          TEXT    NOT NULL DEFAULT 'OPEN',
    listing_type            TEXT    NOT NULL DEFAULT 'EXPIRY_AUCTIONS',
    auction_type_id         INTEGER NOT NULL DEFAULT 16,
    start_time              TEXT    NOT NULL,
    end_time                TEXT    NOT NULL,
    asking_price_usd        INTEGER NOT NULL DEFAULT 5000000,
    current_price_usd       INTEGER NOT NULL DEFAULT 0,
    sale_price_usd          INTEGER,
    next_bid_price_usd      INTEGER NOT NULL DEFAULT 0,
    bidders_count           INTEGER NOT NULL DEFAULT 0,
    bids_count              INTEGER NOT NULL DEFAULT 0,
    is_auto_extended        INTEGER NOT NULL DEFAULT 0,
    seller_shopper_id       TEXT    NOT NULL,
    highest_bidder_shopper  TEXT,
    auto_ext_window_sec     INTEGER NOT NULL DEFAULT 60,
    auto_ext_seconds        INTEGER NOT NULL DEFAULT 300,
    auto_ext_enabled        INTEGER NOT NULL DEFAULT 1,
    radar_visible           INTEGER NOT NULL DEFAULT 0,
    created_at              TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS bids (
    bid_id          TEXT    PRIMARY KEY,
    listing_id      INTEGER NOT NULL,
    shopper_id      TEXT    NOT NULL,
    bid_amount_usd  INTEGER NOT NULL,
    bid_type        TEXT    NOT NULL DEFAULT 'AUCTION',
    bid_status      TEXT    NOT NULL DEFAULT 'ACTIVE',
    is_high_bid     INTEGER NOT NULL DEFAULT 0,
    parent_bid_id   TEXT    NOT NULL DEFAULT '',
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY (listing_id) REFERENCES listings(listing_id),
    FOREIGN KEY (shopper_id) REFERENCES shoppers(shopper_id)
);
`)
	if err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	// Migrate existing DBs: add radar_visible if missing
	d.Conn.Exec(`ALTER TABLE listings ADD COLUMN radar_visible INTEGER NOT NULL DEFAULT 0`)
}

func (d *DB) Close() {
	d.Conn.Close()
}

func (d *DB) SeedDefaults() {
	_, err := d.Conn.Exec(`
INSERT OR IGNORE INTO shoppers (shopper_id, member_id, customer_id, display_name) VALUES
    ('shopper-seller', 10001, 'cust-seller', ''),
    ('shopper-buyer', 10002, 'cust-buyer', '');
`)
	if err != nil {
		log.Fatalf("failed to seed defaults: %v", err)
	}
}

func (d *DB) DropAll() {
	// Use DELETE instead of DROP+CREATE to avoid race conditions
	// with the lifecycle goroutine querying tables mid-reset.
	d.Conn.Exec("PRAGMA foreign_keys=OFF")
	d.Conn.Exec("DELETE FROM bids")
	d.Conn.Exec("DELETE FROM listings")
	d.Conn.Exec("DELETE FROM shoppers")
	d.Conn.Exec("DELETE FROM sqlite_sequence")
	d.Conn.Exec("PRAGMA foreign_keys=ON")
}
