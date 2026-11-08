import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback } from "react";
import { Plus, Minus, Search } from "lucide-react";
import { adminService } from "@/modules/admin/api/adminService";
import type { InventoryLedger } from "@/modules/admin/types";
import { useLocale, useTranslations } from "next-intl";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill } from "@/shared/ui/statusPill";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";
import { formatDateTime } from "@/shared/lib/datetime";
import { CustomSelect } from "@/shared/ui/customSelect";

interface InventoryItem {
  variantId: string;
  productId: string;
  productName: string;
  variantTitle?: string;
  sku: string;
  thumbnail?: string;
  onHandQuantity: number;
  availableQuantity: number;
}

interface InventoryProduct {
  id: string;
  variantId: string;
  productId: string;
  name: string;
  sku: string;
  stock: number;
  available: number;
  image: string;
  lowStockThreshold: number;
}

interface InventoryLogEntry {
  id: string;
  product: string;
  sku: string;
  change: string;
  quantity: number;
  type: string;
  date: string;
  user: string;
}

export function AdminInventory() {
  const t = useTranslations("admin");
  const locale = useLocale();
  const [products, setProducts] = useState<InventoryProduct[]>([]);
  const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [logs, setLogs] = useState<InventoryLogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [selectedProduct, setSelectedProduct] = useState<InventoryProduct | null>(null);
  const [adjustAmount, setAdjustAmount] = useState(0);
  const [adjustReason, setAdjustReason] = useState("restock");
  const [auditProduct, setAuditProduct] = useState<InventoryProduct | null>(null);
  const [auditLogs, setAuditLogs] = useState<InventoryLogEntry[]>([]);
  const [loadingAudit, setLoadingAudit] = useState(false);
  const [loadError, setLoadError] = useState(false);
  const { cursor, hasPrev, next, prev, page } = useCursorPagination();

  const handleViewAudit = async (prod: InventoryProduct) => {
    setAuditProduct(prod);
    setLoadingAudit(true);
    try {
      const res = await adminService.getInventoryLedgers({
        pageSize: 50,
        variantId: prod.variantId,
        productId: prod.productId
      });
      const list = res?.data ?? [];
      setAuditLogs(list.map((l: InventoryLedger) => ({
        id: l.id,
        product: l.sku ? `${l.productName ?? ""} (${l.sku})` : (l.productName ?? "—"),
        sku: l.sku || "",
        change: l.quantity > 0 ? `+${l.quantity}` : `${l.quantity}`,
        quantity: l.quantity,
        type: l.reason ? l.reason.toUpperCase() : "AUDIT",
        date: formatDateTime(l.createdAt, locale) || "—",
        user: "—"
      })));
    } catch {
      notify.error(t("inventory.toast.auditFailed"));
    } finally {
      setLoadingAudit(false);
    }
  };

  const loadInventoryData = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);
      const [invRes, logsRes] = await Promise.all([
        adminService.getInventory({ cursor: cursor || undefined, pageSize: 50 }),
        adminService.getInventoryLedgers(30)
      ]);

      const items: InventoryItem[] = (invRes?.data ?? []) as unknown as InventoryItem[];
      setProducts(items.map((it) => ({
        id: it.variantId || it.productId,
        variantId: it.variantId,
        productId: it.productId,
        name: it.variantTitle ? `${it.productName} (${it.variantTitle})` : it.productName,
        sku: it.sku || "—",
        stock: it.onHandQuantity ?? 0,
        available: it.availableQuantity ?? 0,
        image: it.thumbnail || "",
        lowStockThreshold: 10
      })));
      if (invRes?.meta) setMeta(invRes.meta as CursorMeta);

      const ledgerList = logsRes?.data ?? [];
      setLogs(ledgerList.map((l: InventoryLedger) => ({
        id: l.id,
        product: l.sku ? `${l.productName ?? ""} (${l.sku})` : (l.productName ?? "—"),
        sku: l.sku || "",
        quantity: l.quantity,
        change: l.quantity > 0 ? `+${l.quantity}` : `${l.quantity}`,
        type: l.reason ? l.reason.toUpperCase() : "AUDIT",
        date: formatDateTime(l.createdAt, locale) || "—",
        user: "—"
      })));
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [cursor]);

  useEffect(() => {
    loadInventoryData();
  }, [loadInventoryData]);

  // Real-time listener for live inventory adjustments and audit log events
  useEffect(() => {
    const handleStockUpdate = (e: Event) => {
      const customEvent = e as CustomEvent;
      const detail = customEvent.detail;

      // Optimistically update stock count in table if present
      if (detail && detail.variantId) {
        setProducts((prev) =>
          prev.map((p) =>
            p.variantId === detail.variantId
              ? { ...p, stock: detail.currentStock, available: Math.max(0, detail.currentStock) }
              : p
          )
        );
      }

      // Refresh inventory list and audit trail
      loadInventoryData();

      // If audit modal is currently open for this variant, refresh it too
      if (auditProduct && detail && auditProduct.variantId === detail.variantId) {
        handleViewAudit(auditProduct);
      }
    };

    window.addEventListener("admin-stock-updated", handleStockUpdate);
    return () => {
      window.removeEventListener("admin-stock-updated", handleStockUpdate);
    };
  }, [loadInventoryData, auditProduct]);

  const handleAdjustStock = (prod: InventoryProduct) => {
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
        notify.error(t("inventory.toast.noId"));
        return;
      }
      notify.success(t("inventory.toast.adjusted", { name: selectedProduct.name }));
      setSelectedProduct(null);
      loadInventoryData();
    } catch (err) {
      notify.error(apiErrorMessage(err, t("inventory.toast.adjustFailed")));
    }
  };

  const filtered = products.filter(
    (p) =>
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.sku.toLowerCase().includes(search.toLowerCase())
  );

  if (loadError) {
    return <ErrorState onRetry={loadInventoryData} />;
  }

  return (
    <div className="space-y-5">
      <PageHeader title={t("inventory.title")} />

      <div className="relative w-full sm:max-w-xs">
        <Search className="absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <input
          type="text"
          placeholder={t("inventory.searchPlaceholder")}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full rounded-md border border-border bg-background py-2 pe-3 ps-9 text-sm text-foreground placeholder:text-muted-foreground focus:outline-hidden focus:ring-2 focus:ring-ring/30"
        />
      </div>

      <div className="grid grid-cols-1 gap-5 lg:grid-cols-3">
        {/* Stock List Table */}
        <div className="overflow-hidden rounded-lg border border-border bg-background lg:col-span-2">
          <div className="flex items-center justify-between border-b border-border p-4">
            <h2 className="text-sm font-semibold text-foreground">{t("inventory.currentStockLevels")}</h2>
            <span className="text-xs text-muted-foreground">{t("inventory.catalogItems", { count: products.length })}</span>
          </div>
          <div className="overflow-x-auto">
            <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
              <thead className="text-xs text-muted-foreground">
                <tr className="border-b border-border">
                  <th className="py-2.5 ps-4 text-start font-medium">{t("inventory.colProduct")}</th>
                  <th className="py-2.5 text-start font-medium">{t("inventory.colSku")}</th>
                  <th className="py-2.5 text-start font-medium">{t("inventory.colStockOnHand")}</th>
                  <th className="py-2.5 text-start font-medium">{t("inventory.colStatus")}</th>
                  <th className="py-2.5 pe-4 text-end font-medium">{t("inventory.colActions")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {filtered.map((prod) => (
                  <tr key={prod.id} className="transition-colors hover:bg-secondary/30">
                    <td className="py-3 ps-4">
                      <div className="flex items-center gap-3">
                        {prod.image ? (
                          <img
                            src={prod.image}
                            alt=""
                            loading="lazy"
                            className="size-9 shrink-0 rounded-md border border-border object-cover"
                          />
                        ) : (
                          <div className="size-9 shrink-0 rounded-md border border-border bg-secondary/40" />
                        )}
                        <span className="truncate font-medium text-foreground">{prod.name}</span>
                      </div>
                    </td>
                    <td className="py-3 font-mono text-xs text-muted-foreground">{prod.sku}</td>
                    <td className="py-3 font-medium text-foreground tabular-nums">{t("inventory.units", { count: prod.stock })}</td>
                    <td className="py-3">
                      <StatusPill tone={prod.stock <= prod.lowStockThreshold ? "warning" : "success"}>
                        {prod.stock <= prod.lowStockThreshold ? t("status.lowStock") : t("status.inStock")}
                      </StatusPill>
                    </td>
                    <td className="py-3 pe-4 text-end">
                      <div className="inline-flex items-center gap-1">
                        <button
                          onClick={() => handleViewAudit(prod)}
                          className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                        >
                          {t("inventory.audit")}
                        </button>
                        <button
                          onClick={() => handleAdjustStock(prod)}
                          className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                        >
                          {t("inventory.adjust")}
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </ResponsiveTable>
          </div>

          <CursorPager
            hasPrev={hasPrev}
            hasMore={meta.hasMore}
            page={page}
            onPrev={prev}
            onNext={() => next(meta.nextCursor)}
            loading={loading}
            shown={filtered.length}
          />
        </div>

        {/* Audit Log Timeline */}
        <div className="rounded-lg border border-border bg-background">
          <div className="border-b border-border p-4">
            <h2 className="text-sm font-semibold text-foreground">{t("inventory.globalAuditLog")}</h2>
          </div>
          <div className="max-h-[420px] space-y-0 divide-y divide-border overflow-y-auto">
            {logs.length === 0 ? (
              <p className="p-6 text-center text-sm text-muted-foreground">{t("inventory.auditEmpty")}</p>
            ) : (
              logs.map((log) => (
                <div key={log.id} className="space-y-1 p-4 text-xs">
                  <div className="flex items-start justify-between gap-3">
                    <span className="min-w-0 truncate font-medium text-foreground">{log.product}</span>
                    <span className={`shrink-0 font-medium tabular-nums ${log.change.startsWith("+") ? "text-emerald-700 dark:text-emerald-400" : "text-amber-700 dark:text-amber-400"}`}>
                      {log.change}
                    </span>
                  </div>
                  <div className="text-muted-foreground">{t("inventory.reason", { type: log.type })}</div>
                  <div className="font-mono text-muted-foreground">{log.date}</div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>

      {/* Product Audit History Modal */}
      {auditProduct && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="flex max-h-[85vh] w-full max-w-xl flex-col space-y-4 rounded-lg border border-border bg-background p-5 shadow-xl">
            <div className="flex items-start justify-between border-b border-border pb-3">
              <div>
                <h3 className="text-sm font-semibold text-foreground">{t("inventory.auditTitle")}</h3>
                <p className="mt-0.5 text-xs text-muted-foreground">
                  {auditProduct.name} · <span className="font-mono">{auditProduct.sku}</span>
                </p>
              </div>
              <button
                onClick={() => setAuditProduct(null)}
                className="text-sm font-medium text-muted-foreground hover:text-foreground"
                aria-label={t("actions.close")}
              >
                ✕
              </button>
            </div>

            <div className="min-h-[220px] flex-1 space-y-2 overflow-y-auto pe-1">
              {loadingAudit ? (
                <div className="py-12 text-center text-sm text-muted-foreground">{t("inventory.loadingAudit")}</div>
              ) : auditLogs.length === 0 ? (
                <div className="py-12 text-center text-sm text-muted-foreground">
                  {t("inventory.auditEmpty")}
                </div>
              ) : (
                auditLogs.map((entry) => (
                  <div key={entry.id} className="flex items-center justify-between rounded-md border border-border px-3 py-2 text-xs">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-medium uppercase tracking-wider text-foreground">{entry.type}</span>
                        <span className="font-mono text-muted-foreground">{entry.date}</span>
                      </div>
                    </div>
                    <span className={`text-sm font-medium tabular-nums ${entry.quantity > 0 ? "text-emerald-700 dark:text-emerald-400" : "text-amber-700 dark:text-amber-400"}`}>
                      {t("inventory.units", { count: entry.change })}
                    </span>
                  </div>
                ))
              )}
            </div>

            <div className="flex justify-end border-t border-border pt-3">
              <button
                onClick={() => setAuditProduct(null)}
                className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground hover:bg-secondary"
              >
                {t("actions.close")}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Adjustment Modal */}
      {selectedProduct && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-sm space-y-4 rounded-lg border border-border bg-background p-5 shadow-xl">
            <h3 className="text-sm font-semibold text-foreground">{t("inventory.adjustTitle", { name: selectedProduct.name })}</h3>
            <div className="text-sm text-muted-foreground">{t("inventory.currentStockLabel")} <strong className="font-medium text-foreground">{t("inventory.units", { count: selectedProduct.stock })}</strong></div>

            <div className="space-y-3">
              <div>
                <label className="text-sm font-medium text-foreground">{t("inventory.stockChange")}</label>
                <div className="mt-1 flex items-center gap-2">
                  <button onClick={() => setAdjustAmount((prev) => prev - 1)} className="rounded-md border border-border p-2 text-foreground hover:bg-secondary">
                    <Minus className="size-4" />
                  </button>
                  <input
                    type="number"
                    value={adjustAmount}
                    onChange={(e) => setAdjustAmount(parseInt(e.target.value, 10) || 0)}
                    className="w-full rounded-md border border-border bg-background py-2 text-center text-sm font-medium tabular-nums text-foreground"
                  />
                  <button onClick={() => setAdjustAmount((prev) => prev + 1)} className="rounded-md border border-border p-2 text-foreground hover:bg-secondary">
                    <Plus className="size-4" />
                  </button>
                </div>
              </div>

              <div>
                <label className="text-sm font-medium text-foreground">{t("inventory.reasonLabel")}</label>
                <CustomSelect
                	value={adjustReason}
                	onChange={(value) => setAdjustReason(value)}
                	options={[
                		{ value: "restock", label: t("inventory.reasonRestock") },
                		{ value: "adjustment", label: t("inventory.reasonAdjustment") },
                		{ value: "return", label: t("inventory.reasonReturn") },
                		{ value: "sale", label: t("inventory.reasonSale") },
                	]}
/>
              </div>
            </div>

            <div className="flex justify-end gap-2 border-t border-border pt-3">
              <button onClick={() => setSelectedProduct(null)} className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground hover:bg-secondary">{t("actions.cancel")}</button>
              <button onClick={handleApplyAdjustment} className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground">{t("inventory.applyAdjustment")}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
