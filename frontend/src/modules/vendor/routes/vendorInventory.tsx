"use client";

import { useEffect, useState } from "react";
import {
  Plus,
  Minus,
  Search,
  AlertTriangle,
  Loader2,
  Layers3,
} from "lucide-react";
import { vendorService } from "@/modules/vendor/api/vendorService";
import type {
  StockReason,
  VendorProduct,
  VendorVariant,
} from "@/modules/vendor/types";
import { ErrorState } from "@/shared/ui/errorState";
import { apiErrorMessage, notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";

interface InventoryRow {
  product: VendorProduct;
  variant: VendorVariant;
}

export function VendorInventory() {
  const t = useTranslations("vendor");
  const [rows, setRows] = useState<InventoryRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<Error | null>(null);
  const [search, setSearch] = useState("");
  const [filterLowStock, setFilterLowStock] = useState(false);
  const [draftAdjustments, setDraftAdjustments] = useState<Record<string, string>>({});
  const [reasonByVariant, setReasonByVariant] = useState<Record<string, StockReason>>({});
  const [savingVariantId, setSavingVariantId] = useState<string | null>(null);

  const loadInventory = async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const productResponse = await vendorService.getProducts({ pageSize: 100 });
      const productVariantLists = await Promise.all(
        productResponse.data.map(async (product) => ({
          product,
          variants: (await vendorService.getVariants(product.id, { pageSize: 100 })).data,
        })),
      );

      setRows(
        productVariantLists.flatMap(({ product, variants }) =>
          variants.map((variant) => ({ product, variant })),
        ),
      );
    } catch (error) {
      setLoadError(error instanceof Error ? error : new Error(String(error)));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadInventory();
  }, []);

  const adjustStock = async (variant: VendorVariant, delta: number) => {
    if (delta === 0) return;
    setSavingVariantId(variant.id);
    try {
      await vendorService.adjustStock(variant.id, {
        quantity: delta,
        reason: reasonByVariant[variant.id] ?? "restock",
      });
      notify.success(t("toasts.inventoryUpdated"));
      await loadInventory();
    } catch (error) {
      notify.error(apiErrorMessage(error, t("variants.stockFailed")));
    } finally {
      setSavingVariantId(null);
    }
  };

  const handleDraftSubmit = (variant: VendorVariant) => {
    const parsed = Number.parseInt(draftAdjustments[variant.id] ?? "", 10);
    if (!Number.isFinite(parsed) || parsed === 0) {
      notify.error(t("variants.invalidStockAmount"));
      return;
    }
    void adjustStock(variant, parsed);
  };

  const filtered = rows.filter(({ product, variant }) => {
    const matchesSearch =
      product.title.toLowerCase().includes(search.toLowerCase()) ||
      variant.title.toLowerCase().includes(search.toLowerCase()) ||
      variant.sku.toLowerCase().includes(search.toLowerCase());
    const matchesLowStock = filterLowStock ? product.stock <= 5 : true;
    return matchesSearch && matchesLowStock;
  });

  const lowStockCount = rows.filter(({ product }) => product.stock <= 5).length;

  return (
    <div className="space-y-6 w-full">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("inventory.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t("inventory.subtitle")}</p>
        </div>
        {lowStockCount > 0 && (
          <div className="flex items-center gap-2 px-3.5 py-2 bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20 rounded-xl text-xs font-semibold">
            <AlertTriangle className="size-4 shrink-0" />
            <span>{t("inventory.lowStockWarning", { count: lowStockCount })}</span>
          </div>
        )}
      </div>

      <div className="rounded-2xl border border-border bg-background p-4 text-xs text-muted-foreground shadow-xs">
        {t("inventory.variantLevelNote")}
      </div>

      {loading ? (
        <div className="flex min-h-[40vh] items-center justify-center">
          <Loader2 className="size-8 animate-spin text-accent-brand" />
        </div>
      ) : loadError ? (
        <ErrorState
          title={t("inventory.loadTitle")}
          description={loadError.message || t("inventory.loadDescription")}
          onRetry={() => void loadInventory()}
        />
      ) : (
        <>
          <div className="bg-background border border-border rounded-2xl p-4 shadow-xs flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div className="relative max-w-sm flex-1">
              <Search className="absolute left-3.5 top-2.5 size-4 text-muted-foreground" />
              <input
                type="text"
                placeholder={t("inventory.searchPlaceholder")}
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                className="w-full pl-10 pr-4 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
            </div>
            <button
              onClick={() => setFilterLowStock(!filterLowStock)}
              className={`px-3 py-2 text-xs font-semibold rounded-xl border transition-colors ${
                filterLowStock
                  ? "bg-amber-500/10 text-amber-600 border-amber-500/30"
                  : "bg-secondary text-foreground border-border hover:bg-secondary/80"
              }`}
            >
              {filterLowStock ? t("actions.showAllItems") : t("actions.filterLowStock")}
            </button>
          </div>

          <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
            <div className="overflow-x-auto">
              <table className="w-full min-w-max text-left text-xs">
                <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
                  <tr>
                    <th className="py-3.5 pl-4">{t("inventory.table.productVariant")}</th>
                    <th className="py-3.5">{t("inventory.table.sku")}</th>
                    <th className="py-3.5">{t("inventory.table.skuStock")}</th>
                    <th className="py-3.5">{t("inventory.table.stockReason")}</th>
                    <th className="py-3.5">{t("inventory.table.status")}</th>
                    <th className="py-3.5 pr-4 text-right">{t("inventory.table.adjustment")}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/60">
                  {filtered.map(({ product, variant }) => (
                    <tr key={variant.id} className="hover:bg-secondary/20 transition-colors">
                      <td className="py-3.5 pl-4">
                        <div className="flex items-center gap-2 font-semibold text-foreground">
                          <Layers3 className="size-4 text-muted-foreground" />
                          {product.title}
                        </div>
                        <div className="text-[10px] text-muted-foreground">{variant.title}</div>
                      </td>
                      <td className="py-3.5 font-mono text-muted-foreground">{variant.sku}</td>
                      <td className="py-3.5 font-bold text-base text-foreground">{product.stock}</td>
                      <td className="py-3.5">
                        <select
                          value={reasonByVariant[variant.id] ?? "restock"}
                          onChange={(event) =>
                            setReasonByVariant((current) => ({
                              ...current,
                              [variant.id]: event.target.value as StockReason,
                            }))
                          }
                          className="px-2 py-1.5 bg-secondary/40 border border-border rounded-lg text-xs"
                          aria-label={t("inventory.table.stockReason")}
                        >
                          <option value="restock">{t("inventory.reasonRestock")}</option>
                          <option value="return">{t("inventory.reasonReturn")}</option>
                          <option value="adjustment">{t("inventory.reasonAdjustment")}</option>
                        </select>
                      </td>
                      <td className="py-3.5">
                        <span
                          className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${
                            product.stock === 0
                              ? "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-400"
                              : product.stock <= 5
                                ? "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-400"
                                : "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400"
                          }`}
                        >
                          {product.stock === 0
                            ? t("status.outOfStock")
                            : product.stock <= 5
                              ? t("status.lowStock")
                              : t("status.inStock")}
                        </span>
                      </td>
                      <td className="py-3.5 pr-4 text-right">
                        <div className="inline-flex items-center gap-1.5">
                          <button
                            onClick={() => void adjustStock(variant, -1)}
                            disabled={savingVariantId === variant.id}
                            className="p-1.5 hover:bg-background rounded-lg text-foreground transition-colors disabled:opacity-50"
                            title={t("inventory.decrease")}
                          >
                            <Minus className="size-3.5" />
                          </button>
                          <input
                            type="number"
                            value={draftAdjustments[variant.id] ?? ""}
                            onChange={(event) =>
                              setDraftAdjustments((current) => ({
                                ...current,
                                [variant.id]: event.target.value,
                              }))
                            }
                            placeholder={t("inventory.deltaPlaceholder")}
                            className="w-20 text-center bg-secondary/40 border border-border rounded-lg px-2 py-1.5 font-bold text-xs text-foreground focus:outline-none"
                            aria-label={t("inventory.deltaPlaceholder")}
                          />
                          <button
                            onClick={() => handleDraftSubmit(variant)}
                            disabled={savingVariantId === variant.id}
                            className="px-2 py-1.5 text-[11px] font-semibold text-accent-brand hover:underline disabled:opacity-50"
                          >
                            {t("inventory.applyDelta")}
                          </button>
                          <button
                            onClick={() => void adjustStock(variant, 1)}
                            disabled={savingVariantId === variant.id}
                            className="p-1.5 hover:bg-background rounded-lg text-foreground transition-colors disabled:opacity-50"
                            title={t("inventory.increase")}
                          >
                            <Plus className="size-3.5" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                  {filtered.length === 0 && (
                    <tr>
                      <td colSpan={6} className="py-10 text-center text-muted-foreground">
                        {t("inventory.empty")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
