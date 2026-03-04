import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { ITimeRepository } from '../repositories/ITimeRepository';
import type { TimeResponse, TimeUpdate } from '../entities/ServerTime';

@injectable()
export class UpdateTimeUseCase {
  constructor(@inject(TOKENS.ITimeRepository) private repo: ITimeRepository) {}

  execute(update: TimeUpdate): Promise<TimeResponse> {
    return this.repo.update(update);
  }
}
