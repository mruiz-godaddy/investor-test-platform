import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IAdminRepository } from '../repositories/IAdminRepository';
import type { CartEvent } from '../entities/CartEvent';

@injectable()
export class GetCartEventsUseCase {
  constructor(@inject(TOKENS.IAdminRepository) private repo: IAdminRepository) {}

  execute(): Promise<CartEvent[]> {
    return this.repo.getCartEvents();
  }
}
