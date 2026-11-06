"use client";
import { revalidateCatalog } from "@/app/actions/revalidateCatalog";

import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback } from "react";
import { useRouter } from "@/shared/i18n/navigation";
import { Plus, Search, Edit2, Trash2, History, X } from "lucide-react";
import { catalogService } from "@/modules/catalog/api/catalogService";
import { adminService } from "@/modules/admin/api/adminService";
import type { AdminCategory, Product, ProductVariant } from "@/modules/catalog/types";
import type { InventoryLedger } from "@/modules/admin/types";
import { useTranslations } from "next-intl";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";
import { normalizeMediaUrl } from "@/shared/lib/media";

interface ProductView {
  id: string;
  variantId?: string;
  name: string;
  sku: string;
  category: string | { name?: string };
  brand: string | { name?: string };
  price: number;
  compareAtPrice: number | null;
  cost: number;
  stock: number;
  lowStockThreshold: number;
  status: string;
  image: string;
  variants: ProductVariant[];
  salesCount: number;
}

interface AuditLogEntry {
  id: string;
  productName: string;
  variantTitle: string;
  sku: string;
  change: string;
  quantity: number;
  reason: string;
  date: string;
  ledgerId: string;
}

export function AdminProducts() {
  const t = useTranslations("admin");
  const router = useRouter();

  const [products, setProducts] = useState<ProductView[]>([]);
  const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [categories, setCategories] = useState<AdminCategory[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("All");
  const [selectedStatus, setSelectedStatus] = useState("All");

  const [auditProduct, setAuditProduct] = useState<ProductView | null>(null);
  const [auditLogs, setAuditLogs] = useState<AuditLogEntry[]>([]);
  const [loadingAudit, setLoadingAudit] = useState(false);

  const { cursor, hasPrev, next, prev, reset, page } = useCursorPagination();

  const loadAll = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);

      const productQuery = new URLSearchParams({ pageSize: "20", keyset: "true", _t: String(Date.now()) });
      if (cursor) productQuery.set("cursor", cursor);
      if (debouncedSearch) productQuery.set("search", debouncedSearch);
      if (selectedCategory !== "All") productQuery.set("category", selectedCategory);
      if (selectedStatus !== "All") productQuery.set("status", selectedStatus);

      const [prodsRes, catsRes] = await Promise.all([
        catalogService.getAdminProducts(productQuery.toString()),
        catalogService.getAdminCategories(100),
      ]);

      const catList = (catsRes.data ?? []).filter(
        (category: AdminCategory & { deletedAt?: string }) => !category.deletedAt
      );
      setCategories(catList);

      const list = prodsRes.data ?? [];
      const formatted = list.map((p: Product, idx: number) => {
        let categoryName = "—";
        if (typeof p.category === "string") {
          categoryName = p.category;
        } else if (p.category && typeof p.category === "object" && p.category.name) {
          categoryName = p.category.name;
        } else if (Array.isArray(p.categories) && p.categories.length > 0) {
          categoryName = typeof p.categories[0] === "string" ? p.categories[0] : (p.categories[0]?.name || "—");
        }

        let brandName = "—";
        if (typeof p.brand === "string") {
          brandName = p.brand;
        } else if (p.brand && typeof p.brand === "object" && p.brand.name) {
          brandName = p.brand.name;
        }

        return {
          id: p.id || p.slug || `prod-item-${idx}`,
          variantId: p.variantId,
          name: p.title || p.product?.en || p.name || "—",
          sku: p.sku || "—",
          category: categoryName,
          brand: brandName,
          price: p.extractedPrice || (typeof p.price === "number" ? p.price : parseFloat(String(p.price || "0").replace(/[^0-9.-]+/g, "")) || 0),
          compareAtPrice: p.extractedOriginalPrice || (p.originalPrice ? parseFloat(String(p.originalPrice).replace(/[^0-9.-]+/g, "")) : null),
          cost: 0,
          stock: typeof p.stockQuantity === "number" ? p.stockQuantity : (p.stockCount || (p.stock !== undefined ? p.stock : 0)),
          lowStockThreshold: 10,
          status: p.status || "draft",
          image: normalizeMediaUrl(p.thumbnail || p.img || p.image),
          variants: p.variants || [],
          salesCount: 0,
        };
      });
      setProducts(formatted);
      if (prodsRes?.meta) setMeta(prodsRes.meta as CursorMeta);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [cursor, debouncedSearch, selectedCategory, selectedStatus]);

  useEffect(() => {
    loadAll();
  }, [loadAll]);

  // Debounce the search box
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      reset();
    }, 350);
    return () => clearTimeout(timer);
  }, [search, reset]);

  const handleViewAudit = async (prod: ProductView) => {
    setAuditProduct(prod);
    setLoadingAudit(true);
    try {
      const res = await adminService.getInventoryLedgers({
        pageSize: 50,
        productId: prod.id,
      });
      const list = res?.data ?? [];
      const mapped = list.map((l: InventoryLedger, idx: number) => ({
        id: l.id || `audit-${idx}`,
        productName: l.productName || prod.name,
        variantTitle: l.variantTitle || "Default Variant",
        sku: l.sku || prod.sku,
        change: l.quantity > 0 ? `+${l.quantity}` : `${l.quantity}`,
        quantity: l.quantity,
        reason: l.reason || "restock",
        date: l.createdAt ? new Date(l.createdAt).toLocaleString() : "Recently",
        ledgerId: l.id ? l.id.slice(0, 8) : "",
      }));
      setAuditLogs(mapped);
    } catch {
      notify.error(t("products.toast.auditFailed"));
    } finally {
      setLoadingAudit(false);
    }
  };

  const handleDeleteProduct = async (id: string) => {
    if (!confirm(t("products.confirmDelete"))) return;

    try {
      await catalogService.deleteProduct(id);
      setProducts((prev) => prev.filter((p) => p.id !== id));
      void revalidateCatalog().catch(() => false);
      notify.success(t("products.toast.deleted"));
    } catch (err) {
      notify.error(apiErrorMessage(err, t("products.toast.deleteFailed")));
    }
  };

  if (loadError) {
    return <ErrorState onRetry={loadAll} />;
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("products.title")}
        actions={
          <>
            <button
              onClick={() => router.push("/admin/products/drafts")}
              className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground transition-colors hover:bg-secondary"
            >
              {t("products.viewDrafts")}
            </button>
            <button
              onClick={() => router.push("/admin/products/new")}
              className="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
            >
              <Plus className="size-4" />
              {t("products.addNew")}
            </button>
          </>
        }
      />

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="relative w-full sm:max-w-xs">
          <Search className="absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder={t("products.searchPlaceholder")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-md border border-border bg-background py-2 pe-3 ps-9 text-sm text-foreground placeholder:text-muted-foreground focus:outline-hidden focus:ring-2 focus:ring-ring/30"
          />
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <select
            value={selectedCategory}
            onChange={(e) => {
              setSelectedCategory(e.target.value);
              reset();
            }}
            className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground focus:outline-hidden focus:ring-2 focus:ring-ring/30"
          >
            <option value="All">{t("products.allCategories")}</option>
            {categories.map((c, idx) => (
              <option key={c.id ? `cat-opt-${c.id}` : `cat-opt-${idx}`} value={c.name}>
                {c.name}
              </option>
            ))}
          </select>
          <select
            value={selectedStatus}
            onChange={(e) => {
              setSelectedStatus(e.target.value);
              reset();
            }}
            className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground focus:outline-hidden focus:ring-2 focus:ring-ring/30"
          >
            <option value="All">{t("status.all")}</option>
            <option value="draft">{t("status.draft")}</option>
            <option value="published">{t("status.published")}</option>
            <option value="archived">{t("status.archived")}</option>
          </select>
        </div>
      </div>

      {/* Products Table */}
      <div className="overflow-hidden rounded-lg border border-border bg-background">
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
            <thead className="text-xs text-muted-foreground">
              <tr className="border-b border-border">
                <th className="py-2.5 ps-4 text-start font-medium">{t("products.colProduct")}</th>
                <th className="py-2.5 text-start font-medium">{t("products.colSku")}</th>
                <th className="py-2.5 text-start font-medium">{t("products.colCategory")}</th>
                <th className="py-2.5 text-start font-medium">{t("products.colPrice")}</th>
                <th className="py-2.5 text-start font-medium">{t("products.colStock")}</th>
                <th className="py-2.5 text-start font-medium">{t("products.colStatus")}</th>
                <th className="py-2.5 pe-4 text-end font-medium">{t("products.colActions")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {products.length > 0 ? (
                products.map((prod, idx) => (
                  <tr key={prod.id ? `prod-row-${prod.id}-${idx}` : `prod-row-${idx}`} className="hover:bg-secondary/20 transition-colors">
                    <td className="py-3 ps-4">
                      <div className="flex items-center gap-3">
                        <img
                          src={normalizeMediaUrl(prod.image)}
                          alt=""
                          className="size-10 shrink-0 rounded-md border border-border object-cover"
                        />
                        <div className="min-w-0">
                          <div className="truncate text-sm font-medium text-foreground">
                            {prod.name}
                          </div>
                          <div className="truncate text-xs text-muted-foreground">
                            {t("products.brandLabel", {
                              name:
                                typeof prod.brand === "object"
                                  ? prod.brand?.name || "—"
                                  : String(prod.brand || "—"),
                            })}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="py-3 font-mono text-xs text-muted-foreground">{prod.sku}</td>
                    <td className="py-3 text-foreground">
                      {(typeof prod.category === "object"
                        ? prod.category?.name || "—"
                        : String(prod.category || "—"))}
                    </td>
                    <td className="py-3">
                      <div className="font-medium text-foreground tabular-nums">EGP {prod.price.toFixed(2)}</div>
                      {prod.compareAtPrice ? (
                        <div className="text-xs text-muted-foreground line-through tabular-nums">
                          EGP {prod.compareAtPrice.toFixed(2)}
                        </div>
                      ) : null}
                    </td>
                    <td className="py-3">
                      <span className={`font-medium tabular-nums ${prod.stock <= 10 ? "text-amber-700 dark:text-amber-400" : "text-foreground"}`}>
                        {t("products.units", { count: prod.stock })}
                      </span>
                      {prod.variants && prod.variants.length > 1 ? (
                        <span className="block text-xs text-muted-foreground">
                          {t("products.variantCount", { count: prod.variants.length })}
                        </span>
                      ) : null}
                    </td>
                    <td className="py-3">
                      <StatusPill tone={statusTone(prod.status)}>
                        {t(`status.${String(prod.status).toLowerCase()}`)}
                      </StatusPill>
                    </td>
                    <td className="py-3 pe-4 text-end">
                      <div className="flex items-center justify-end gap-1.5">
                        <button
                          onClick={() => handleViewAudit(prod)}
                          className="p-1.5 text-muted-foreground hover:text-accent-brand hover:bg-secondary rounded-lg transition-colors"
                          title={t("products.viewAudit")}
                        >
                          <History className="size-4" />
                        </button>
                        <button
                          onClick={() => router.push(`/admin/products/${prod.id}`)}
                          className="p-1.5 text-muted-foreground hover:text-accent-brand hover:bg-secondary rounded-lg transition-colors"
                          title={t("products.editProduct")}
                        >
                          <Edit2 className="size-4" />
                        </button>
                        <button
                          onClick={() => handleDeleteProduct(prod.id)}
                          className="p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
                          title={t("products.deleteProduct")}
                        >
                          <Trash2 className="size-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={7} className="text-center py-8 text-muted-foreground">
                    {t("products.empty")}
                  </td>
                </tr>
              )}
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
          shown={products.length}
        />
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
                  <h3 className="font-bold text-base text-foreground">{t("products.audit.title")}</h3>
                  <p className="text-xs text-muted-foreground">
                    {auditProduct.name} &bull; <span className="font-mono">{auditProduct.sku}</span>
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="text-end">
                  <div className="text-[10px] text-muted-foreground uppercase font-semibold">
                    {t("products.audit.currentStock")}
                  </div>
                  <div className="text-xs font-bold text-accent-brand">
                    {t("products.units", { count: auditProduct.stock })}
                  </div>
                </div>
                <button
                  onClick={() => setAuditProduct(null)}
                  className="text-muted-foreground hover:text-foreground text-sm font-semibold p-1"
                >
                  <X className="size-4" />
                </button>
              </div>
            </div>

            {auditProduct.variants && auditProduct.variants.length > 1 && (
              <div className="flex flex-wrap gap-1.5 p-2 rounded-xl bg-secondary/20 border border-border">
                {auditProduct.variants.map((v: ProductVariant, vIdx: number) => (
                  <span
                    key={v.id || vIdx}
                    className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-lg bg-background border border-border text-[11px]"
                  >
                    <span className="font-semibold text-foreground">{v.title}</span>
                    <span className="font-bold text-accent-brand">
                      {v.stock?.availableQuantity ?? 0} {t("products.colStock").toLowerCase()}
                    </span>
                  </span>
                ))}
              </div>
            )}

            <div className="flex-1 overflow-y-auto space-y-2 pe-1 min-h-[220px]">
              {loadingAudit ? (
                <div className="py-12 text-center text-xs text-muted-foreground">{t("products.audit.loading")}</div>
              ) : auditLogs.length === 0 ? (
                <div className="py-12 text-center text-xs text-muted-foreground">
                  {t("products.audit.empty")}
                </div>
              ) : (
                auditLogs.map((entry, idx) => (
                  <div key={entry.id ? `audit-${entry.id}-${idx}` : `audit-${idx}`} className="p-3 border border-border rounded-xl bg-secondary/15 flex items-center justify-between text-xs">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-foreground">{entry.variantTitle}</span>
                        <span className="px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wider bg-secondary text-secondary-foreground border border-border">
                          {entry.reason}
                        </span>
                      </div>
                      <div className="text-[11px] text-muted-foreground mt-0.5 font-mono">
                        SKU: {entry.sku} &bull; {entry.date} &bull; Ref: {entry.ledgerId}
                      </div>
                    </div>
                    <div className="text-end">
                      <span className={`text-sm font-bold ${entry.quantity > 0 ? "text-emerald-600 dark:text-emerald-400" : "text-amber-600 dark:text-amber-400"}`}>
                        {t("products.units", { count: entry.change })}
                      </span>
                    </div>
                  </div>
                ))
              )}
            </div>

            <div className="flex items-center justify-between pt-3 border-t border-border">
              <div className="text-xs text-muted-foreground">
                {t("products.audit.totalBalance")}{" "}
                <span className="font-bold text-foreground">
                  {t("products.units", { count: auditLogs.reduce((acc, curr) => acc + (curr.quantity || 0), 0) })}
                </span>
              </div>
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
    </div>
  );
}
