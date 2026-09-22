import { useState } from "react";
import { Plus, Minus, Search, History, AlertTriangle } from "lucide-react";
import {
  INITIAL_PRODUCTS,
  INITIAL_INVENTORY_LOGS,
  getLocalAdminData,
  setLocalAdminData
} from "../data/mockAdminData";
import { toast } from "sonner";

export function AdminInventory() {
  const [products, setProducts] = useState(() => getLocalAdminData("products", INITIAL_PRODUCTS));
  const [logs, setLogs] = useState(() => getLocalAdminData("inventory_logs", INITIAL_INVENTORY_LOGS));
  const [search, setSearch] = useState("");
  const [selectedProduct, setSelectedProduct] = useState(null);
  const [adjustAmount, setAdjustAmount] = useState(0);
  const [adjustReason, setAdjustReason] = useState("Restock");

  const saveProducts = (updatedProds) => {
    setProducts(updatedProds);
    setLocalAdminData("products", updatedProds);
  };

  const saveLogs = (updatedLogs) => {
    setLogs(updatedLogs);
    setLocalAdminData("inventory_logs", updatedLogs);
  };

  const handleAdjustStock = (prod) => {
    setSelectedProduct(prod);
    setAdjustAmount(0);
    setAdjustReason("Restock");
  };

  const handleApplyAdjustment = () => {
    if (!selectedProduct || adjustAmount === 0) return;

    const newStock = Math.max(0, selectedProduct.stock + adjustAmount);
    const updatedProducts = products.map((p) =>
      p.id === selectedProduct.id ? { ...p, stock: newStock } : p
    );

    const changeFormatted = adjustAmount > 0 ? `+${adjustAmount}` : `${adjustAmount}`;
    const newLog = {
      id: `inv-log-${Date.now()}`,
      product: selectedProduct.name,
      sku: selectedProduct.sku,
      change: changeFormatted,
      type: adjustReason,
      newStock,
      date: new Date().toISOString().replace("T", " ").substring(0, 16),
      user: "Admin Manager"
    };

    saveProducts(updatedProducts);
    saveLogs([newLog, ...logs]);
    setSelectedProduct(null);
    toast.success(`Inventory updated for ${selectedProduct.name}`);
  };

  const filtered = products.filter(
    (p) =>
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.sku.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Inventory & Stock Control</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Monitor real-time warehouse stock levels, perform manual adjustments, and track stock audit logs.
          </p>
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs">
        <div className="relative max-w-sm">
          <Search className="absolute left-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search inventory by product or SKU..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
          />
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Stock List Table */}
        <div className="lg:col-span-2 bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
          <div className="p-4 border-b border-border bg-secondary/20 flex items-center justify-between">
            <h3 className="font-bold text-sm text-foreground">Current Stock Levels</h3>
            <span className="text-xs text-muted-foreground">{products.length} catalog items</span>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
                <tr>
                  <th className="py-3 pl-4">Product</th>
                  <th className="py-3">SKU</th>
                  <th className="py-3">Stock On Hand</th>
                  <th className="py-3">Status</th>
                  <th className="py-3 pr-4 text-right">Adjust Stock</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {filtered.map((prod) => (
                  <tr key={prod.id} className="hover:bg-secondary/20 transition-colors">
                    <td className="py-3.5 pl-4">
                      <div className="flex items-center gap-2.5">
                        <img src={prod.image} alt={prod.name} className="size-9 rounded-lg object-cover border border-border" />
                        <span className="font-semibold text-foreground line-clamp-1">{prod.name}</span>
                      </div>
                    </td>
                    <td className="py-3.5 font-mono text-muted-foreground">{prod.sku}</td>
                    <td className="py-3.5 font-bold text-foreground">{prod.stock} units</td>
                    <td className="py-3.5">
                      {prod.stock <= (prod.lowStockThreshold || 10) ? (
                        <span className="px-2 py-0.5 bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-400 font-bold rounded-full text-[10px] inline-flex items-center gap-1">
                          <AlertTriangle className="size-3" /> Low Stock
                        </span>
                      ) : (
                        <span className="px-2 py-0.5 bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400 font-bold rounded-full text-[10px]">
                          In Stock
                        </span>
                      )}
                    </td>
                    <td className="py-3.5 pr-4 text-right">
                      <button
                        onClick={() => handleAdjustStock(prod)}
                        className="px-3 py-1.5 bg-secondary hover:bg-secondary/80 text-foreground text-xs font-semibold rounded-lg border border-border transition-colors"
                      >
                        Adjust
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Audit Log Timeline */}
        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center gap-2 mb-4 border-b border-border pb-3">
              <History className="size-4 text-accent-brand" />
              <h3 className="font-bold text-sm text-foreground">Stock Adjustment Audit Log</h3>
            </div>
            <div className="space-y-3 max-h-[400px] overflow-y-auto pr-1">
              {logs.map((log) => (
                <div key={log.id} className="p-3 border border-border rounded-xl bg-secondary/10 space-y-1 text-xs">
                  <div className="flex items-center justify-between">
                    <span className="font-semibold text-foreground">{log.product}</span>
                    <span className={`font-bold ${log.change.startsWith("+") ? "text-emerald-600" : "text-amber-600"}`}>
                      {log.change}
                    </span>
                  </div>
                  <div className="text-[10px] text-muted-foreground flex justify-between">
                    <span>Reason: {log.type}</span>
                    <span>By: {log.user}</span>
                  </div>
                  <div className="text-[10px] text-muted-foreground font-mono">{log.date}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Adjustment Modal */}
      {selectedProduct && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-sm w-full p-6 shadow-xl space-y-4">
            <h3 className="font-bold text-base text-foreground">Adjust Stock: {selectedProduct.name}</h3>
            <div className="text-xs text-muted-foreground">Current Stock: <strong className="text-foreground">{selectedProduct.stock} units</strong></div>

            <div className="space-y-3">
              <div>
                <label className="text-xs font-semibold text-foreground">Stock Change (+/-)</label>
                <div className="flex items-center gap-3 mt-1">
                  <button onClick={() => setAdjustAmount((prev) => prev - 1)} className="p-2 border border-border rounded-xl hover:bg-secondary">
                    <Minus className="size-4" />
                  </button>
                  <input
                    type="number"
                    value={adjustAmount}
                    onChange={(e) => setAdjustAmount(parseInt(e.target.value, 10) || 0)}
                    className="w-full text-center py-2 bg-secondary/30 border border-border rounded-xl text-xs font-bold text-foreground"
                  />
                  <button onClick={() => setAdjustAmount((prev) => prev + 1)} className="p-2 border border-border rounded-xl hover:bg-secondary">
                    <Plus className="size-4" />
                  </button>
                </div>
              </div>

              <div>
                <label className="text-xs font-semibold text-foreground">Reason for Adjustment</label>
                <select
                  value={adjustReason}
                  onChange={(e) => setAdjustReason(e.target.value)}
                  className="w-full px-3 py-2 mt-1 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                >
                  <option value="Restock">Warehouse Restock</option>
                  <option value="Damaged Adjustment">Damaged / Written Off</option>
                  <option value="Physical Audit Sync">Inventory Audit Sync</option>
                  <option value="Customer Return">Customer Return</option>
                </select>
              </div>
            </div>

            <div className="flex justify-end gap-3 pt-3 border-t border-border">
              <button onClick={() => setSelectedProduct(null)} className="px-4 py-2 text-xs font-semibold border border-border rounded-xl">Cancel</button>
              <button onClick={handleApplyAdjustment} className="px-4 py-2 text-xs font-semibold bg-accent-brand text-white rounded-xl">Apply Adjustment</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
