"use client";

import { useEffect, useState } from "react";
import { Search, Plus, X, Loader2 } from "lucide-react";
import { sellerService } from "@/modules/seller/api/sellerService";
import type {
  SellerProduct,
  SellerVariant,
} from "@/modules/seller/types";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { apiErrorMessage, notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";
import { useRouter } from "@/shared/i18n/navigation";


function parseMajorPrice(value: string): number | undefined {
  const parsed = Number(value);
  if (value.trim() === "" || !Number.isFinite(parsed) || parsed <= 0) return undefined;
  return Math.round(parsed * 100);
}

interface VariantForm {
  title: string;
  sku: string;
  price: string;
  crossedOutPrice: string;
  currency: string;
  isDefault: boolean;
  attributes: string;
}

const EMPTY_VARIANT_FORM: VariantForm = {
  title: "",
  sku: "",
  price: "",
  crossedOutPrice: "",
  currency: "EGP",
  isDefault: true,
  attributes: "",
};

export function SellerProducts() {
  const t = useTranslations("seller");
  const router = useRouter();

  const [products, setProducts] = useState<SellerProduct[]>([]);
  const [meta, setMeta] = useState<{ page: number; totalPages: number; totalItems: number } | null>(null);
  const [page, setPage] = useState(1);
  const [variants, setVariants] = useState<SellerVariant[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<Error | null>(null);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");

  const [lifecycleId, setLifecycleId] = useState<string | null>(null);

  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [bulkBusy, setBulkBusy] = useState(false);

  const [selectedProduct, setSelectedProduct] = useState<SellerProduct | null>(null);
  const [editingVariant, setEditingVariant] = useState<SellerVariant | null>(null);
  const [variantForm, setVariantForm] = useState<VariantForm>(EMPTY_VARIANT_FORM);
  const [savingVariant, setSavingVariant] = useState(false);
  const [stockAmount, setStockAmount] = useState("10");
  const [stockReason, setStockReason] = useState<"restock" | "return" | "adjustment">("restock");
  const [savingStock, setSavingStock] = useState(false);

  const loadProducts = async (nextPage = page) => {
    setLoading(true);
    setLoadError(null);
    try {
      const productResponse = await sellerService.getProducts({ page: nextPage, pageSize: 20 });
      setProducts(productResponse.data);
      const nextMeta = productResponse.meta as { page?: number; totalPages?: number; totalItems?: number } | undefined;
      if (nextMeta && typeof nextMeta.page === "number") {
        setMeta({ page: nextMeta.page, totalPages: nextMeta.totalPages ?? 1, totalItems: nextMeta.totalItems ?? productResponse.data.length });
      } else {
        setMeta({ page: nextPage, totalPages: 1, totalItems: productResponse.data.length });
      }
    } catch (error) {
      setLoadError(error instanceof Error ? error : new Error(String(error)));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadProducts();
  }, []);

  const openVariants = async (product: SellerProduct) => {
    setSelectedProduct(product);
    setEditingVariant(null);
    setVariantForm(EMPTY_VARIANT_FORM);
    try {
      const response = await sellerService.getVariants(product.id, { pageSize: 100 });
      setVariants(response.data);
    } catch (error) {
      notify.error(apiErrorMessage(error, t("variants.loadFailed")));
    }
  };

  const closeVariants = () => {
    setSelectedProduct(null);
    setVariants([]);
  };

  const handleOpenAdd = () => {
    router.push("/seller/products/new");
  };

  const handleOpenEdit = (product: SellerProduct) => {
    router.push(`/seller/products/${product.id}/edit`);
  };

  const handleLifecycle = async (product: SellerProduct, action: "publish" | "archive" | "delete") => {
    const confirmed = action === "delete"
      ? window.confirm(t("products.confirmDelete"))
      : true;
    if (!confirmed) return;

    setLifecycleId(product.id);
    try {
      if (action === "publish") await sellerService.publishProduct(product.id);
      if (action === "archive") await sellerService.archiveProduct(product.id);
      if (action === "delete") await sellerService.deleteProduct(product.id);
      notify.success(
        action === "publish"
          ? t("toasts.productPublished")
          : action === "archive"
            ? t("toasts.productArchived")
            : t("toasts.productRemoved"),
      );
      await loadProducts();
    } catch (error) {
      notify.error(apiErrorMessage(error, t("products.lifecycleFailed")));
    } finally {
      setLifecycleId(null);
    }
  };

  const handleOpenVariantCreate = () => {
    setEditingVariant(null);
    setVariantForm(EMPTY_VARIANT_FORM);
  };

  const handleOpenVariantEdit = (variant: SellerVariant) => {
    setEditingVariant(variant);
    setVariantForm({
      title: variant.title,
      sku: variant.sku,
      price: variant.price ? String(variant.price / 100) : "",
      crossedOutPrice: variant.crossedOutPrice ? String(variant.crossedOutPrice / 100) : "",
      currency: variant.currency ?? "EGP",
      isDefault: variant.isDefault,
      attributes: variant.attributes
        ? Object.entries(variant.attributes).map(([key, value]) => `${key}=${String(value)}`).join(", ")
        : "",
    });
  };

  const handleSubmitVariant = async (event: React.FormEvent) => {
    if (!selectedProduct || !variantForm.title.trim()) return;
    event.preventDefault();

    const attributes = Object.fromEntries(
      variantForm.attributes
        .split(",")
        .map((part) => part.trim())
        .filter(Boolean)
        .map((part) => {
          const separator = part.indexOf("=");
          if (separator <= 0) return null;
          return [part.slice(0, separator).trim(), part.slice(separator + 1).trim()];
        })
        .filter((entry): entry is [string, string] => Boolean(entry)),
    );

    setSavingVariant(true);
    try {
      const payload = {
        title: variantForm.title.trim(),
        sku: variantForm.sku.trim() || undefined,
        price: variantForm.price ? parseMajorPrice(variantForm.price) : undefined,
        crossedOutPrice: variantForm.crossedOutPrice ? parseMajorPrice(variantForm.crossedOutPrice) : undefined,
        currency: variantForm.currency.trim() || "EGP",
        attributes,
        isDefault: variantForm.isDefault,
      };

      if (editingVariant) {
        await sellerService.updateVariant(editingVariant.id, payload);
        notify.success(t("toasts.variantUpdated"));
      } else {
        await sellerService.createVariant(selectedProduct.id, payload);
        notify.success(t("toasts.variantAdded"));
      }
      const response = await sellerService.getVariants(selectedProduct.id, { pageSize: 100 });
      setVariants(response.data);
      setEditingVariant(null);
      setVariantForm(EMPTY_VARIANT_FORM);
    } catch (error) {
      notify.error(apiErrorMessage(error, t("variants.saveFailed")));
    } finally {
      setSavingVariant(false);
    }
  };

  const handleDeleteVariant = async (variant: SellerVariant) => {
    if (!selectedProduct || !window.confirm(t("variants.confirmDelete"))) return;
    try {
      await sellerService.deleteVariant(variant.id);
      const response = await sellerService.getVariants(selectedProduct.id, { pageSize: 100 });
      setVariants(response.data);
      notify.success(t("toasts.variantRemoved"));
    } catch (error) {
      notify.error(apiErrorMessage(error, t("variants.deleteFailed")));
    }
  };

  const handleAdjustStock = async (variant: SellerVariant) => {
    const quantity = Number.parseInt(stockAmount, 10);
    if (!Number.isFinite(quantity) || quantity === 0) {
      notify.error(t("variants.invalidStockAmount"));
      return;
    }
    setSavingStock(true);
    try {
      await sellerService.adjustStock(variant.id, { quantity, reason: stockReason });
      notify.success(t("toasts.inventoryUpdated"));
      if (selectedProduct) {
        const response = await sellerService.getVariants(selectedProduct.id, { pageSize: 100 });
        setVariants(response.data);
      }
      await loadProducts();
    } catch (error) {
      notify.error(apiErrorMessage(error, t("variants.stockFailed")));
    } finally {
      setSavingStock(false);
    }
  };

  const filtered = products.filter((product) => {
    const matchesSearch = `${product.title} ${product.sku ?? ""} ${product.slug}`.toLowerCase().includes(search.toLowerCase());
    const matchesStatus = statusFilter === "all" || product.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  useEffect(() => {
    setSelectedIds((current) => {
      const valid = new Set(products.map((product) => product.id));
      const next = new Set([...current].filter((id) => valid.has(id)));
      return next.size === current.size ? current : next;
    });
  }, [products]);

  const toggleSelect = (id: string) => {
    setSelectedIds((current) => {
      const next = new Set(current);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const toggleSelectAll = () => {
    setSelectedIds((current) =>
      current.size === filtered.length ? new Set<string>() : new Set(filtered.map((product) => product.id)),
    );
  };

  const handleBulk = async (action: "publish" | "archive") => {
    if (selectedIds.size === 0) return;
    const confirmed = window.confirm(t("products.bulkConfirm", { count: selectedIds.size }));
    if (!confirmed) return;
    setBulkBusy(true);
    const results = await Promise.allSettled(
      [...selectedIds].map((id) =>
        action === "publish" ? sellerService.publishProduct(id) : sellerService.archiveProduct(id),
      ),
    );
    const failed = results.filter((r) => r.status === "rejected").length;
    const succeeded = results.length - failed;
    if (succeeded > 0) {
      notify.success(
        action === "publish"
          ? t("products.bulkPublished", { count: succeeded })
          : t("products.bulkArchived", { count: succeeded }),
      );
    }
    if (failed > 0) notify.error(t("products.bulkPartial", { failed }));
    setSelectedIds(new Set());
    setBulkBusy(false);
    await loadProducts();
  };

  const statusBadge = (product: SellerProduct) => (
    <StatusPill tone={statusTone(product.status)}>
      {t(`products.status.${product.status}`)}
    </StatusPill>
  );

  return (
    <div className="space-y-5 w-full">
      <PageHeader
        title={t("products.title")}
        description={t("products.subtitle")}
        actions={
          <button
            onClick={handleOpenAdd}
            className="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
          >
            <Plus className="size-4" />
            {t("actions.addNewProduct")}
          </button>
        }
      />

      {loading ? (
        <div className="flex min-h-[40vh] items-center justify-center">
          <div className="size-6 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent" />
        </div>
      ) : loadError ? (
        <ErrorState
          title={t("products.loadTitle")}
          description={loadError.message || t("products.loadDescription")}
          onRetry={() => void loadProducts()}
        />
      ) : (
        <>
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="relative w-full sm:max-w-xs">
              <Search className="absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                placeholder={t("products.searchPlaceholder")}
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                className="w-full rounded-md border border-border bg-background py-2 pe-3 ps-9 text-sm text-foreground placeholder:text-muted-foreground focus:outline-hidden focus:ring-2 focus:ring-ring/30"
              />
            </div>
            <select
              value={statusFilter}
              onChange={(event) => setStatusFilter(event.target.value)}
              className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
              aria-label={t("products.table.status")}
            >
              <option value="all">{t("products.filterAll")}</option>
              <option value="draft">{t("products.status.draft")}</option>
              <option value="published">{t("products.status.published")}</option>
              <option value="archived">{t("products.status.archived")}</option>
            </select>
          </div>

          <div className="overflow-hidden rounded-lg border border-border bg-background">
            {selectedIds.size > 0 ? (
              <div className="flex flex-wrap items-center gap-2 border-b border-border bg-secondary/40 px-4 py-2.5 text-sm">
                <span className="font-medium text-foreground">
                  {t("products.bulkSelected", { count: selectedIds.size })}
                </span>
                <button
                  onClick={() => void handleBulk("publish")}
                  disabled={bulkBusy}
                  className="rounded-md border border-border bg-background px-2.5 py-1.5 text-xs font-medium text-foreground disabled:opacity-50"
                >
                  {t("actions.publish")}
                </button>
                <button
                  onClick={() => void handleBulk("archive")}
                  disabled={bulkBusy}
                  className="rounded-md border border-border bg-background px-2.5 py-1.5 text-xs font-medium text-foreground disabled:opacity-50"
                >
                  {t("actions.archive")}
                </button>
                <button
                  onClick={() => setSelectedIds(new Set())}
                  disabled={bulkBusy}
                  className="rounded-md px-2.5 py-1.5 text-xs font-medium text-muted-foreground hover:bg-background disabled:opacity-50"
                >
                  {t("actions.clear")}
                </button>
              </div>
            ) : null}
            <div className="overflow-x-auto">
              <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
                <thead className="text-xs text-muted-foreground">
                  <tr className="border-b border-border">
                    <th className="py-2.5 ps-4 w-8">
                      <input
                        type="checkbox"
                        checked={filtered.length > 0 && selectedIds.size === filtered.length}
                        onChange={toggleSelectAll}
                        aria-label={t("products.bulkSelectAll")}
                      />
                    </th>
                    <th className="py-2.5 text-start font-medium">{t("products.table.details")}</th>
                    <th className="py-2.5 text-start font-medium">{t("products.table.sku")}</th>
                    <th className="py-2.5 text-start font-medium">{t("products.table.price")}</th>
                    <th className="py-2.5 text-start font-medium">{t("products.table.stock")}</th>
                    <th className="py-2.5 text-start font-medium">{t("products.table.status")}</th>
                    <th className="py-2.5 pe-4 text-end font-medium">{t("products.table.actions")}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/60">
                  {filtered.map((product) => (
                    <tr key={product.id} className="transition-colors hover:bg-secondary/30">
                      <td className="py-3 ps-4">
                        <input
                          type="checkbox"
                          checked={selectedIds.has(product.id)}
                          onChange={() => toggleSelect(product.id)}
                          aria-label={product.title}
                        />
                      </td>
                      <td className="py-3">
                        <div className="flex items-center gap-3">
                          {product.thumbnailUrl ? (
                            <img src={product.thumbnailUrl} alt="" className="size-10 shrink-0 rounded-md border border-border object-cover" />
                          ) : (
                            <div className="size-10 shrink-0 rounded-md border border-border bg-secondary/40" />
                          )}
                          <div className="min-w-0">
                            <div className="truncate text-sm font-medium text-foreground">{product.title}</div>
                            <div className="truncate text-xs text-muted-foreground">{product.slug}</div>
                          </div>
                        </div>
                      </td>
                      <td className="py-3 font-mono text-xs text-muted-foreground">{product.sku ?? "—"}</td>
                      <td className="py-3 font-medium text-foreground tabular-nums">
                        {product.price ? `${product.currency ?? "EGP"} ${(product.price / 100).toFixed(2)}` : "—"}
                      </td>
                      <td className="py-3 font-medium text-foreground tabular-nums">
                        {t("products.units", { count: product.stock })}
                      </td>
                      <td className="py-3">{statusBadge(product)}</td>
                      <td className="py-3 pe-4 text-end">
                        <div className="inline-flex items-center gap-1">
                          <button
                            onClick={() => void openVariants(product)}
                            className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                          >
                            {t("variants.manage")}
                          </button>
                          <button
                            onClick={() => handleOpenEdit(product)}
                            className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                          >
                            {t("actions.edit")}
                          </button>
                          {product.status !== "published" ? (
                            <button
                              onClick={() => void handleLifecycle(product, "publish")}
                              disabled={lifecycleId === product.id}
                              className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground disabled:opacity-50"
                            >
                              {t("actions.publish")}
                            </button>
                          ) : (
                            <button
                              onClick={() => void handleLifecycle(product, "archive")}
                              disabled={lifecycleId === product.id}
                              className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground disabled:opacity-50"
                            >
                              {t("actions.archive")}
                            </button>
                          )}
                          <button
                            onClick={() => void handleLifecycle(product, "delete")}
                            disabled={lifecycleId === product.id}
                            className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive disabled:opacity-50"
                          >
                            {t("actions.delete")}
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                  {filtered.length === 0 ? (
                    <tr>
                      <td colSpan={7} className="py-10 text-center text-sm text-muted-foreground">
                        <p>{t("products.empty")}</p>
                        {products.length === 0 ? (
                          <button
                            onClick={handleOpenAdd}
                            className="mt-3 inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-2 text-xs font-medium text-primary-foreground"
                          >
                            <Plus className="size-3.5" />
                            {t("actions.addNewProduct")}
                          </button>
                        ) : null}
                      </td>
                    </tr>
                  ) : null}
                </tbody>
              </ResponsiveTable>
            </div>
            {meta && meta.totalPages > 1 ? (
              <div className="flex items-center justify-between border-t border-border px-4 py-3 text-sm">
                <span className="text-xs text-muted-foreground">
                  {t("orders.pagination", { page: meta.page, total: meta.totalPages, count: meta.totalItems })}
                </span>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => { const next = Math.max(1, page - 1); setPage(next); void loadProducts(next); }}
                    disabled={page <= 1 || loading}
                    className="rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-foreground disabled:opacity-50"
                  >
                    {t("actions.previous")}
                  </button>
                  <button
                    onClick={() => { const next = page + 1; setPage(next); void loadProducts(next); }}
                    disabled={page >= meta.totalPages || loading}
                    className="rounded-md border border-border px-2.5 py-1.5 text-xs font-medium text-foreground disabled:opacity-50"
                  >
                    {t("actions.next")}
                  </button>
                </div>
              </div>
            ) : null}
          </div>
        </>
      )}

      {selectedProduct && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="max-h-[92vh] w-full max-w-3xl space-y-4 overflow-y-auto rounded-lg border border-border bg-background p-5 shadow-xl">
            <div className="flex items-start justify-between border-b border-border pb-3">
              <div>
                <h3 className="text-sm font-semibold text-foreground">{t("variants.title")}</h3>
                <p className="mt-0.5 text-xs text-muted-foreground">{selectedProduct.title}</p>
              </div>
              <button
                onClick={closeVariants}
                className="text-muted-foreground hover:text-foreground"
                aria-label={t("actions.cancel")}
              >
                <X className="size-4" />
              </button>
            </div>

            <div className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-border bg-secondary/40 px-3 py-2 text-xs">
              <span className="font-medium text-foreground">{t("variants.sellable")}</span>
              <span className="text-muted-foreground">{t("variants.sellableNote")}</span>
            </div>

            <div className="overflow-x-auto rounded-md border border-border">
              <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
                <thead className="text-xs text-muted-foreground">
                  <tr className="border-b border-border">
                    <th className="py-2.5 ps-3 text-start font-medium">{t("variants.table.title")}</th>
                    <th className="py-2.5 text-start font-medium">{t("variants.table.sku")}</th>
                    <th className="py-2.5 text-start font-medium">{t("variants.table.price")}</th>
                    <th className="py-2.5 text-start font-medium">{t("variants.table.default")}</th>
                    <th className="py-2.5 text-start font-medium">{t("variants.table.stockAdjustment")}</th>
                    <th className="py-2.5 pe-3 text-end font-medium">{t("variants.table.actions")}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {variants.map((variant) => (
                    <tr key={variant.id} className="transition-colors hover:bg-secondary/30">
                      <td className="py-3 ps-3">
                        <span className="font-medium text-foreground">{variant.title}</span>
                        {variant.isDefault ? (
                          <span className="ms-2 align-middle">
                            <StatusPill>{t("variants.defaultYes")}</StatusPill>
                          </span>
                        ) : null}
                      </td>
                      <td className="py-3 font-mono text-xs text-muted-foreground">{variant.sku}</td>
                      <td className="py-3 font-medium text-foreground tabular-nums">
                        {variant.price ? `${variant.currency ?? "EGP"} ${(variant.price / 100).toFixed(2)}` : "—"}
                      </td>
                      <td className="py-3">
                        <StatusPill tone={variant.isDefault ? "info" : "neutral"}>
                          {variant.isDefault ? t("variants.defaultYes") : t("variants.defaultNo")}
                        </StatusPill>
                      </td>
                      <td className="py-3">
                        <div className="flex flex-wrap items-center gap-1.5">
                          <input
                            type="number"
                            value={stockAmount}
                            onChange={(event) => setStockAmount(event.target.value)}
                            className="w-20 rounded-md border border-border bg-background px-2 py-1.5 text-xs text-foreground"
                            aria-label={t("variants.stockAmount")}
                          />
                          <select
                            value={stockReason}
                            onChange={(event) => setStockReason(event.target.value as "restock" | "return" | "adjustment")}
                            className="rounded-md border border-border bg-background px-2 py-1.5 text-xs text-foreground"
                            aria-label={t("variants.stockReason")}
                          >
                            <option value="restock">{t("variants.reasonRestock")}</option>
                            <option value="return">{t("variants.reasonReturn")}</option>
                            <option value="adjustment">{t("variants.reasonAdjustment")}</option>
                          </select>
                          <button
                            onClick={() => void handleAdjustStock(variant)}
                            disabled={savingStock}
                            className="rounded-md border border-border px-2 py-1.5 text-xs font-medium text-foreground transition-colors hover:bg-secondary disabled:opacity-50"
                          >
                            {t("variants.applyStock")}
                          </button>
                        </div>
                      </td>
                      <td className="py-3 pe-3 text-end">
                        <div className="inline-flex items-center gap-1">
                          <button
                            onClick={() => handleOpenVariantEdit(variant)}
                            className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                          >
                            {t("actions.edit")}
                          </button>
                          <button
                            onClick={() => void handleDeleteVariant(variant)}
                            className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                          >
                            {t("actions.delete")}
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                  {variants.length === 0 ? (
                    <tr>
                      <td colSpan={7} className="py-8 text-center text-sm text-muted-foreground">
                        {t("variants.empty")}
                      </td>
                    </tr>
                  ) : null}
                </tbody>
              </ResponsiveTable>
            </div>

            <form onSubmit={handleSubmitVariant} className="space-y-4 rounded-md border border-border p-4">
              <div className="flex items-center justify-between">
                <h4 className="text-sm font-semibold text-foreground">
                  {editingVariant ? t("variants.edit") : t("variants.add")}
                </h4>
                <button
                  type="button"
                  onClick={handleOpenVariantCreate}
                  className="inline-flex items-center gap-1 rounded-md border border-border px-2 py-1 text-xs font-medium text-foreground hover:bg-secondary"
                >
                  <Plus className="size-3.5" /> {t("variants.add")}
                </button>
              </div>
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="space-y-1">
                  <label className="text-sm font-medium text-foreground">{t("variants.form.title")}</label>
                  <input
                    required
                    value={variantForm.title}
                    onChange={(event) => setVariantForm({ ...variantForm, title: event.target.value })}
                    className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-sm font-medium text-foreground">{t("variants.form.sku")}</label>
                  <input
                    value={variantForm.sku}
                    onChange={(event) => setVariantForm({ ...variantForm, sku: event.target.value })}
                    className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                  />
                </div>
              </div>
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                <div className="space-y-1">
                  <label className="text-sm font-medium text-foreground">{t("variants.form.price")}</label>
                  <input
                    type="number"
                    min="0.01"
                    step="0.01"
                    value={variantForm.price}
                    onChange={(event) => setVariantForm({ ...variantForm, price: event.target.value })}
                    className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-sm font-medium text-foreground">{t("variants.form.compareAtPrice")}</label>
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    value={variantForm.crossedOutPrice}
                    onChange={(event) => setVariantForm({ ...variantForm, crossedOutPrice: event.target.value })}
                    className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-sm font-medium text-foreground">{t("variants.form.currency")}</label>
                  <input
                    value={variantForm.currency}
                    onChange={(event) => setVariantForm({ ...variantForm, currency: event.target.value.toUpperCase() })}
                    className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                  />
                </div>
              </div>
              <div className="space-y-1">
                <label className="text-sm font-medium text-foreground">{t("variants.form.attributes")}</label>
                <input
                  value={variantForm.attributes}
                  onChange={(event) => setVariantForm({ ...variantForm, attributes: event.target.value })}
                  placeholder={t("variants.form.attributesPlaceholder")}
                  className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                />
              </div>
              <label className="flex items-center gap-2 text-sm font-medium text-foreground">
                <input
                  type="checkbox"
                  checked={variantForm.isDefault}
                  onChange={(event) => setVariantForm({ ...variantForm, isDefault: event.target.checked })}
                />
                {t("variants.form.isDefault")}
              </label>
              <div className="flex justify-end">
                <button
                  type="submit"
                  disabled={savingVariant}
                  className="inline-flex items-center gap-2 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground disabled:opacity-60"
                >
                  {savingVariant ? <Loader2 className="size-4 animate-spin" /> : null}
                  {t("actions.saveChanges")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
