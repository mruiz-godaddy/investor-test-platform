import { useState } from 'react';
import type { TimeResponse, TimeUpdate } from '../../../domain/entities/ServerTime';
import { formatDateTime } from '../../../lib/formatters';

interface Props {
  time: TimeResponse | undefined;
  onUpdate: (update: TimeUpdate) => void;
}

export default function TimeControlPanel({ time, onUpdate }: Props) {
  const [offsetSeconds, setOffsetSeconds] = useState(0);
  const [freezeAt, setFreezeAt] = useState('');

  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4">
      <h3 className="text-sm font-semibold text-gray-900">Time Control</h3>
      {time && (
        <div className="mt-2 text-sm">
          <p className="text-gray-500">Server Time: <span className="font-medium text-gray-900">{formatDateTime(time.serverTime)}</span></p>
          <p className="text-gray-500">Mode: <span className={`font-medium ${
            time.mode === 'frozen' ? 'text-blue-600' : time.mode === 'offset' ? 'text-amber-600' : 'text-green-600'
          }`}>{time.mode}</span></p>
        </div>
      )}

      <div className="mt-4 space-y-3">
        <div className="flex items-end gap-2">
          <div className="flex-1">
            <label className="block text-xs font-medium text-gray-700">Offset (seconds)</label>
            <input type="number" value={offsetSeconds} onChange={(e) => setOffsetSeconds(Number(e.target.value))}
              className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
          </div>
          <button onClick={() => onUpdate({ offsetSeconds })}
            className="rounded-md bg-amber-600 px-3 py-2 text-sm font-semibold text-white hover:bg-amber-500">Offset</button>
        </div>

        <div className="flex items-end gap-2">
          <div className="flex-1">
            <label className="block text-xs font-medium text-gray-700">Freeze At</label>
            <input type="datetime-local" value={freezeAt} onChange={(e) => setFreezeAt(e.target.value)}
              className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
          </div>
          <button onClick={() => onUpdate({ freezeAt: new Date(freezeAt).toISOString() })}
            disabled={!freezeAt}
            className="rounded-md bg-blue-600 px-3 py-2 text-sm font-semibold text-white hover:bg-blue-500 disabled:opacity-50">Freeze</button>
        </div>

        <button onClick={() => onUpdate({ reset: true })}
          className="rounded-md bg-gray-600 px-4 py-2 text-sm font-semibold text-white hover:bg-gray-500">Reset Time</button>
      </div>
    </div>
  );
}
