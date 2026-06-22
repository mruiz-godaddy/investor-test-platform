import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IAdminRepository } from '../repositories/IAdminRepository';

@injectable()
export class ClearCartEventsUseCase {
  constructor(@inject(TOKENS.IAdminRepository) private repo: IAdminRepository) {}

  execute(): Promise<{ status: string }> {
    return this.repo.clearCartEvents();
  }
}
