import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IAppRepository } from '../repositories/IAppRepository';

@injectable()
export class GetAppListingUseCase {
  constructor(@inject(TOKENS.IAppRepository) private repo: IAppRepository) {}

  execute(id: number, shopperId: string): Promise<unknown> {
    return this.repo.getListing(id, shopperId);
  }
}
