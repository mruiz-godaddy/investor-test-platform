import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IListingRepository } from '../repositories/IListingRepository';
import type { AdminListing } from '../entities/Listing';

@injectable()
export class UpdateEndTimeUseCase {
  constructor(@inject(TOKENS.IListingRepository) private repo: IListingRepository) {}

  execute(id: number, update: { endTime?: string; addSeconds?: number }): Promise<AdminListing> {
    return this.repo.updateEndTime(id, update);
  }
}
