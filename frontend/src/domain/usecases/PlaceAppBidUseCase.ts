import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IAppRepository } from '../repositories/IAppRepository';
import type { BidResult } from '../entities/BidResult';

@injectable()
export class PlaceAppBidUseCase {
  constructor(@inject(TOKENS.IAppRepository) private repo: IAppRepository) {}

  execute(id: number, shopperId: string, body: { usdBidAmount: number; isTosAccepted: boolean }): Promise<BidResult> {
    return this.repo.placeBid(id, shopperId, body);
  }
}
