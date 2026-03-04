import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IListingRepository } from '../repositories/IListingRepository';
import type { AdminListing } from '../entities/Listing';

@injectable()
export class GetListingsUseCase {
  constructor(@inject(TOKENS.IListingRepository) private repo: IListingRepository) {}

  execute(): Promise<AdminListing[]> {
    return this.repo.getAll();
  }
}
