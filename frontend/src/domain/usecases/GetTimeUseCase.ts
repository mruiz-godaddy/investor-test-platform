import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { ITimeRepository } from '../repositories/ITimeRepository';
import type { TimeResponse } from '../entities/ServerTime';

@injectable()
export class GetTimeUseCase {
  constructor(@inject(TOKENS.ITimeRepository) private repo: ITimeRepository) {}

  execute(): Promise<TimeResponse> {
    return this.repo.get();
  }
}
