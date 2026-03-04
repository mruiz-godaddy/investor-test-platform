import type { BidResult } from '../entities/BidResult';

export interface IAppRepository {
  getListing(id: number, shopperId: string): Promise<unknown>;
  placeBid(id: number, shopperId: string, body: { usdBidAmount: number; isTosAccepted: boolean }): Promise<BidResult>;
  getBiddingListings(shopperId: string): Promise<unknown>;
}
