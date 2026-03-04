import { z } from 'zod';

export const adminBidSchema = z.object({
  bidId: z.string(),
  shopperId: z.string(),
  bidAmountUsd: z.number(),
  bidType: z.enum(['AUCTION', 'PROXY']),
  bidStatus: z.enum(['ACTIVE', 'CANCELLED']),
  isHighBid: z.boolean(),
  parentBidId: z.string(),
  createdAt: z.string(),
});

export const appBidHistoryEntrySchema = z.object({
  bidAmount: z.array(z.object({ cost: z.number(), currency: z.string() })),
  bidDate: z.string(),
  bidExpirationDate: z.string(),
  bidder: z.number(),
  comment: z.string(),
});
