import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IAdminRepository } from '../repositories/IAdminRepository';
import type { DatabaseExport } from '../repositories/IAdminRepository';

@injectable()
export class ExportDatabaseUseCase {
  constructor(@inject(TOKENS.IAdminRepository) private repo: IAdminRepository) {}

  execute(): Promise<DatabaseExport> {
    return this.repo.exportDatabase();
  }
}
