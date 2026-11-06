"use client";

import { useEffect, useState } from "react";
import { Search, Plus, X, Loader2 } from "lucide-react";
import { vendorService } from "@/modules/vendor/api/vendorService";
import type {
  VendorProduct,
  VendorVariant,
} from "@/modules/vendor/types";
import { catalogService } from "@/modules/catalog/api/catalogService";
import type { Brand, Category } from "@/modules/catalog/types";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { apiErrorMessage, notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";
import { useVendorContext } from "./vendorContext";

interface ProductForm {
  title: string;
  slug: string;
  brandId: string;
  categoryId: string;
  description: string;
  productType: "simple" | "variable";
  variationAxes: string;
  sku: string;
  price: string;
  crossesOutPrice: string;
  currency: string;
  initialStock: string;
}

const EMPTY_FORM: ProductForm = {
  title: "",
  slug: "",
  brandId: "",
  categoryId: "",
  description: "",
  productType: "simple",
  variationAxes: "",
  sku: "",
  price: "",
  crossesOutPrice: "",
  currency: "EGP",
  initialStock: "0",
};

function slugify(value: string) {
  return value
    .trim()
    .toLowerCase()
    .normalize("NFKD")
    .replace(/[^\w\s-]/g, "")
    .replace(/[\s_-]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

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

export function VendorProducts() {
  const t = useTranslations("vendor");
  const commercialLock = t("products.commercialLock");

  const [products, setProducts] = useState<VendorProduct[]>([]);
  const [meta, setMeta] = useState<{ page: number; totalPages: number; totalItems: number } | null>(null);
  const [page, setPage] = useState(1);
  const [categories, setCategories] = useState<Category[]>([]);
  const [brands, setBrands] = useState<Brand[]>([]);
  const [variants, setVariants] = useState<VendorVariant[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<Error | null>(null);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingProduct, setEditingProduct] = useState<VendorProduct | null>(null);
  const [formData, setFormData] = useState<ProductForm>(EMPTY_FORM);
  const [saving, setSaving] = useState(false);
  const [lifecycleId, setLifecycleId] = useState<string | null>(null);

  const [selectedProduct, setSelectedProduct] = useState<VendorProduct | null>(null);
  const [editingVariant, setEditingVariant] = useState<VendorVariant | null>(null);
  const [variantForm, setVariantForm] = useState<VariantForm>(EMPTY_VARIANT_FORM);
  const [savingVariant, setSavingVariant] = useState(false);
  const [stockAmount, setStockAmount] = useState("10");
  const [stockReason, setStockReason] = useState<"restock" | "return" | "adjustment">("restock");
  const [savingStock, setSavingStock] = useState(false);

  const loadProducts = async (nextPage = page) => {
    setLoading(true);
    setLoadError(null);
    try {
      const [productResponse, categoryList, brandList] = await Promise.all([
        vendorService.getProducts({ page: nextPage, pageSize: 20 }),
        catalogService.getCategories(100),
        catalogService.getBrands(100),
      ]);
      setProducts(productResponse.data);
      const nextMeta = productResponse.meta as { page?: number; totalPages?: number; totalItems?: number } | undefined;
      if (nextMeta && typeof nextMeta.page === "number") {
        setMeta({ page: nextMeta.page, totalPages: nextMeta.totalPages ?? 1, totalItems: nextMeta.totalItems ?? productResponse.data.length });
      } else {
        setMeta({ page: nextPage, totalPages: 1, totalItems: productResponse.data.length });
      }
      setCategories(Array.isArray(categoryList) ? categoryList : categoryList.data);
      setBrands(Array.isArray(brandList) ? brandList : brandList.data);
    } catch (error) {
      setLoadError(error instanceof Error ? error : new Error(String(error)));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadProducts();
  }, []);

  const openVariants = async (product: VendorProduct) => {
    setSelectedProduct(product);
    setEditingVariant(null);
    setVariantForm(EMPTY_VARIANT_FORM);
    try {
      const response = await vendorService.getVariants(product.id, { pageSize: 100 });
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
    setEditingProduct(null);
    setFormData(EMPTY_FORM);
    setIsModalOpen(true);
  };

  const handleOpenEdit = (product: VendorProduct) => {
    setEditingProduct(product);
    setFormData({
      ...EMPTY_FORM,
      title: product.title,
      slug: product.slug,
      productType: (product.productType as "simple" | "variable") || "simple",
      sku: product.sku ?? "",
      price: product.price ? String(product.price / 100) : "",
      currency: product.currency ?? "EGP",
    });
    setIsModalOpen(true);
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!formData.title.trim()) {
      notify.error(t("products.validation.titleRequired"));
      return;
    }
    if (!editingProduct && !formData.categoryId) {
      notify.error(t("products.validation.categoryRequired"));
      return;
    }
    if (formData.title.trim().length > 200) {
      notify.error(t("products.validation.titleTooLong"));
      return;
    }
    if (!editingProduct && formData.slug.trim() && !/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(formData.slug.trim().toLowerCase())) {
      notify.error(t("products.validation.slugInvalid"));
      return;
    }
    if (formData.productType === "variable" && !editingProduct && formData.variationAxes.split(",").map((a) => a.trim()).filter(Boolean).length === 0) {
      notify.error(t("products.validation.axesRequired"));
      return;
    }
    if (formData.price && parseMajorPrice(formData.price) === undefined) {
      notify.error(t("products.validation.priceInvalid"));
      return;
    }
    if (formData.crossesOutPrice && parseMajorPrice(formData.crossesOutPrice) === undefined) {
      notify.error(t("products.validation.priceInvalid"));
      return;
    }
    if (formData.price && formData.crossesOutPrice && (parseMajorPrice(formData.crossesOutPrice) ?? 0) < (parseMajorPrice(formData.price) ?? 0)) {
      notify.error(t("products.validation.compareAtTooLow"));
      return;
    }
    if (!/^[A-Z]{3}$/.test(formData.currency.trim().toUpperCase())) {
      notify.error(t("products.validation.currencyInvalid"));
      return;
    }
    const initialStockValue = Number.parseInt(formData.initialStock, 10);
    if (!editingProduct && (!Number.isInteger(initialStockValue) || initialStockValue < 0)) {
      notify.error(t("products.validation.stockInvalid"));
      return;
    }
    if (editingProduct && formData.categoryId && !categories.some((category) => category.id === formData.categoryId)) {
      // Existing category details are not returned by the summary API. Only send
      // the category when the form supplies a different one (handled on create).
    }

    const priceMinor = formData.price ? parseMajorPrice(formData.price) : undefined;
    const crossedOutMinor = formData.crossesOutPrice
      ? parseMajorPrice(formData.crossesOutPrice)
      : undefined;
    setSaving(true);

    try {
      if (editingProduct) {
        await vendorService.updateProduct(editingProduct.id, {
          title: formData.title.trim(),
          description: formData.description.trim() || undefined,
          productType: formData.productType,
          variationAxes: formData.productType === "variable"
            ? formData.variationAxes.split(",").map((axis) => axis.trim()).filter(Boolean)
            : [],
        });
        notify.success(apiErrorMessage(null, t("toasts.productUpdated")));
        setIsModalOpen(false);
        await loadProducts();
        return;
      }

      const slug = formData.slug.trim() || slugify(formData.title);
      const created = await vendorService.createProduct({
        brandId: formData.brandId || undefined,
        categoryId: formData.categoryId,
        slug,
        title: formData.title.trim(),
        description: formData.description.trim() || undefined,
        productType: formData.productType,
        variationAxes: formData.productType === "variable"
          ? formData.variationAxes.split(",").map((axis) => axis.trim()).filter(Boolean)
          : undefined,
      });
      const productId = created.data.id;

      if (formData.productType === "simple") {
        const variant = await vendorService.createVariant(productId, {
          title: formData.title.trim(),
          sku: formData.sku.trim() || undefined,
          price: priceMinor,
          crossedOutPrice: crossedOutMinor,
          currency: formData.currency.trim() || "EGP",
          isDefault: true,
        });
        const initialStock = Number.parseInt(formData.initialStock, 10);
        if (Number.isFinite(initialStock) && initialStock !== 0) {
          await vendorService.adjustStock(variant.data.id, {
            quantity: initialStock,
            reason: "restock",
          });
        }
      } else if (formData.price) {
        const variant = await vendorService.createVariant(productId, {
          title: formData.title.trim(),
          sku: formData.sku.trim() || undefined,
          price: priceMinor,
          crossedOutPrice: crossedOutMinor,
          currency: formData.currency.trim() || "EGP",
          isDefault: true,
        });
        const initialStock = Number.parseInt(formData.initialStock, 10);
        if (Number.isFinite(initialStock) && initialStock !== 0) {
          await vendorService.adjustStock(variant.data.id, {
            quantity: initialStock,
            reason: "restock",
          });
        }
      }

      notify.success(t("toasts.productAdded"));
      setIsModalOpen(false);
      await loadProducts();
    } catch (error) {
      notify.error(apiErrorMessage(error, t("products.saveFailed")));
    } finally {
      setSaving(false);
    }
  };

  const handleLifecycle = async (product: VendorProduct, action: "publish" | "archive" | "delete") => {
    const confirmed = action === "delete"
      ? window.confirm(t("products.confirmDelete"))
      : true;
    if (!confirmed) return;

    setLifecycleId(product.id);
    try {
      if (action === "publish") await vendorService.publishProduct(product.id);
      if (action === "archive") await vendorService.archiveProduct(product.id);
      if (action === "delete") await vendorService.deleteProduct(product.id);
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

  const handleOpenVariantEdit = (variant: VendorVariant) => {
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
        await vendorService.updateVariant(editingVariant.id, payload);
        notify.success(t("toasts.variantUpdated"));
      } else {
        await vendorService.createVariant(selectedProduct.id, payload);
        notify.success(t("toasts.variantAdded"));
      }
      const response = await vendorService.getVariants(selectedProduct.id, { pageSize: 100 });
      setVariants(response.data);
      setEditingVariant(null);
      setVariantForm(EMPTY_VARIANT_FORM);
    } catch (error) {
      notify.error(apiErrorMessage(error, t("variants.saveFailed")));
    } finally {
      setSavingVariant(false);
    }
  };

  const handleDeleteVariant = async (variant: VendorVariant) => {
    if (!selectedProduct || !window.confirm(t("variants.confirmDelete"))) return;
    try {
      await vendorService.deleteVariant(variant.id);
      const response = await vendorService.getVariants(selectedProduct.id, { pageSize: 100 });
      setVariants(response.data);
      notify.success(t("toasts.variantRemoved"));
    } catch (error) {
      notify.error(apiErrorMessage(error, t("variants.deleteFailed")));
    }
  };

  const handleAdjustStock = async (variant: VendorVariant) => {
    const quantity = Number.parseInt(stockAmount, 10);
    if (!Number.isFinite(quantity) || quantity === 0) {
      notify.error(t("variants.invalidStockAmount"));
      return;
    }
    setSavingStock(true);
    try {
      await vendorService.adjustStock(variant.id, { quantity, reason: stockReason });
      notify.success(t("toasts.inventoryUpdated"));
      if (selectedProduct) {
        const response = await vendorService.getVariants(selectedProduct.id, { pageSize: 100 });
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

  const statusBadge = (product: VendorProduct) => (
    <StatusPill tone={statusTone(product.status)}>
      {t(`products.status.${product.status}`)}
    </StatusPill>
  );

  return (
    <div className="space-y-5 w-full">
      <PageHeader
        title={t("products.title")}
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
            <div className="overflow-x-auto">
              <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
                <thead className="text-xs text-muted-foreground">
                  <tr className="border-b border-border">
                    <th className="py-2.5 ps-4 text-start font-medium">{t("products.table.details")}</th>
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
                      <td colSpan={6} className="py-10 text-center text-sm text-muted-foreground">
                        {t("products.empty")}
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

      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <form
            onSubmit={handleSubmit}
            className="max-h-[90vh] w-full max-w-lg space-y-4 overflow-y-auto rounded-lg border border-border bg-background p-5 shadow-xl"
          >
            <div className="flex items-center justify-between border-b border-border pb-3">
              <h3 className="text-sm font-semibold text-foreground">
                {editingProduct ? t("products.editTitle") : t("products.addTitle")}
              </h3>
              <button
                type="button"
                onClick={() => setIsModalOpen(false)}
                className="text-muted-foreground hover:text-foreground"
                aria-label={t("actions.cancel")}
              >
                <X className="size-4" />
              </button>
            </div>

            <div className="space-y-1">
              <label className="text-sm font-medium text-foreground">{t("products.form.title")}</label>
              <input
                required
                value={formData.title}
                onChange={(event) => setFormData({ ...formData, title: event.target.value })}
                className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
              />
            </div>

            {!editingProduct && (
              <>
                <div className="space-y-1">
                  <label className="text-sm font-medium text-foreground">{t("products.form.slug")}</label>
                  <input
                    value={formData.slug}
                    onChange={(event) => setFormData({ ...formData, slug: event.target.value })}
                    placeholder={formData.title ? slugify(formData.title) : t("products.form.slugPlaceholder")}
                    className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                  />
                </div>
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <div className="space-y-1">
                    <label className="text-sm font-medium text-foreground">{t("products.form.category")}</label>
                    <select
                      required
                      value={formData.categoryId}
                      onChange={(event) => setFormData({ ...formData, categoryId: event.target.value })}
                      className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                    >
                      <option value="">{t("products.form.chooseCategory")}</option>
                      {categories.map((category) => (
                        <option key={category.id} value={category.id}>{category.name}</option>
                      ))}
                    </select>
                  </div>
                  <div className="space-y-1">
                    <label className="text-sm font-medium text-foreground">{t("products.form.brand")}</label>
                    <select
                      value={formData.brandId}
                      onChange={(event) => setFormData({ ...formData, brandId: event.target.value })}
                      className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                    >
                      <option value="">{t("products.form.noBrand")}</option>
                      {brands.map((brand) => (
                        <option key={brand.id} value={brand.id}>{brand.name}</option>
                      ))}
                    </select>
                  </div>
                </div>
              </>
            )}

            <div className="space-y-1">
              <label className="text-sm font-medium text-foreground">{t("products.form.description")}</label>
              <textarea
                value={formData.description}
                onChange={(event) => setFormData({ ...formData, description: event.target.value })}
                rows={4}
                className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
              />
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div className="space-y-1">
                <label className="text-sm font-medium text-foreground">{t("products.form.productType")}</label>
                <select
                  value={formData.productType}
                  onChange={(event) => setFormData({ ...formData, productType: event.target.value as "simple" | "variable" })}
                  className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                >
                  <option value="simple">{t("products.form.simple")}</option>
                  <option value="variable">{t("products.form.variable")}</option>
                </select>
              </div>
              {formData.productType === "variable" && (
                <div className="space-y-1">
                  <label className="text-sm font-medium text-foreground">{t("products.form.variationAxes")}</label>
                  <input
                    value={formData.variationAxes}
                    onChange={(event) => setFormData({ ...formData, variationAxes: event.target.value })}
                    placeholder={t("products.form.variationAxesPlaceholder")}
                    className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                  />
                </div>
              )}
            </div>

            {!editingProduct && (
              <div className="space-y-4 rounded-md border border-border p-4">
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <div className="space-y-1">
                    <label className="text-sm font-medium text-foreground">{t("products.form.skuCode")}</label>
                    <input
                      value={formData.sku}
                      onChange={(event) => setFormData({ ...formData, sku: event.target.value })}
                      className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-sm font-medium text-foreground">{t("products.form.sellingPrice")}</label>
                    <input
                      type="number"
                      min="0.01"
                      step="0.01"
                      value={formData.price}
                      onChange={(event) => setFormData({ ...formData, price: event.target.value })}
                      className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                    />
                  </div>
                </div>
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                  <div className="space-y-1">
                    <label className="text-sm font-medium text-foreground">{t("products.form.compareAtPrice")}</label>
                    <input
                      type="number"
                      min="0"
                      step="0.01"
                      value={formData.crossesOutPrice}
                      onChange={(event) => setFormData({ ...formData, crossesOutPrice: event.target.value })}
                      className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-sm font-medium text-foreground">{t("products.form.currency")}</label>
                    <input
                      value={formData.currency}
                      onChange={(event) => setFormData({ ...formData, currency: event.target.value.toUpperCase() })}
                      className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-sm font-medium text-foreground">{t("products.form.initialStock")}</label>
                    <input
                      type="number"
                      min="0"
                      value={formData.initialStock}
                      onChange={(event) => setFormData({ ...formData, initialStock: event.target.value })}
                      className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                    />
                  </div>
                </div>
                <p className="text-xs text-muted-foreground">{t("products.form.initialPurchaseNote")}</p>
              </div>
            )}

            <div className="flex justify-end gap-2 border-t border-border pt-3">
              <button
                type="button"
                onClick={() => setIsModalOpen(false)}
                className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground hover:bg-secondary"
              >
                {t("actions.cancel")}
              </button>
              <button
                type="submit"
                disabled={saving}
                className="inline-flex items-center gap-2 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground disabled:opacity-60"
              >
                {saving ? <Loader2 className="size-4 animate-spin" /> : null}
                {t("actions.saveChanges")}
              </button>
            </div>
          </form>
        </div>
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
                      <td colSpan={6} className="py-8 text-center text-sm text-muted-foreground">
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
