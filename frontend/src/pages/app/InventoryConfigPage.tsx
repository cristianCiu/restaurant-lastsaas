import { useState, useEffect } from 'react';
import { Package, Plus, Edit2, Trash2, X, Check } from 'lucide-react';
import { useAuth } from '../../contexts/AuthContext';
import { useTenant } from '../../contexts/TenantContext';
import { inventoryApi } from '../../api/client';
import LoadingSpinner from '../../components/LoadingSpinner';
import type { StockItem } from '../../types';
import { toast } from 'sonner';

interface ItemForm {
  name: string;
  category: string;
  unit: string;
  parLevel: string;
  leadTimeDays: string;
}

const emptyForm: ItemForm = { name: '', category: '', unit: '', parLevel: '', leadTimeDays: '1' };

export default function InventoryConfigPage() {
  const { user } = useAuth();
  const { activeTenant } = useTenant();
  const [items, setItems] = useState<StockItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<ItemForm>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [search, setSearch] = useState('');

  const locationId = activeTenant?.tenantId || '';

  const load = () => {
    if (!locationId) return;
    setLoading(true);
    inventoryApi.listStockItems(locationId)
      .then((data) => setItems(data.stockItems))
      .catch(() => toast.error('Failed to load stock items'))
      .finally(() => setLoading(false));
  };

  useEffect(() => { load(); }, [locationId]);

  const resetForm = () => { setForm(emptyForm); setEditingId(null); setShowForm(false); };

  const openEdit = (item: StockItem) => {
    setForm({
      name: item.name,
      category: item.category,
      unit: item.unit,
      parLevel: item.parLevel?.toString() || '',
      leadTimeDays: item.leadTimeDays.toString(),
    });
    setEditingId(item.id);
    setShowForm(true);
  };

  const handleSave = async () => {
    if (!form.name || !form.category || !form.unit) {
      toast.error('Name, category, and unit are required');
      return;
    }
    setSaving(true);
    try {
      const data = {
        locationId,
        name: form.name,
        category: form.category,
        unit: form.unit,
        parLevel: form.parLevel ? parseFloat(form.parLevel) : undefined,
        leadTimeDays: parseInt(form.leadTimeDays) || 1,
      };
      if (editingId) {
        await inventoryApi.updateStockItem(editingId, data);
        toast.success('Stock item updated');
      } else {
        await inventoryApi.createStockItem(data);
        toast.success('Stock item created');
      }
      resetForm();
      load();
    } catch {
      toast.error(editingId ? 'Failed to update item' : 'Failed to create item');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Delete "${name}"?`)) return;
    try {
      await inventoryApi.deleteStockItem(id);
      toast.success('Stock item deleted');
      load();
    } catch {
      toast.error('Failed to delete item');
    }
  };

  const filtered = items.filter(
    (i) => i.name.toLowerCase().includes(search.toLowerCase()) ||
           i.category.toLowerCase().includes(search.toLowerCase())
  );

  const grouped = filtered.reduce<Record<string, StockItem[]>>((acc, item) => {
    if (!acc[item.category]) acc[item.category] = [];
    acc[item.category].push(item);
    return acc;
  }, {});

  if (loading) return <LoadingSpinner size="lg" />;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-3">
            <Package className="w-7 h-7 text-primary-400" />
            Stock Config
          </h1>
          <p className="text-sm text-dark-400 mt-1">Manage stock items, categories, and par levels</p>
        </div>
        <button
          onClick={() => { resetForm(); setShowForm(true); }}
          className="flex items-center gap-2 px-4 py-2 bg-primary-500 hover:bg-primary-600 text-white text-sm font-medium rounded-lg transition-colors"
        >
          <Plus className="w-4 h-4" />
          Add Item
        </button>
      </div>

      {showForm && (
        <div className="mb-6 bg-dark-900/50 border border-dark-800 rounded-xl p-5">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-semibold text-white">{editingId ? 'Edit Item' : 'New Item'}</h3>
            <button onClick={resetForm} className="text-dark-400 hover:text-white transition-colors">
              <X className="w-5 h-5" />
            </button>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-5 gap-3 mb-4">
            <div>
              <label className="block text-xs text-dark-400 mb-1">Name *</label>
              <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })}
                className="w-full bg-dark-800 border border-dark-700 rounded-lg px-3 py-2 text-sm text-white placeholder-dark-500 focus:outline-none focus:border-primary-500" />
            </div>
            <div>
              <label className="block text-xs text-dark-400 mb-1">Category *</label>
              <input value={form.category} onChange={(e) => setForm({ ...form, category: e.target.value })}
                className="w-full bg-dark-800 border border-dark-700 rounded-lg px-3 py-2 text-sm text-white placeholder-dark-500 focus:outline-none focus:border-primary-500" />
            </div>
            <div>
              <label className="block text-xs text-dark-400 mb-1">Unit *</label>
              <input value={form.unit} onChange={(e) => setForm({ ...form, unit: e.target.value })}
                placeholder="cases, lbs, etc."
                className="w-full bg-dark-800 border border-dark-700 rounded-lg px-3 py-2 text-sm text-white placeholder-dark-500 focus:outline-none focus:border-primary-500" />
            </div>
            <div>
              <label className="block text-xs text-dark-400 mb-1">Par Level</label>
              <input type="number" step="0.1" min="0" value={form.parLevel}
                onChange={(e) => setForm({ ...form, parLevel: e.target.value })}
                className="w-full bg-dark-800 border border-dark-700 rounded-lg px-3 py-2 text-sm text-white placeholder-dark-500 focus:outline-none focus:border-primary-500" />
            </div>
            <div>
              <label className="block text-xs text-dark-400 mb-1">Lead Time (days)</label>
              <input type="number" min="1" value={form.leadTimeDays}
                onChange={(e) => setForm({ ...form, leadTimeDays: e.target.value })}
                className="w-full bg-dark-800 border border-dark-700 rounded-lg px-3 py-2 text-sm text-white placeholder-dark-500 focus:outline-none focus:border-primary-500" />
            </div>
          </div>
          <div className="flex justify-end gap-2">
            <button onClick={resetForm} className="px-4 py-2 text-sm text-dark-400 hover:text-white transition-colors">Cancel</button>
            <button onClick={handleSave} disabled={saving}
              className="flex items-center gap-2 px-4 py-2 bg-primary-500 hover:bg-primary-600 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors">
              <Check className="w-4 h-4" />
              {saving ? 'Saving...' : editingId ? 'Update' : 'Create'}
            </button>
          </div>
        </div>
      )}

      <div className="mb-6">
        <input type="text" value={search} onChange={(e) => setSearch(e.target.value)}
          placeholder="Search items..."
          className="w-full bg-dark-800 border border-dark-700 rounded-xl px-4 py-3 text-sm text-white placeholder-dark-500 focus:outline-none focus:border-primary-500" />
      </div>

      {items.length === 0 && (
        <div className="text-center py-16">
          <Package className="w-12 h-12 text-dark-600 mx-auto mb-4" />
          <p className="text-dark-400">No stock items yet. Click "Add Item" to get started.</p>
        </div>
      )}

      <div className="space-y-6">
        {Object.entries(grouped).map(([category, categoryItems]) => (
          <div key={category}>
            <h3 className="text-sm font-medium text-dark-400 uppercase tracking-wider mb-3 px-1">{category}</h3>
            <div className="space-y-2">
              {categoryItems.map((item) => (
                <div key={item.id} className="bg-dark-900/50 border border-dark-800 rounded-xl px-4 py-3 flex items-center justify-between">
                  <div className="flex items-center gap-4">
                    <div>
                      <p className="text-sm font-medium text-white">{item.name}</p>
                      <p className="text-xs text-dark-500">{item.unit}{item.parLevel != null ? ` · Par: ${item.parLevel}` : ''}{item.leadTimeDays > 0 ? ` · Lead: ${item.leadTimeDays}d` : ''}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-1">
                    <button onClick={() => openEdit(item)} className="p-2 text-dark-400 hover:text-white transition-colors rounded-lg hover:bg-dark-800">
                      <Edit2 className="w-4 h-4" />
                    </button>
                    <button onClick={() => handleDelete(item.id, item.name)} className="p-2 text-dark-400 hover:text-red-400 transition-colors rounded-lg hover:bg-dark-800">
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
