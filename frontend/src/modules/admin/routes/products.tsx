"use client";
import { revalidateCatalog } from "@/app/actions/revalidateCatalog";

import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback } from "react";
import { useRouter } from "@/shared/i18n/navigation";
import {
  Plus,
  Search,
  Filter,
  Edit2,
  Trash2,
  History,
  X,
  FileText,
} from "lucide-react";
import { catalogService } from "@/modules/catalog/api/catalogService";
import { adminService } from "@/modules/admin/api/adminService";
import { useTranslations } from "next-intl";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";
import { normalizeMediaUrl } from "@/shared/lib/media";

export function AdminProducts() {
  const t = useTranslations("admin");
  const router = useRouter();

  const [products, setProducts] = useState<any[]>([]);
  const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [categories, setCategories] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("All");
  const [selectedStatus, setSelectedStatus] = useState("All");

  const [auditProduct, setAuditProduct] = useState<any>(null);
  const [auditLogs, setAuditLogs] = useState<any[]>([]);
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

      const [prodsRes, catsRes]: [any, any] = await Promise.all([
        catalogService.getAdminProducts(productQuery.toString()),
        catalogService.getAdminCategories(100),
      ]);

      const catList = (Array.isArray(catsRes) ? catsRes : (Array.isArray(catsRes?.data) ? catsRes.data : []))
        .filter((category: any) => !category.deletedAt);
      setCategories(catList);

      const list = Array.isArray(prodsRes) ? prodsRes : (Array.isArray(prodsRes?.data) ? prodsRes.data : []);
      const formatted = list.map((p: any, idx: number) => {
        let categoryName = "Electronics";
        if (typeof p.category === "string") {
          categoryName = p.category;
        } else if (p.category && typeof p.category === "object" && p.category.name) {
          categoryName = p.category.name;
        } else if (Array.isArray(p.categories) && p.categories.length > 0) {
          categoryName = typeof p.categories[0] === "string" ? p.categories[0] : (p.categories[0]?.name || "Electronics");
        }

        let brandName = "PRIM";
        if (typeof p.brand === "string") {
          brandName = p.brand;
        } else if (p.brand && typeof p.brand === "object" && p.brand.name) {
          brandName = p.brand.name;
        }

        return {
          id: p.id || p.slug || `prod-item-${idx}`,
          variantId: p.variantId,
          name: p.title || p.product?.en || p.name || "Product",
          sku: p.sku || `SKU-${(p.id || "001").slice(0, 8)}`,
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
      if (prodsRes?.meta) setMeta(prodsRes.meta);
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

  const handleViewAudit = async (prod: any) => {
    setAuditProduct(prod);
    setLoadingAudit(true);
    try {
      const res = await adminService.getInventoryLedgers({
        pageSize: 50,
        productId: prod.id,
      });
      const list = Array.isArray(res?.data) ? res.data : (Array.isArray(res) ? res : []);
      const mapped = list.map((l: any, idx: number) => ({
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
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.toast.deleteFailed")));
    }
  };

  if (loadError) {
    return <ErrorState onRetry={loadAll} />;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("products.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t("products.subtitle")}
          </p>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <button
            onClick={() => router.push("/admin/products/drafts")}
            className="flex items-center gap-2 px-3.5 py-2.5 bg-background border border-border hover:bg-secondary text-foreground font-semibold text-xs rounded-xl shadow-xs transition-all"
            title={t("products.draftsTitle")}
          >
            <FileText className="size-4 text-accent-brand" />
            <span>{t("products.viewDrafts")}</span>
          </button>
          <button
            onClick={() => router.push("/admin/products/new")}
            className="flex items-center gap-2 px-4 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all"
          >
            <Plus className="size-4" />
            <span>{t("products.addNew")}</span>
          </button>
        </div>
      </div>

      {/* Filters Bar */}
      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs flex flex-col md:flex-row gap-4 justify-between items-center">
        <div className="relative w-full md:w-80">
          <Search className="absolute start-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder={t("products.searchPlaceholder")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full ps-10 pe-4 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground placeholder:text-muted-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
          />
        </div>

        <div className="flex items-center gap-3 w-full md:w-auto">
          <Filter className="size-4 text-muted-foreground" />
          <select
            value={selectedCategory}
            onChange={(e) => {
              setSelectedCategory(e.target.value);
              reset();
            }}
            className="px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
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
            className="px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
          >
            <option value="All">{t("status.all")}</option>
            <option value="draft">{t("status.draft")}</option>
            <option value="published">{t("status.published")}</option>
            <option value="archived">{t("status.archived")}</option>
          </select>
        </div>
      </div>

      {/* Products Table */}
      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full sm:min-w-max text-start text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 ps-4">{t("products.colProduct")}</th>
                <th className="py-3.5">{t("products.colSku")}</th>
                <th className="py-3.5">{t("products.colCategory")}</th>
                <th className="py-3.5">{t("products.colPrice")}</th>
                <th className="py-3.5">{t("products.colStock")}</th>
                <th className="py-3.5">{t("products.colStatus")}</th>
                <th className="py-3.5 pe-4 text-end">{t("products.colActions")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {products.length > 0 ? (
                products.map((prod, idx) => (
                  <tr key={prod.id ? `prod-row-${prod.id}-${idx}` : `prod-row-${idx}`} className="hover:bg-secondary/20 transition-colors">
                    <td className="py-3.5 ps-4">
                      <div className="flex items-center gap-3">
                        <img
                          src={normalizeMediaUrl(prod.image)}
                          alt={prod.name}
                          className="size-11 rounded-xl object-cover border border-border"
                        />
                        <div>
                          <div className="font-semibold text-foreground text-xs line-clamp-1">
                            {prod.name}
                          </div>
                          <div className="text-[10px] text-muted-foreground">
                            {t("products.brandLabel", {
                              name: typeof prod.brand === "object" ? prod.brand?.name : String(prod.brand || "PRIM"),
                            })}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="py-3.5 font-mono text-muted-foreground font-medium">{prod.sku}</td>
                    <td className="py-3.5 font-medium text-foreground">
                      {typeof prod.category === "object" ? prod.category?.name : String(prod.category || "General")}
                    </td>
                    <td className="py-3.5">
                      <div className="font-bold text-foreground">EGP {prod.price.toFixed(2)}</div>
                      {prod.compareAtPrice && (
                        <div className="text-[10px] text-muted-foreground line-through">
                          EGP {prod.compareAtPrice.toFixed(2)}
                        </div>
                      )}
                    </td>
                    <td className="py-3.5">
                      <span className={`font-bold ${prod.stock <= 10 ? "text-amber-600 dark:text-amber-400" : "text-foreground"}`}>
                        {t("products.units", { count: prod.stock })}
                      </span>
                      {prod.variants && prod.variants.length > 1 && (
                        <span className="text-[10px] text-muted-foreground block font-medium">
                          {prod.variants.length} variants
                        </span>
                      )}
                    </td>
                    <td className="py-3.5 text-foreground">
                      {t(`status.${String(prod.status).toLowerCase()}`) || prod.status}
                    </td>
                    <td className="py-3.5 pe-4 text-end">
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
                {auditProduct.variants.map((v: any, vIdx: number) => (
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
