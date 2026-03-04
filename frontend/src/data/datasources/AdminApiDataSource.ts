import { injectable } from 'tsyringe';
import api from '../../lib/axios';
import { adminListingSchema, createListingResponseSchema } from '../schemas/listingSchema';
import { bidResultSchema } from '../schemas/bidResultSchema';
import { shopperSchema } from '../schemas/shopperSchema';
import { shopperDetailSchema } from '../schemas/shopperDetailSchema';
import { configSnapshotSchema } from '../schemas/configSchema';
import { timeResponseSchema } from '../schemas/timeSchema';
import { scenarioResultSchema } from '../schemas/scenarioSchema';
import { z } from 'zod';

@injectable()
export class AdminApiDataSource {
  async createListing(body: Record<string, unknown>) {
    const { data } = await api.post('/admin/listings', body);
    return createListingResponseSchema.parse(data);
  }

  async getListings() {
    const { data } = await api.get('/admin/listings');
    return z.array(adminListingSchema).parse(data);
  }

  async getListing(id: number) {
    const { data } = await api.get(`/admin/listings/${id}`);
    return adminListingSchema.parse(data);
  }

  async updateListingStatus(id: number, listingStatus: string) {
    const { data } = await api.put(`/admin/listings/${id}/status`, { listingStatus });
    return adminListingSchema.parse(data);
  }

  async updateEndTime(id: number, body: { endTime?: string; addSeconds?: number }) {
    const { data } = await api.put(`/admin/listings/${id}/endtime`, body);
    return adminListingSchema.parse(data);
  }

  async placeSniperBid(id: number, body: { shopperId: string; bidAmountUsd: number }) {
    const { data } = await api.post(`/admin/listings/${id}/sniper-bid`, body);
    return bidResultSchema.parse(data);
  }

  async createShopper(body: Record<string, unknown>) {
    const { data } = await api.post('/admin/shoppers', body);
    return shopperSchema.parse(data);
  }

  async getShoppers() {
    const { data } = await api.get('/admin/shoppers');
    return z.array(shopperSchema).parse(data);
  }

  async getShopper(id: string) {
    const { data } = await api.get(`/admin/shoppers/${id}`);
    return shopperDetailSchema.parse(data);
  }

  async reset() {
    const { data } = await api.post('/admin/reset');
    return data as { status: string };
  }

  async loadScenario(name: string) {
    const { data } = await api.post(`/admin/scenarios/${name}`);
    return scenarioResultSchema.parse(data);
  }

  async getConfig() {
    const { data } = await api.get('/admin/config');
    return configSnapshotSchema.parse(data);
  }

  async updateConfig(body: Record<string, unknown>) {
    const { data } = await api.put('/admin/config', body);
    return configSnapshotSchema.parse(data);
  }

  async getTime() {
    const { data } = await api.get('/admin/time');
    return timeResponseSchema.parse(data);
  }

  async updateTime(body: Record<string, unknown>) {
    const { data } = await api.put('/admin/time', body);
    return timeResponseSchema.parse(data);
  }

  async setupSystem() {
    const { data } = await api.post('/admin/setup');
    return data as { status: string; shoppers: number; listings: number };
  }

  async wipeDatabase() {
    const { data } = await api.post('/admin/wipe');
    return data as { status: string };
  }

  async exportDatabase() {
    const { data } = await api.get('/admin/export');
    return data as { version: number; exportedAt: string; shoppers: unknown[]; listings: unknown[]; bids: unknown[] };
  }

  async importDatabase(body: { shoppers: unknown[]; listings: unknown[]; bids: unknown[] }) {
    const { data } = await api.post('/admin/import', body);
    return data as { status: string; shoppers: number; listings: number; bids: number };
  }
}
