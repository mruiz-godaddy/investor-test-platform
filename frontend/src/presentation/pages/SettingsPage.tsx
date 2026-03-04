import { useSettingsViewModel } from '../hooks/useSettingsViewModel';
import ConfigForm from '../components/settings/ConfigForm';
import TimeControlPanel from '../components/settings/TimeControlPanel';
import DatabasePanel from '../components/settings/DatabasePanel';

export default function SettingsPage() {
  const {
    config, time, updateConfig, updateTime,
    wipeDatabase, isWiping,
    exportDatabase, isExporting,
    importDatabase, isImporting,
  } = useSettingsViewModel();

  return (
    <div>
      <h2 className="text-xl font-bold text-gray-900 dark:text-white">Settings</h2>
      <div className="mt-4 grid grid-cols-1 gap-6 lg:grid-cols-3">
        <ConfigForm config={config} onUpdate={updateConfig} />
        <TimeControlPanel time={time} onUpdate={updateTime} />
        <DatabasePanel
          onWipe={wipeDatabase}
          onExport={exportDatabase}
          onImport={importDatabase}
          isWiping={isWiping}
          isExporting={isExporting}
          isImporting={isImporting}
        />
      </div>
    </div>
  );
}
