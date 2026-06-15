import { useState, useEffect } from 'react';
import { useTenant } from '../../contexts/TenantContext';
import { inventoryApi } from '../../api/client';
import LoadingSpinner from '../../components/LoadingSpinner';
import ConflictBadge from '../../components/ConflictBadge';
import type { ForecastItem } from '../../types';
import { toast } from 'sonner';

export default function ForecastPage() {
  const { activeTenant } = useTenant();
  const [forecast, setForecast] = useState<ForecastItem[]>([]);
  const [loading, setLoading] = useState(true);

  const locationId = activeTenant?.tenantId || '';

  useEffect(() => {
    if (!locationId) return;
    setLoading(true);
    inventoryApi.getForecast(locationId)
      .then((data) => setForecast(data.forecast))
      .catch(() => toast.error('Failed to load forecast'))
      .finally(() => setLoading(false));
  }, [locationId]);

  const totalToOrder = forecast.reduce((sum, f) => sum + f.suggestedOrderQty, 0);
  const coldStartItems = forecast.filter((f) => f.countDays < 2);

  if (loading) {
    return <LoadingSpinner size="lg" />;
  }

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-white">Forecast & Orders</h1>
        <p className="text-sm text-dark-400 mt-1">
          Based on the last 7 days of stock counts
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
        <div className="bg-dark-900/50 border border-dark-800 rounded-xl p-4">
          <p className="text-sm text-dark-400">Items to Order</p>
          <p className="text-2xl font-bold text-white mt-1">{forecast.length}</p>
        </div>
        <div className="bg-dark-900/50 border border-dark-800 rounded-xl p-4">
          <p className="text-sm text-dark-400">Total Est. Qty</p>
          <p className="text-2xl font-bold text-white mt-1">{totalToOrder.toFixed(1)}</p>
        </div>
        <div className="bg-dark-900/50 border border-dark-800 rounded-xl p-4">
          <p className="text-sm text-dark-400">Insufficient Data</p>
          <p className="text-2xl font-bold text-white mt-1">{coldStartItems.length}</p>
        </div>
      </div>

      {forecast.length === 0 && (
        <div className="text-center py-16">
          <p className="text-dark-400">
            Not enough data yet. Submit at least 2 stock counts to see forecasts.
          </p>
        </div>
      )}

      <div className="space-y-2">
        {forecast.map((item) => (
          <div
            key={item.stockItemId}
            className="bg-dark-900/50 border border-dark-800 rounded-xl p-4 flex items-center justify-between"
          >
            <div className="flex-1">
              <div className="flex items-center gap-2">
                <h4 className="text-sm font-medium text-white">{item.itemName}</h4>
                <span className="text-xs text-dark-500">{item.category}</span>
              </div>
              <p className="text-xs text-dark-400 mt-0.5">
                Avg: {item.sevenDayAvg.toFixed(1)} {item.unit} &middot; Last: {item.lastQty} &middot; {item.countDays}d of data
              </p>
            </div>
            <div className="text-right">
              <p className="text-lg font-semibold text-primary-400">
                {item.suggestedOrderQty.toFixed(1)}
              </p>
              <p className="text-xs text-dark-500">{item.unit}</p>
            </div>
          </div>
        ))}
      </div>

      {coldStartItems.length > 0 && (
        <div className="mt-6">
          <ConflictBadge count={coldStartItems.length} />
          <p className="text-xs text-dark-500 mt-2">
            Items with insufficient data use par level as suggested order qty.
            Keep counting daily for accurate forecasts.
          </p>
        </div>
      )}
    </div>
  );
}
