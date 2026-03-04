import type { Shopper } from '../entities/Shopper';

export interface IShopperRepository {
  getAll(): Promise<Shopper[]>;
  create(req: { shopperId: string; memberId: number; customerId?: string; displayName?: string }): Promise<Shopper>;
}
