import { z } from 'zod';
import { adminBidSchema, appBidHistoryEntrySchema } from './bidSchema';

export const adminListingSchema = z.object({
  listingId: z.number(),
  domainName: z.string(),
  listingStatus: z.enum(['OPEN', 'SOLD', 'CLOSED']),
  listingType: z.string(),
  auctionTypeId: z.number(),
  startTime: z.string(),
  endTime: z.string(),
  askingPriceUsd: z.number(),
  currentPriceUsd: z.number(),
  salePriceUsd: z.number().nullable(),
  nextBidPriceUsd: z.number(),
  biddersCount: z.number(),
  bidsCount: z.number(),
  isAutoExtended: z.boolean(),
  sellerShopperId: z.string(),
  highestBidderShopper: z.string(),
  autoExtEnabled: z.boolean(),
  autoExtWindowSec: z.number(),
  autoExtSeconds: z.number(),
  radarVisible: z.boolean(),
  createdAt: z.string(),
  bidHistory: z.array(adminBidSchema),
});

const priceArraySchema = z.array(z.object({ cost: z.number(), currency: z.string() }));

export const appListingSchema = z.object({
  listingId: z.number(),
  domainName: z.string(),
  listingEndTime: z.string(),
  listingType: z.string(),
  auctionTypeId: z.number(),
  auctionEndTime: z.string(),
  price: priceArraySchema,
  minimumBidPrice: priceArraySchema,
  minimumOfferPrice: priceArraySchema,
  nextBidPrice: z.array(z.object({ cost: z.number(), currency: z.string() })),
  bidPrice: priceArraySchema,
  numberOfBids: z.number(),
  numberOfBidders: z.number(),
  domainId: z.number(),
  age: z.number(),
  traffic: z.number(),
  visits: z.number(),
  watchers: z.number(),
  isPinned: z.boolean(),
  isWatching: z.boolean(),
  renewalPfid: z.number(),
  estimatedValueRank: z.string(),
  memberBiddingStatus: z.string(),
  proxyBidPrice: priceArraySchema,
  bidHistory: z.array(appBidHistoryEntrySchema),
}).passthrough();

export const createListingRequestSchema = z.object({
  domainName: z.string(),
  sellerShopperId: z.string(),
  askingPriceUsd: z.number().optional(),
  endTime: z.string().optional(),
  startTime: z.string().optional(),
  auctionTypeId: z.number().optional(),
  listingType: z.string().optional(),
  autoExtEnabled: z.boolean().optional(),
  autoExtWindowSec: z.number().optional(),
  autoExtSeconds: z.number().optional(),
  radarVisible: z.boolean().optional(),
});

export const createListingResponseSchema = z.object({
  listingId: z.number(),
  domainName: z.string(),
  endTime: z.string(),
  listingStatus: z.string(),
});
