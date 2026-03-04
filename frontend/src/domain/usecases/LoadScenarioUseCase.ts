import { injectable, inject } from 'tsyringe';
import { TOKENS } from '../../di/tokens';
import type { IScenarioRepository } from '../repositories/IScenarioRepository';
import type { ScenarioName, ScenarioResult } from '../entities/Scenario';

@injectable()
export class LoadScenarioUseCase {
  constructor(@inject(TOKENS.IScenarioRepository) private repo: IScenarioRepository) {}

  execute(name: ScenarioName): Promise<ScenarioResult> {
    return this.repo.load(name);
  }
}
