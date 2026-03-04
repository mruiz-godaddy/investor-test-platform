import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IShopperRepository } from '../repositories/IShopperRepository';
import type { Shopper } from '../entities/Shopper';

@injectable()
export class GetShoppersUseCase {
  constructor(@inject(TOKENS.IShopperRepository) private repo: IShopperRepository) {}

  execute(): Promise<Shopper[]> {
    return this.repo.getAll();
  }
}
