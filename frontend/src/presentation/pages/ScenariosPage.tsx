import { useScenariosViewModel } from '../hooks/useScenariosViewModel';
import ScenarioGrid from '../components/scenarios/ScenarioGrid';
import ScenarioResultPanel from '../components/scenarios/ScenarioResultPanel';

export default function ScenariosPage() {
  const { result, loadingName, loadScenario } = useScenariosViewModel();

  return (
    <div>
      <h2 className="text-xl font-bold text-gray-900 dark:text-white">Scenarios</h2>
      <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">Load a predefined scenario. This resets the database first.</p>
      <div className="mt-4">
        <ScenarioGrid onLoad={loadScenario} loadingName={loadingName} />
      </div>
      {result && (
        <div className="mt-6">
          <ScenarioResultPanel result={result} />
        </div>
      )}
    </div>
  );
}
