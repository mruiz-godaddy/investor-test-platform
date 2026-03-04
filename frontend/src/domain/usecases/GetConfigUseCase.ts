import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IConfigRepository } from '../repositories/IConfigRepository';
import type { ConfigSnapshot } from '../entities/ServerConfig';

@injectable()
export class GetConfigUseCase {
  constructor(@inject(TOKENS.IConfigRepository) private repo: IConfigRepository) {}

  execute(): Promise<ConfigSnapshot> {
    return this.repo.get();
  }
}
