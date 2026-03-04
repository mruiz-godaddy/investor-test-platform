import { z } from 'zod';

export const shopperBidSchema = z.object({
  bidId: z.string(),
  listingId: z.number(),
  shopperId: z.string(),
  bidAmountUsd: z.number(),
  bidType: z.string(),
  bidStatus: z.string(),
  isHighBid: z.boolean(),
  parentBidId: z.string(),
  createdAt: z.string(),
  domainName: z.string().optional().default(''),
  listingStatus: z.string().optional().default(''),
  highestBidderShopper: z.string().optional().default(''),
});

export const shopperDetailSchema = z.object({
  shopperId: z.string(),
  memberId: z.number(),
  customerId: z.string(),
  displayName: z.string(),
  bidHistory: z.array(shopperBidSchema),
});
