import { SCENARIOS } from '../../../domain/entities/Scenario';
import type { ScenarioName } from '../../../domain/entities/Scenario';
import ScenarioCard from './ScenarioCard';

interface Props {
  onLoad: (name: ScenarioName) => void;
  loadingName: ScenarioName | null;
}

export default function ScenarioGrid({ onLoad, loadingName }: Props) {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {SCENARIOS.map((scenario) => (
        <ScenarioCard
          key={scenario.name}
          scenario={scenario}
          onLoad={() => onLoad(scenario.name)}
          isLoading={loadingName === scenario.name}
        />
      ))}
    </div>
  );
}
