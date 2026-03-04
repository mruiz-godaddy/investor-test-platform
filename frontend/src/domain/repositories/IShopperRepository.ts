import type { Shopper } from '../entities/Shopper';
import type { ShopperDetail } from '../entities/ShopperDetail';

export interface IShopperRepository {
  getAll(): Promise<Shopper[]>;
  getById(id: string): Promise<ShopperDetail>;
  create(req: { shopperId: string; memberId: number; customerId?: string; displayName?: string }): Promise<Shopper>;
}
