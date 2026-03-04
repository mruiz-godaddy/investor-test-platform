import { NavLink } from 'react-router-dom';
import { ROUTES } from '../router/routes';

const navItems = [
  { to: ROUTES.AUCTIONS, label: 'Auctions' },
  { to: ROUTES.LISTINGS, label: 'Listings' },
  { to: ROUTES.SHOPPERS, label: 'Shoppers' },
  { to: ROUTES.SCENARIOS, label: 'Scenarios' },
  { to: ROUTES.SETTINGS, label: 'Settings' },
  { to: ROUTES.AUTO_GENERATOR, label: 'Auto-Generator' },
];

export default function Sidebar() {
  return (
    <aside className="flex h-screen w-56 flex-col border-r border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900">
      <div className="border-b border-gray-200 dark:border-gray-700 px-4 py-4">
        <h1 className="text-lg font-bold text-gray-900 dark:text-white">Investor Test Server</h1>
        <p className="text-xs text-gray-500 dark:text-gray-400">Admin Panel</p>
      </div>
      <nav className="flex-1 space-y-1 px-2 py-4">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === '/'}
            className={({ isActive }) =>
              `block rounded-md px-3 py-2 text-sm font-medium ${
                isActive
                  ? 'bg-indigo-50 text-indigo-700'
                  : 'text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-800 hover:text-gray-900'
              }`
            }
          >
            {item.label}
          </NavLink>
        ))}
      </nav>
    </aside>
  );
}
