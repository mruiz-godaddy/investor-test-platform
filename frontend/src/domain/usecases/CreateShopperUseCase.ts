import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IShopperRepository } from '../repositories/IShopperRepository';
import type { Shopper } from '../entities/Shopper';

@injectable()
export class CreateShopperUseCase {
  constructor(@inject(TOKENS.IShopperRepository) private repo: IShopperRepository) {}

  execute(req: { shopperId: string; memberId: number; customerId?: string; displayName?: string }): Promise<Shopper> {
    return this.repo.create(req);
  }
}
