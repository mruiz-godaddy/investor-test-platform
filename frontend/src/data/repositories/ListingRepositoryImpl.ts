import { injectable, inject } from 'tsyringe';
import type { IListingRepository } from '../../domain/repositories/IListingRepository';
import type { AdminListing, CreateListingRequest, CreateListingResponse, ListingStatus } from '../../domain/entities/Listing';
import type { BidResult } from '../../domain/entities/BidResult';
import { AdminApiDataSource } from '../datasources/AdminApiDataSource';
import { mapAdminListing } from '../mappers/ListingMapper';

@injectable()
export class ListingRepositoryImpl implements IListingRepository {
  constructor(private ds: AdminApiDataSource) {}

  async getAll(): Promise<AdminListing[]> {
    const dtos = await this.ds.getListings();
    return dtos.map(mapAdminListing);
  }

  async getById(id: number): Promise<AdminListing> {
    const dto = await this.ds.getListing(id);
    return mapAdminListing(dto);
  }

  async create(req: CreateListingRequest): Promise<CreateListingResponse> {
    return this.ds.createListing(req as unknown as Record<string, unknown>);
  }

  async updateStatus(id: number, status: ListingStatus): Promise<AdminListing> {
    const dto = await this.ds.updateListingStatus(id, status);
    return mapAdminListing(dto);
  }

  async updateEndTime(id: number, update: { endTime?: string; addSeconds?: number }): Promise<AdminListing> {
    const dto = await this.ds.updateEndTime(id, update);
    return mapAdminListing(dto);
  }

  async updateRadarVisible(id: number, radarVisible: boolean): Promise<AdminListing> {
    const dto = await this.ds.updateRadarVisible(id, radarVisible);
    return mapAdminListing(dto);
  }

  async placeSniperBid(id: number, shopperId: string, bidAmountUsd: number): Promise<BidResult> {
    return this.ds.placeSniperBid(id, { shopperId, bidAmountUsd });
  }
}
