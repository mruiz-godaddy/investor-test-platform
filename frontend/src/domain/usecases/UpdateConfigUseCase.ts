import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IConfigRepository } from '../repositories/IConfigRepository';
import type { ConfigSnapshot, ConfigUpdate } from '../entities/ServerConfig';

@injectable()
export class UpdateConfigUseCase {
  constructor(@inject(TOKENS.IConfigRepository) private repo: IConfigRepository) {}

  execute(config: ConfigUpdate): Promise<ConfigSnapshot> {
    return this.repo.update(config);
  }
}
