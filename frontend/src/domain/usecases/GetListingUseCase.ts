import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IListingRepository } from '../repositories/IListingRepository';
import type { AdminListing } from '../entities/Listing';

@injectable()
export class GetListingUseCase {
  constructor(@inject(TOKENS.IListingRepository) private repo: IListingRepository) {}

  execute(id: number): Promise<AdminListing> {
    return this.repo.getById(id);
  }
}
