export interface BidResult {
  listingId: number;
  bidId: string;
  bidAmountUsd: number;
  isHighestBidder: boolean;
  status: string;
}
