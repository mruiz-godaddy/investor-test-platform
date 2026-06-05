import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IAdminRepository, SetupResult } from '../repositories/IAdminRepository';

@injectable()
export class SetupSystemUseCase {
  constructor(@inject(TOKENS.IAdminRepository) private repo: IAdminRepository) {}

  execute(durationMinutes?: number, appShopperId?: string): Promise<SetupResult> {
    return this.repo.setupSystem(durationMinutes, appShopperId);
  }
}
