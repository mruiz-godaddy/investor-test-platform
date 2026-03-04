import type { ScenarioResult } from '../../domain/entities/Scenario';
import type { z } from 'zod';
import type { scenarioResultSchema } from '../schemas/scenarioSchema';

type ScenarioDto = z.infer<typeof scenarioResultSchema>;

export function mapScenarioResult(dto: ScenarioDto): ScenarioResult {
  return {
    scenario: dto.scenario,
    description: dto.description,
    config: dto.config,
    shoppers: dto.shoppers,
    listings: dto.listings,
    bidsPlaced: dto.bidsPlaced,
  };
}
