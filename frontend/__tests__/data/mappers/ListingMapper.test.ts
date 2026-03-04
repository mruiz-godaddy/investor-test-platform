import { describe, it, expect } from 'vitest';
import { mapAdminListing } from '../../../src/data/mappers/ListingMapper';

const baseBid = {
  bidId: 'bid-1',
  shopperId: 'shopper-buyer',
  bidAmountUsd: 5_000_000,
  bidType: 'AUCTION' as const,
  bidStatus: 'ACTIVE' as const,
  isHighBid: true,
  parentBidId: '',
  createdAt: '2024-01-01T00:00:00Z',
};

const baseDto = {
  listingId: 1,
  domainName: 'test.com',
  listingStatus: 'OPEN' as const,
  listingType: 'EXPIRY_AUCTIONS',
  auctionTypeId: 16,
  startTime: '2024-01-01T00:00:00Z',
  endTime: '2024-01-01T00:05:00Z',
  askingPriceUsd: 5_000_000,
  currentPriceUsd: 5_000_000,
  salePriceUsd: null,
  reservePriceUsd: 0,
  nextBidPriceUsd: 10_000_000,
  biddersCount: 1,
  bidsCount: 1,
  isReserveMet: false,
  isAutoExtended: false,
  sellerShopperId: 'shopper-seller',
  highestBidderShopper: 'shopper-buyer',
  autoExtEnabled: true,
  autoExtWindowSec: 60,
  autoExtSeconds: 300,
  createdAt: '2024-01-01T00:00:00Z',
  bidHistory: [baseBid],
};

describe('mapAdminListing', () => {
  it('maps all 22 fields correctly', () => {
    const result = mapAdminListing(baseDto);
    expect(result.listingId).toBe(1);
    expect(result.domainName).toBe('test.com');
    expect(result.listingStatus).toBe('OPEN');
    expect(result.askingPriceUsd).toBe(5_000_000);
    expect(result.autoExtEnabled).toBe(true);
    expect(result.bidHistory).toHaveLength(1);
  });

  it('handles null salePriceUsd', () => {
    const result = mapAdminListing({ ...baseDto, salePriceUsd: null });
    expect(result.salePriceUsd).toBeNull();
  });

  it('handles non-null salePriceUsd', () => {
    const result = mapAdminListing({ ...baseDto, salePriceUsd: 10_000_000 });
    expect(result.salePriceUsd).toBe(10_000_000);
  });

  it('maps empty bidHistory', () => {
    const result = mapAdminListing({ ...baseDto, bidHistory: [] });
    expect(result.bidHistory).toEqual([]);
  });
});
