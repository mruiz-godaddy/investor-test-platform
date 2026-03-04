import type { BidType, BidStatus } from './Bid';

export interface ShopperBid {
  bidId: string;
  listingId: number;
  shopperId: string;
  bidAmountUsd: number;
  bidType: BidType;
  bidStatus: BidStatus;
  isHighBid: boolean;
  parentBidId: string;
  createdAt: string;
  domainName: string;
  listingStatus: string;
  highestBidderShopper: string;
}
