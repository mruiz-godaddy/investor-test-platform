import { useState } from 'react';

interface Props {
  onSubmit: (req: { shopperId: string; memberId: number; customerId?: string; displayName?: string }) => void;
}

export default function CreateShopperForm({ onSubmit }: Props) {
  const [shopperId, setShopperId] = useState('');
  const [memberId, setMemberId] = useState('');
  const [customerId, setCustomerId] = useState('');
  const [displayName, setDisplayName] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({
      shopperId,
      memberId: Number(memberId),
      ...(customerId && { customerId }),
      ...(displayName && { displayName }),
    });
    setShopperId('');
    setMemberId('');
    setCustomerId('');
    setDisplayName('');
  };

  return (
    <form onSubmit={handleSubmit} className="rounded-lg border border-gray-200 bg-white p-4">
      <h3 className="text-sm font-semibold text-gray-900">Create Shopper</h3>
      <div className="mt-3 grid grid-cols-2 gap-3">
        <div>
          <label className="block text-xs font-medium text-gray-700">Shopper ID *</label>
          <input required value={shopperId} onChange={(e) => setShopperId(e.target.value)}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700">Member ID *</label>
          <input required type="number" value={memberId} onChange={(e) => setMemberId(e.target.value)}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700">Customer ID</label>
          <input value={customerId} onChange={(e) => setCustomerId(e.target.value)}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label className="block text-xs font-medium text-gray-700">Display Name</label>
          <input value={displayName} onChange={(e) => setDisplayName(e.target.value)}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm" />
        </div>
      </div>
      <button type="submit" className="mt-3 rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500">
        Create
      </button>
    </form>
  );
}
