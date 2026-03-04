import { Outlet } from 'react-router-dom';
import Sidebar from './Sidebar';
import { useToastStore } from '../stores/toastStore';

export default function AppLayout() {
  const { toasts, removeToast } = useToastStore();

  return (
    <div className="flex h-screen bg-gray-50 dark:bg-gray-950 text-gray-900 dark:text-white">
      <Sidebar />
      <main className="flex-1 overflow-y-auto p-6">
        <Outlet />
      </main>
      {/* Toast container */}
      <div className="fixed bottom-4 right-4 z-50 space-y-2">
        {toasts.map((toast) => (
          <div
            key={toast.id}
            className={`rounded-lg px-4 py-3 text-sm font-medium shadow-lg ${
              toast.type === 'error'
                ? 'bg-red-600 text-white'
                : toast.type === 'success'
                  ? 'bg-green-600 text-white'
                  : 'bg-gray-800 text-white'
            }`}
          >
            <div className="flex items-center gap-2">
              <span className="flex-1">{toast.message}</span>
              <button onClick={() => removeToast(toast.id)} className="text-white/80 hover:text-white">
                &times;
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
