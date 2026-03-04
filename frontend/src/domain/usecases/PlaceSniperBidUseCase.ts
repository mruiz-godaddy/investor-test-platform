import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IListingRepository } from '../repositories/IListingRepository';
import type { BidResult } from '../entities/BidResult';

@injectable()
export class PlaceSniperBidUseCase {
  constructor(@inject(TOKENS.IListingRepository) private repo: IListingRepository) {}

  execute(id: number, shopperId: string, bidAmountUsd: number): Promise<BidResult> {
    return this.repo.placeSniperBid(id, shopperId, bidAmountUsd);
  }
}
