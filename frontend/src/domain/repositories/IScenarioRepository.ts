import type { ScenarioName, ScenarioResult } from '../entities/Scenario';

export interface IScenarioRepository {
  load(name: ScenarioName): Promise<ScenarioResult>;
}
