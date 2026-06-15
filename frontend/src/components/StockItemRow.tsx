import { useState } from 'react';
import type { StockItem, StockCountEntry } from '../types';

interface StockItemRowProps {
  item: StockItem;
  entry: StockCountEntry | null;
  onChange: (entry: StockCountEntry) => void;
}

export default function StockItemRow({ item, entry, onChange }: StockItemRowProps) {
  const [quantity, setQuantity] = useState(entry?.quantity?.toString() || '');
  const [received, setReceived] = useState(entry?.received?.toString() || '');
  const [waste, setWaste] = useState(entry?.waste?.toString() || '');

  const handleQuantityChange = (val: string) => {
    setQuantity(val);
    const qty = parseFloat(val) || 0;
    onChange({
      stockItemId: item.id,
      quantity: qty,
      unit: item.unit,
      received: received ? parseFloat(received) : undefined,
      waste: waste ? parseFloat(waste) : undefined,
    });
  };

  const handleReceivedChange = (val: string) => {
    setReceived(val);
    const qty = parseFloat(quantity) || 0;
    onChange({
      stockItemId: item.id,
      quantity: qty,
      unit: item.unit,
      received: val ? parseFloat(val) : undefined,
      waste: waste ? parseFloat(waste) : undefined,
    });
  };

  const handleWasteChange = (val: string) => {
    setWaste(val);
    const qty = parseFloat(quantity) || 0;
    onChange({
      stockItemId: item.id,
      quantity: qty,
      unit: item.unit,
      received: received ? parseFloat(received) : undefined,
      waste: val ? parseFloat(val) : undefined,
    });
  };

  return (
    <div className="bg-dark-900/50 border border-dark-800 rounded-xl p-4">
      <div className="flex items-center justify-between mb-3">
        <div>
          <h4 className="text-sm font-medium text-white">{item.name}</h4>
          <span className="text-xs text-dark-500">{item.category} &middot; {item.unit}</span>
        </div>
        {item.parLevel != null && (
          <span className="text-xs text-dark-400">Par: {item.parLevel}</span>
        )}
      </div>
      <div className="grid grid-cols-3 gap-3">
        <div>
          <label className="block text-xs text-dark-400 mb-1">Count</label>
          <input
            type="number"
            step="0.1"
            min="0"
            value={quantity}
            onChange={(e) => handleQuantityChange(e.target.value)}
            className="w-full bg-dark-800 border border-dark-700 rounded-lg px-3 py-2 text-sm text-white placeholder-dark-500 focus:outline-none focus:border-primary-500"
            placeholder="0"
          />
        </div>
        <div>
          <label className="block text-xs text-dark-400 mb-1">Received</label>
          <input
            type="number"
            step="0.1"
            min="0"
            value={received}
            onChange={(e) => handleReceivedChange(e.target.value)}
            className="w-full bg-dark-800 border border-dark-700 rounded-lg px-3 py-2 text-sm text-white placeholder-dark-500 focus:outline-none focus:border-primary-500"
            placeholder="—"
          />
        </div>
        <div>
          <label className="block text-xs text-dark-400 mb-1">Waste</label>
          <input
            type="number"
            step="0.1"
            min="0"
            value={waste}
            onChange={(e) => handleWasteChange(e.target.value)}
            className="w-full bg-dark-800 border border-dark-700 rounded-lg px-3 py-2 text-sm text-white placeholder-dark-500 focus:outline-none focus:border-primary-500"
            placeholder="—"
          />
        </div>
      </div>
    </div>
  );
}
