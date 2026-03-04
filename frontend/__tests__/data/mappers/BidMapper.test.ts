import { describe, it, expect } from 'vitest';
import { mapAdminBid } from '../../../src/data/mappers/BidMapper';

describe('mapAdminBid', () => {
  it('maps AUCTION bid correctly', () => {
    const result = mapAdminBid({
      bidId: 'bid-123',
      shopperId: 'shopper-buyer',
      bidAmountUsd: 5_000_000,
      bidType: 'AUCTION',
      bidStatus: 'ACTIVE',
      isHighBid: true,
      parentBidId: '',
      createdAt: '2024-01-01T00:00:00Z',
    });

    expect(result.bidType).toBe('AUCTION');
    expect(result.bidStatus).toBe('ACTIVE');
    expect(result.isHighBid).toBe(true);
    expect(result.parentBidId).toBe('');
  });

  it('maps PROXY bid with CANCELLED status', () => {
    const result = mapAdminBid({
      bidId: 'bid-456',
      shopperId: 'shopper-buyer-a',
      bidAmountUsd: 20_000_000,
      bidType: 'PROXY',
      bidStatus: 'CANCELLED',
      isHighBid: false,
      parentBidId: 'bid-123',
      createdAt: '2024-01-01T00:00:00Z',
    });

    expect(result.bidType).toBe('PROXY');
    expect(result.bidStatus).toBe('CANCELLED');
    expect(result.isHighBid).toBe(false);
    expect(result.parentBidId).toBe('bid-123');
  });
});
