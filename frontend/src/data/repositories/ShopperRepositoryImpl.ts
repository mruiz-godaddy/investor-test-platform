import { injectable } from 'tsyringe';
import type { IShopperRepository } from '../../domain/repositories/IShopperRepository';
import type { Shopper } from '../../domain/entities/Shopper';
import { AdminApiDataSource } from '../datasources/AdminApiDataSource';
import { mapShopper } from '../mappers/ShopperMapper';

@injectable()
export class ShopperRepositoryImpl implements IShopperRepository {
  constructor(private ds: AdminApiDataSource) {}

  async getAll(): Promise<Shopper[]> {
    const dtos = await this.ds.getShoppers();
    return dtos.map(mapShopper);
  }

  async create(req: { shopperId: string; memberId: number; customerId?: string; displayName?: string }): Promise<Shopper> {
    const dto = await this.ds.createShopper(req as unknown as Record<string, unknown>);
    return mapShopper(dto);
  }
}
