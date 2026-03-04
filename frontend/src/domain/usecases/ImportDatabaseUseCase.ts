import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IAdminRepository, ImportResult } from '../repositories/IAdminRepository';

@injectable()
export class ImportDatabaseUseCase {
  constructor(@inject(TOKENS.IAdminRepository) private repo: IAdminRepository) {}

  execute(data: { shoppers: unknown[]; listings: unknown[]; bids: unknown[] }): Promise<ImportResult> {
    return this.repo.importDatabase(data);
  }
}
