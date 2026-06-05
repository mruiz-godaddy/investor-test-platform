import type { AdminListing, CreateListingRequest, CreateListingResponse } from '../entities/Listing';
import type { ListingStatus } from '../entities/Listing';
import type { BidResult } from '../entities/BidResult';

export interface IListingRepository {
  getAll(): Promise<AdminListing[]>;
  getById(id: number): Promise<AdminListing>;
  create(req: CreateListingRequest): Promise<CreateListingResponse>;
  updateStatus(id: number, status: ListingStatus): Promise<AdminListing>;
  updateEndTime(id: number, update: { endTime?: string; addSeconds?: number }): Promise<AdminListing>;
  updateRadarVisible(id: number, radarVisible: boolean): Promise<AdminListing>;
  placeSniperBid(id: number, shopperId: string, bidAmountUsd: number): Promise<BidResult>;
}
