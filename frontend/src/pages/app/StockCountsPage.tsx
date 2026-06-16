import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { useTenant } from '../../contexts/TenantContext';
import { inventoryApi } from '../../api/client';
import StockItemRow from '../../components/StockItemRow';
import LoadingSpinner from '../../components/LoadingSpinner';
import type { StockItem, StockCountEntry } from '../../types';
import { toast } from 'sonner';

export default function StockCountsPage() {
  const { user } = useAuth();
  const { activeTenant } = useTenant();
  const [items, setItems] = useState<StockItem[]>([]);
  const [entries, setEntries] = useState<Map<string, StockCountEntry>>(new Map());
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [search, setSearch] = useState('');

  const locationId = activeTenant?.tenantId || '';

  useEffect(() => {
    if (!locationId) return;
    setLoading(true);
    inventoryApi.listStockItems(locationId)
      .then((data) => {
        setItems(data.stockItems);
        const initial = new Map<string, StockCountEntry>();
        data.stockItems.forEach((item) => {
          initial.set(item.id, { stockItemId: item.id, quantity: 0, unit: item.unit });
        });
        setEntries(initial);
      })
      .catch(() => toast.error('Failed to load stock items'))
      .finally(() => setLoading(false));
  }, [locationId]);

  const handleEntryChange = useCallback((itemId: string, entry: StockCountEntry) => {
    setEntries((prev) => {
      const next = new Map(prev);
      next.set(itemId, entry);
      return next;
    });
  }, []);

  const handleSubmit = async () => {
    if (!locationId) return;
    setSubmitting(true);
    const counts: StockCountEntry[] = Array.from(entries.values()).filter(
      (e) => e.quantity > 0
    );
    if (counts.length === 0) {
      toast.error('Enter at least one count');
      setSubmitting(false);
      return;
    }
    try {
      await inventoryApi.submitStockCount({
        locationId,
        shift: 'close',
        counts,
      });
      toast.success('Stock count submitted');
    } catch {
      toast.error('Failed to submit stock count');
    } finally {
      setSubmitting(false);
    }
  };

  const filteredItems = items.filter(
    (item) =>
      item.name.toLowerCase().includes(search.toLowerCase()) ||
      item.category.toLowerCase().includes(search.toLowerCase())
  );

  const grouped = filteredItems.reduce<Record<string, StockItem[]>>((acc, item) => {
    if (!acc[item.category]) acc[item.category] = [];
    acc[item.category].push(item);
    return acc;
  }, {});

  if (loading) {
    return <LoadingSpinner size="lg" />;
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-white">Stock Count</h1>
          <p className="text-sm text-dark-400 mt-1">
            {user?.displayName} &middot; Closing shift
          </p>
        </div>
        <button
          onClick={handleSubmit}
          disabled={submitting}
          className="px-4 py-2 bg-primary-500 hover:bg-primary-600 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors"
        >
          {submitting ? 'Submitting...' : 'Submit Count'}
        </button>
      </div>

      <div className="mb-6">
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search items..."
          className="w-full bg-dark-800 border border-dark-700 rounded-xl px-4 py-3 text-sm text-white placeholder-dark-500 focus:outline-none focus:border-primary-500"
        />
      </div>

      {items.length === 0 && (
        <div className="text-center py-16">
          <p className="text-dark-400">No stock items yet. Add some in settings.</p>
        </div>
      )}

      <div className="space-y-6">
        {Object.entries(grouped).map(([category, categoryItems]) => (
          <div key={category}>
            <h3 className="text-sm font-medium text-dark-400 uppercase tracking-wider mb-3 px-1">
              {category}
            </h3>
            <div className="space-y-3">
              {categoryItems.map((item) => (
                <StockItemRow
                  key={item.id}
                  item={item}
                  entry={entries.get(item.id) || null}
                  onChange={(entry) => handleEntryChange(item.id, entry)}
                />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
