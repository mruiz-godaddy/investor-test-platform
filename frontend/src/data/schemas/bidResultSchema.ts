import { z } from 'zod';

export const bidResultSchema = z.object({
  listingId: z.number(),
  bidId: z.string(),
  bidAmountUsd: z.number(),
  isHighestBidder: z.boolean(),
  status: z.string(),
});
