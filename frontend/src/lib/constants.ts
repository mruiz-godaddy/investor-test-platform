export const LISTINGS_POLL_INTERVAL_MS = 3000;

export const DEFAULT_ASKING_PRICE_MICROS = 5_000_000;
export const DEFAULT_RESERVE_PRICE_MICROS = 0;
export const DEFAULT_AUCTION_TYPE_ID = 16;
export const DEFAULT_LISTING_TYPE = 'EXPIRY_AUCTIONS';
export const DEFAULT_AUTO_EXT_ENABLED = true;
export const DEFAULT_AUTO_EXT_WINDOW_SEC = 60;
export const DEFAULT_AUTO_EXT_SECONDS = 300;
export const DEFAULT_LISTING_DURATION_MINUTES = 5;

export const SEEDED_SHOPPERS = [
  { shopperId: 'shopper-seller', memberId: 10001 },
  { shopperId: 'shopper-buyer', memberId: 10002 },
] as const;
