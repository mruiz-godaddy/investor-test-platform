import { injectable } from 'tsyringe';
import type { IAppRepository } from '../../domain/repositories/IAppRepository';
import type { BidResult } from '../../domain/entities/BidResult';
import { AppApiDataSource } from '../datasources/AppApiDataSource';

@injectable()
export class AppRepositoryImpl implements IAppRepository {
  constructor(private ds: AppApiDataSource) {}

  async getListing(id: number, shopperId: string): Promise<unknown> {
    return this.ds.getListing(id, shopperId);
  }

  async placeBid(id: number, shopperId: string, body: { usdBidAmount: number; isTosAccepted: boolean }): Promise<BidResult> {
    return this.ds.placeBid(id, shopperId, body);
  }

  async getBiddingListings(shopperId: string): Promise<unknown> {
    return this.ds.getBiddingListings(shopperId);
  }
}
