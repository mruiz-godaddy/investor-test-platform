export const BidType = {
  AUCTION: 'AUCTION',
  PROXY: 'PROXY',
} as const;
export type BidType = (typeof BidType)[keyof typeof BidType];

export const BidStatus = {
  ACTIVE: 'ACTIVE',
  CANCELLED: 'CANCELLED',
} as const;
export type BidStatus = (typeof BidStatus)[keyof typeof BidStatus];

export interface AdminBid {
  bidId: string;
  shopperId: string;
  bidAmountUsd: number;
  bidType: BidType;
  bidStatus: BidStatus;
  isHighBid: boolean;
  parentBidId: string;
  createdAt: string;
}
