import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IShopperRepository } from '../repositories/IShopperRepository';
import type { ShopperDetail } from '../entities/ShopperDetail';

@injectable()
export class GetShopperUseCase {
  constructor(@inject(TOKENS.IShopperRepository) private repo: IShopperRepository) {}

  execute(id: string): Promise<ShopperDetail> {
    return this.repo.getById(id);
  }
}
