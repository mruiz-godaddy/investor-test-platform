import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IAdminRepository, GenerateBinResult, GenerateBinOptions } from '../repositories/IAdminRepository';

@injectable()
export class GenerateBinUseCase {
  constructor(@inject(TOKENS.IAdminRepository) private repo: IAdminRepository) {}

  execute(opts: GenerateBinOptions = {}): Promise<GenerateBinResult> {
    return this.repo.generateBin(opts);
  }
}
