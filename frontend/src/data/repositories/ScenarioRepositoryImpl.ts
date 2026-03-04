import { injectable } from 'tsyringe';
import type { IScenarioRepository } from '../../domain/repositories/IScenarioRepository';
import type { ScenarioName, ScenarioResult } from '../../domain/entities/Scenario';
import { AdminApiDataSource } from '../datasources/AdminApiDataSource';
import { mapScenarioResult } from '../mappers/ScenarioMapper';

@injectable()
export class ScenarioRepositoryImpl implements IScenarioRepository {
  constructor(private ds: AdminApiDataSource) {}

  async load(name: ScenarioName): Promise<ScenarioResult> {
    const dto = await this.ds.loadScenario(name);
    return mapScenarioResult(dto);
  }
}
