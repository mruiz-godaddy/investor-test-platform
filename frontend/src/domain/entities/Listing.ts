import type { AdminBid } from './Bid';

export const ListingStatus = {
  OPEN: 'OPEN',
  SOLD: 'SOLD',
  CLOSED: 'CLOSED',
} as const;
export type ListingStatus = (typeof ListingStatus)[keyof typeof ListingStatus];

export interface AdminListing {
  listingId: number;
  domainName: string;
  listingStatus: ListingStatus;
  listingType: string;
  auctionTypeId: number;
  startTime: string;
  endTime: string;
  askingPriceUsd: number;
  currentPriceUsd: number;
  salePriceUsd: number | null;
  reservePriceUsd: number;
  nextBidPriceUsd: number;
  biddersCount: number;
  bidsCount: number;
  isReserveMet: boolean;
  isAutoExtended: boolean;
  sellerShopperId: string;
  highestBidderShopper: string;
  autoExtEnabled: boolean;
  autoExtWindowSec: number;
  autoExtSeconds: number;
  createdAt: string;
  bidHistory: AdminBid[];
}

export interface CreateListingRequest {
  domainName: string;
  sellerShopperId: string;
  askingPriceUsd?: number;
  reservePriceUsd?: number;
  endTime?: string;
  startTime?: string;
  auctionTypeId?: number;
  listingType?: string;
  autoExtEnabled?: boolean;
  autoExtWindowSec?: number;
  autoExtSeconds?: number;
}

export interface CreateListingResponse {
  listingId: number;
  domainName: string;
  endTime: string;
  listingStatus: string;
}
