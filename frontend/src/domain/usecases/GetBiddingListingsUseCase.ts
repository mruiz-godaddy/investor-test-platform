import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IAppRepository } from '../repositories/IAppRepository';

@injectable()
export class GetBiddingListingsUseCase {
  constructor(@inject(TOKENS.IAppRepository) private repo: IAppRepository) {}

  execute(shopperId: string): Promise<unknown> {
    return this.repo.getBiddingListings(shopperId);
  }
}
