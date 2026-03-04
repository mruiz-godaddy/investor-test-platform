import { injectable } from 'tsyringe';
import type { IAdminRepository, DatabaseExport, ImportResult, SetupResult } from '../../domain/repositories/IAdminRepository';
import { AdminApiDataSource } from '../datasources/AdminApiDataSource';

@injectable()
export class AdminRepositoryImpl implements IAdminRepository {
  constructor(private ds: AdminApiDataSource) {}

  async setupSystem(): Promise<SetupResult> {
    return this.ds.setupSystem();
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
