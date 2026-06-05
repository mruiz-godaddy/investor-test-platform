export interface DatabaseExport {
  version: number;
  exportedAt: string;
  shoppers: unknown[];
  listings: unknown[];
  bids: unknown[];
}

export interface ImportResult {
  status: string;
  shoppers: number;
  listings: number;
  bids: number;
}

export interface SetupResult {
  status: string;
  shoppers: number;
  listings: number;
}

export interface IAdminRepository {
  setupSystem(durationMinutes?: number, appShopperId?: string): Promise<SetupResult>;
  resetDatabase(): Promise<{ status: string }>;
  wipeDatabase(): Promise<{ status: string }>;
  exportDatabase(): Promise<DatabaseExport>;
  importDatabase(data: { shoppers: unknown[]; listings: unknown[]; bids: unknown[] }): Promise<ImportResult>;
}
