import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IListingRepository } from '../repositories/IListingRepository';
import type { AdminListing, ListingStatus } from '../entities/Listing';

@injectable()
export class UpdateListingStatusUseCase {
  constructor(@inject(TOKENS.IListingRepository) private repo: IListingRepository) {}

  execute(id: number, status: ListingStatus): Promise<AdminListing> {
    return this.repo.updateStatus(id, status);
  }
}
