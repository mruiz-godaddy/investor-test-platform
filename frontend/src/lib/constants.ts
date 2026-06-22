export const LISTINGS_POLL_INTERVAL_MS = 3000;

export const DEFAULT_ASKING_PRICE_MICROS = 5_000_000;
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

// BIN / closeout / OCO inventory types. Each auctionTypeId drives a distinct
// X-Itc-Code in the app (AuctionRequestAddToCart). Used by the BIN Domains page
// to filter listings and by the BIN create control's inventory-type dropdown.
export const BIN_INVENTORY_TYPES = [
  { auctionTypeId: 20, listingType: 'CLOSEOUT_DOMAINS', label: 'Closeout (GoDaddy)', itc: 'expirycloseout' },
  { auctionTypeId: 39, listingType: 'CLOSEOUT_DOMAINS', label: 'Closeout (Partner)', itc: 'expirycloseout' },
  { auctionTypeId: 11, listingType: 'BUY_IT_NOW', label: 'Member BIN', itc: 'm2mbin' },
  { auctionTypeId: 10, listingType: 'MEMBER_LISTINGS', label: 'OCO + Buy Now', itc: 'm2mocobin' },
  { auctionTypeId: 9, listingType: 'MEMBER_LISTINGS', label: 'OCO (offer)', itc: 'm2moco' },
] as const;

export const BIN_AUCTION_TYPE_IDS: number[] = BIN_INVENTORY_TYPES.map((t) => t.auctionTypeId);
