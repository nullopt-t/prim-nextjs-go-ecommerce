import { useState, useEffect } from "react";
import { Plus, Minus, Search, History, AlertTriangle } from "lucide-react";
import { adminService } from "@/services/admin";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

export function AdminInventory() {
  const t = useTranslations("admin");
  const [products, setProducts] = useState<any[]>([]);
  const [logs, setLogs] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [selectedProduct, setSelectedProduct] = useState<any>(null);
  const [adjustAmount, setAdjustAmount] = useState(0);
  const [adjustReason, setAdjustReason] = useState("restock");
  const [auditProduct, setAuditProduct] = useState<any>(null);
  const [auditLogs, setAuditLogs] = useState<any[]>([]);
  const [loadingAudit, setLoadingAudit] = useState(false);

  const handleViewAudit = async (prod: any) => {
    setAuditProduct(prod);
    setLoadingAudit(true);
    try {
      const res = await adminService.getInventoryLedgers({
        limit: 50,
        variantId: prod.variantId,
        productId: prod.productId
      });
      const list = Array.isArray(res?.data) ? res.data : (Array.isArray(res) ? res : []);
      setAuditLogs(list.map((l: any) => ({
        id: l.id,
        product: `${l.productName} (${l.sku})`,
        sku: l.sku,
        change: l.quantity > 0 ? `+${l.quantity}` : `${l.quantity}`,
        quantity: l.quantity,
        type: l.reason ? l.reason.toUpperCase() : "AUDIT",
        date: l.createdAt ? l.createdAt.replace("T", " ").substring(0, 16) : new Date().toISOString().substring(0, 16),
        user: "System / Admin"
      })));
    } catch {
      toast.error(t("inventory.toast.auditFailed"));
    } finally {
      setLoadingAudit(false);
    }
  };

  const loadInventoryData = async () => {
    try {
      setLoading(true);
      const [invRes, logsRes]: [any, any] = await Promise.all([
        adminService.getInventory({ pageSize: 50 }),
        adminService.getInventoryLedgers(30)
      ]);

      const items = Array.isArray(invRes?.data) ? invRes.data : (Array.isArray(invRes) ? invRes : []);
      setProducts(items.map((it: any) => ({
        id: it.variantId || it.productId,
        variantId: it.variantId,
        productId: it.productId,
        name: `${it.productName} (${it.variantTitle || "Default"})`,
        sku: it.sku,
        stock: it.onHandQuantity !== undefined ? it.onHandQuantity : 25,
        available: it.availableQuantity !== undefined ? it.availableQuantity : 25,
        image: it.thumbnail || "http://localhost:9000/catalog/products/macbook-air-15-m3.jpg",
        lowStockThreshold: 10
      })));

      const ledgerList = Array.isArray(logsRes?.data) ? logsRes.data : (Array.isArray(logsRes) ? logsRes : []);
      setLogs(ledgerList.map((l: any) => ({
        id: l.id,
        product: `${l.productName} (${l.sku})`,
        sku: l.sku,
        change: l.quantity > 0 ? `+${l.quantity}` : `${l.quantity}`,
        type: l.reason ? l.reason.toUpperCase() : "AUDIT",
        date: l.createdAt ? l.createdAt.replace("T", " ").substring(0, 16) : new Date().toISOString().substring(0, 16),
        user: "Warehouse Manager"
      })));
    } catch {
      toast.error(t("inventory.toast.loadFailed"));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadInventoryData();
  }, []);

  const handleAdjustStock = (prod: any) => {
    setSelectedProduct(prod);
    setAdjustAmount(0);
    setAdjustReason("restock");
  };

  const handleApplyAdjustment = async () => {
    if (!selectedProduct || adjustAmount === 0) return;

    try {
      if (selectedProduct.variantId) {
        await adminService.adjustInventory(selectedProduct.variantId, adjustAmount, adjustReason);
      } else if (selectedProduct.productId) {
        await adminService.adjustProductInventory(selectedProduct.productId, adjustAmount, adjustReason);
      } else {
        toast.error(t("inventory.toast.noId"));
        return;
      }
      toast.success(t("inventory.toast.adjusted", { name: selectedProduct.name }));
      setSelectedProduct(null);
      loadInventoryData();
    } catch (err: any) {
      toast.error(err?.response?.data?.message || t("inventory.toast.adjustFailed"));
    }
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
          <h1 className="text-2xl font-bold text-foreground">{t("inventory.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t("inventory.subtitle")}
          </p>
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs">
        <div className="relative max-w-sm">
          <Search className="absolute left-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder={t("inventory.searchPlaceholder")}
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
            <h3 className="font-bold text-sm text-foreground">{t("inventory.currentStockLevels")}</h3>
            <span className="text-xs text-muted-foreground">{t("inventory.catalogItems", { count: products.length })}</span>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
                <tr>
                  <th className="py-3 pl-4">{t("inventory.colProduct")}</th>
                  <th className="py-3">{t("inventory.colSku")}</th>
                  <th className="py-3">{t("inventory.colStockOnHand")}</th>
                  <th className="py-3">{t("inventory.colStatus")}</th>
                  <th className="py-3 pr-4 text-right">{t("inventory.colActions")}</th>
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
                    <td className="py-3.5 font-bold text-foreground">{t("inventory.units", { count: prod.stock })}</td>
                    <td className="py-3.5">
                      {prod.stock <= (prod.lowStockThreshold || 10) ? (
                        <span className="px-2 py-0.5 bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-400 font-bold rounded-full text-[10px] inline-flex items-center gap-1">
                          <AlertTriangle className="size-3" /> {t("status.lowStock")}
                        </span>
                      ) : (
                        <span className="px-2 py-0.5 bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400 font-bold rounded-full text-[10px]">
                          {t("status.inStock")}
                        </span>
                      )}
                    </td>
                    <td className="py-3.5 pr-4 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          onClick={() => handleViewAudit(prod)}
                          className="px-2.5 py-1.5 bg-secondary/80 hover:bg-secondary text-foreground text-xs font-semibold rounded-lg border border-border transition-colors flex items-center gap-1.5"
                          title={t("inventory.viewAudit")}
                        >
                          <History className="size-3.5 text-accent-brand" />
                          <span>{t("inventory.audit")}</span>
                        </button>
                        <button
                          onClick={() => handleAdjustStock(prod)}
                          className="px-3 py-1.5 bg-accent-brand text-white hover:bg-accent-brand/90 text-xs font-semibold rounded-lg shadow-xs transition-colors"
                        >
                          {t("inventory.adjust")}
                        </button>
                      </div>
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
              <h3 className="font-bold text-sm text-foreground">{t("inventory.globalAuditLog")}</h3>
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
                    <span>{t("inventory.reason", { type: log.type })}</span>
                    <span>{t("inventory.by", { user: log.user })}</span>
                  </div>
                  <div className="text-[10px] text-muted-foreground font-mono">{log.date}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Product Audit History Modal */}
      {auditProduct && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-xl w-full p-6 shadow-xl space-y-4 max-h-[85vh] flex flex-col">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2.5">
                <div className="p-2 bg-accent-brand/10 text-accent-brand rounded-xl">
                  <History className="size-5" />
                </div>
                <div>
                  <h3 className="font-bold text-base text-foreground">{t("inventory.auditTitle")}</h3>
                  <p className="text-xs text-muted-foreground">{auditProduct.name} &bull; <span className="font-mono">{auditProduct.sku}</span></p>
                </div>
              </div>
              <button
                onClick={() => setAuditProduct(null)}
                className="text-muted-foreground hover:text-foreground text-sm font-semibold p-1"
              >
                ✕
              </button>
            </div>

            <div className="flex-1 overflow-y-auto space-y-2 pr-1 min-h-[220px]">
              {loadingAudit ? (
                <div className="py-12 text-center text-xs text-muted-foreground">{t("inventory.loadingAudit")}</div>
              ) : auditLogs.length === 0 ? (
                <div className="py-12 text-center text-xs text-muted-foreground">
                  {t("inventory.auditEmpty")}
                </div>
              ) : (
                auditLogs.map((entry) => (
                  <div key={entry.id} className="p-3 border border-border rounded-xl bg-secondary/15 flex items-center justify-between text-xs">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-foreground uppercase tracking-wider text-[11px]">{entry.type}</span>
                        <span className="text-[10px] text-muted-foreground font-mono">{entry.date}</span>
                      </div>
                      <div className="text-[11px] text-muted-foreground mt-0.5">
                        {t("inventory.recordedBy", { user: entry.user })}
                      </div>
                    </div>
                    <div className="text-right">
                      <span className={`text-sm font-bold ${entry.quantity > 0 ? "text-emerald-600 dark:text-emerald-400" : "text-amber-600 dark:text-amber-400"}`}>
                        {t("inventory.units", { count: entry.change })}
                      </span>
                    </div>
                  </div>
                ))
              )}
            </div>

            <div className="flex justify-end pt-3 border-t border-border">
              <button
                onClick={() => setAuditProduct(null)}
                className="px-4 py-2 text-xs font-semibold bg-secondary hover:bg-secondary/80 text-foreground border border-border rounded-xl"
              >
                {t("actions.close")}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Adjustment Modal */}
      {selectedProduct && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-sm w-full p-6 shadow-xl space-y-4">
            <h3 className="font-bold text-base text-foreground">{t("inventory.adjustTitle", { name: selectedProduct.name })}</h3>
            <div className="text-xs text-muted-foreground">{t("inventory.currentStockLabel")} <strong className="text-foreground">{t("inventory.units", { count: selectedProduct.stock })}</strong></div>

            <div className="space-y-3">
              <div>
                <label className="text-xs font-semibold text-foreground">{t("inventory.stockChange")}</label>
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
                <label className="text-xs font-semibold text-foreground">{t("inventory.reasonLabel")}</label>
                <select
                  value={adjustReason}
                  onChange={(e) => setAdjustReason(e.target.value)}
                  className="w-full px-3 py-2 mt-1 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                >
                  <option value="restock">{t("inventory.reasonRestock")}</option>
                  <option value="adjustment">{t("inventory.reasonAdjustment")}</option>
                  <option value="return">{t("inventory.reasonReturn")}</option>
                  <option value="sale">{t("inventory.reasonSale")}</option>
                </select>
              </div>
            </div>

            <div className="flex justify-end gap-3 pt-3 border-t border-border">
              <button onClick={() => setSelectedProduct(null)} className="px-4 py-2 text-xs font-semibold border border-border rounded-xl">{t("actions.cancel")}</button>
              <button onClick={handleApplyAdjustment} className="px-4 py-2 text-xs font-semibold bg-accent-brand text-white rounded-xl">{t("inventory.applyAdjustment")}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
