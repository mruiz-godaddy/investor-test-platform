import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IListingRepository } from '../repositories/IListingRepository';
import type { CreateListingRequest, CreateListingResponse } from '../entities/Listing';

@injectable()
export class CreateListingUseCase {
  constructor(@inject(TOKENS.IListingRepository) private repo: IListingRepository) {}

  execute(req: CreateListingRequest): Promise<CreateListingResponse> {
    return this.repo.create(req);
  }
}
