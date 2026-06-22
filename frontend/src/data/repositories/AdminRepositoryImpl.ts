import { injectable } from 'tsyringe';
import type { IAdminRepository, DatabaseExport, ImportResult, SetupResult, GenerateBinResult, GenerateBinOptions } from '../../domain/repositories/IAdminRepository';
import type { CartEvent } from '../../domain/entities/CartEvent';
import { AdminApiDataSource } from '../datasources/AdminApiDataSource';

@injectable()
export class AdminRepositoryImpl implements IAdminRepository {
  constructor(private ds: AdminApiDataSource) {}

  async setupSystem(durationMinutes?: number, appShopperId?: string): Promise<SetupResult> {
    return this.ds.setupSystem(durationMinutes, appShopperId);
  }

  async generateBin(opts: GenerateBinOptions): Promise<GenerateBinResult> {
    return this.ds.generateBin(opts);
  }

  async getCartEvents(): Promise<CartEvent[]> {
    return this.ds.getCartEvents();
  }

  async clearCartEvents(): Promise<{ status: string }> {
    return this.ds.clearCartEvents();
  }

  async resetDatabase(): Promise<{ status: string }> {
    return this.ds.reset();
  }

  async wipeDatabase(): Promise<{ status: string }> {
    return this.ds.wipeDatabase();
  }

  async exportDatabase(): Promise<DatabaseExport> {
    return this.ds.exportDatabase();
  }

  async importDatabase(data: { shoppers: unknown[]; listings: unknown[]; bids: unknown[] }): Promise<ImportResult> {
    return this.ds.importDatabase(data);
  }
}
