"use client";

import { useEffect, useState } from "react";
import {
  Search,
  Plus,
  Edit2,
  Trash2,
  X,
  Archive,
  Upload,
  Layers3,
  Loader2,
  Package2,
  Link2,
  Link2Off,
} from "lucide-react";
import { vendorService } from "@/modules/vendor/api/vendorService";
import type {
  VendorProduct,
  VendorVariant,
} from "@/modules/vendor/types";
import { catalogService } from "@/modules/catalog/api/catalogService";
import type { Brand, Category } from "@/modules/catalog/types";
import { ErrorState } from "@/shared/ui/errorState";
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

  const loadProducts = async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const [productResponse, categoryList, brandList] = await Promise.all([
        vendorService.getProducts({ pageSize: 100 }),
        catalogService.getCategories(100),
        catalogService.getBrands(100),
      ]);
      setProducts(productResponse.data);
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
    if (!formData.title.trim() || (!editingProduct && !formData.categoryId)) return;
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
    <span
      className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${
        product.status === "published"
          ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400"
          : product.status === "archived"
            ? "bg-secondary text-muted-foreground"
            : "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-400"
      }`}
    >
      {t(`products.status.${product.status}`)}
    </span>
  );

  return (
    <div className="space-y-6 w-full">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("products.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t("products.subtitle")}</p>
        </div>
        <button
          onClick={handleOpenAdd}
          className="flex items-center gap-2 px-4 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0"
        >
          <Plus className="size-4" />
          <span>{t("actions.addNewProduct")}</span>
        </button>
      </div>

      {loading ? (
        <div className="flex min-h-[40vh] items-center justify-center">
          <Loader2 className="size-8 animate-spin text-accent-brand" />
        </div>
      ) : loadError ? (
        <ErrorState
          title={t("products.loadTitle")}
          description={loadError.message || t("products.loadDescription")}
          onRetry={() => void loadProducts()}
        />
      ) : (
        <>
          <div className="bg-background border border-border rounded-2xl p-4 shadow-xs flex flex-col sm:flex-row sm:items-center gap-4">
            <div className="relative max-w-sm flex-1">
              <Search className="absolute left-3.5 top-2.5 size-4 text-muted-foreground" />
              <input
                type="text"
                placeholder={t("products.searchPlaceholder")}
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                className="w-full pl-10 pr-4 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-accent-brand"
              />
            </div>
            <select
              value={statusFilter}
              onChange={(event) => setStatusFilter(event.target.value)}
              className="px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              aria-label={t("products.table.status")}
            >
              <option value="all">{t("products.filterAll")}</option>
              <option value="draft">{t("products.status.draft")}</option>
              <option value="published">{t("products.status.published")}</option>
              <option value="archived">{t("products.status.archived")}</option>
            </select>
          </div>

          <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
            <div className="overflow-x-auto">
              <table className="w-full min-w-max text-left text-xs">
                <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
                  <tr>
                    <th className="py-3.5 pl-4">{t("products.table.details")}</th>
                    <th className="py-3.5">{t("products.table.sku")}</th>
                    <th className="py-3.5">{t("products.table.price")}</th>
                    <th className="py-3.5">{t("products.table.stock")}</th>
                    <th className="py-3.5">{t("products.table.status")}</th>
                    <th className="py-3.5 pr-4 text-right">{t("products.table.actions")}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/60">
                  {filtered.map((product) => (
                    <tr key={product.id} className="hover:bg-secondary/20 transition-colors">
                      <td className="py-3.5 pl-4">
                        <div className="flex items-center gap-3">
                          {product.thumbnailUrl ? (
                            <img src={product.thumbnailUrl} alt={product.title} className="size-10 rounded-xl object-cover border border-border shrink-0" />
                          ) : (
                            <div className="size-10 rounded-xl border border-border bg-secondary/40 flex items-center justify-center shrink-0">
                              <Package2 className="size-4 text-muted-foreground" />
                            </div>
                          )}
                          <div>
                            <div className="font-semibold text-foreground text-xs">{product.title}</div>
                            <div className="text-[10px] text-muted-foreground">{product.slug}</div>
                          </div>
                        </div>
                      </td>
                      <td className="py-3.5 font-mono text-muted-foreground">{product.sku ?? "—"}</td>
                      <td className="py-3.5 font-bold text-foreground">
                        {product.price ? `${product.currency ?? "EGP"} ${(product.price / 100).toFixed(2)}` : "—"}
                      </td>
                      <td className="py-3.5 font-bold text-foreground">{t("products.units", { count: product.stock })}</td>
                      <td className="py-3.5">{statusBadge(product)}</td>
                      <td className="py-3.5 pr-4 text-right">
                        <div className="inline-flex items-center gap-1">
                          <button
                            onClick={() => void openVariants(product)}
                            className="inline-flex items-center gap-1 px-2 py-1 text-muted-foreground hover:text-accent-brand"
                            title={t("variants.manage")}
                          >
                            <Layers3 className="size-4" />
                            <span className="hidden xl:inline">{t("variants.manage")}</span>
                          </button>
                          <button onClick={() => handleOpenEdit(product)} className="p-1.5 text-muted-foreground hover:text-accent-brand" title={t("actions.edit")}>
                            <Edit2 className="size-4" />
                          </button>
                          {product.status !== "published" && (
                            <button
                              onClick={() => void handleLifecycle(product, "publish")}
                              disabled={lifecycleId === product.id}
                              className="p-1.5 text-muted-foreground hover:text-emerald-600 disabled:opacity-50"
                              title={t("actions.publish")}
                            >
                              <Upload className="size-4" />
                            </button>
                          )}
                          {product.status === "published" && (
                            <button
                              onClick={() => void handleLifecycle(product, "archive")}
                              disabled={lifecycleId === product.id}
                              className="p-1.5 text-muted-foreground hover:text-amber-600 disabled:opacity-50"
                              title={t("actions.archive")}
                            >
                              <Archive className="size-4" />
                            </button>
                          )}
                          <button
                            onClick={() => void handleLifecycle(product, "delete")}
                            disabled={lifecycleId === product.id}
                            className="p-1.5 text-muted-foreground hover:text-destructive disabled:opacity-50"
                            title={t("actions.delete")}
                          >
                            <Trash2 className="size-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                  {filtered.length === 0 && (
                    <tr>
                      <td colSpan={6} className="py-10 text-center text-muted-foreground">
                        {t("products.empty")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}

      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <form onSubmit={handleSubmit} className="bg-background border border-border rounded-2xl max-w-lg w-full p-6 shadow-xl space-y-4 max-h-[90vh] overflow-y-auto">
            <div className="flex justify-between items-center border-b border-border pb-3">
              <h3 className="font-bold text-base text-foreground">
                {editingProduct ? t("products.editTitle") : t("products.addTitle")}
              </h3>
              <button type="button" onClick={() => setIsModalOpen(false)} className="text-muted-foreground">
                <X className="size-5" />
              </button>
            </div>

            <div className="space-y-1">
              <label className="text-xs font-semibold text-foreground">{t("products.form.title")}</label>
              <input
                required
                value={formData.title}
                onChange={(event) => setFormData({ ...formData, title: event.target.value })}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
            </div>

            {!editingProduct && (
              <>
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">{t("products.form.slug")}</label>
                  <input
                    value={formData.slug}
                    onChange={(event) => setFormData({ ...formData, slug: event.target.value })}
                    placeholder={formData.title ? slugify(formData.title) : t("products.form.slugPlaceholder")}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                  />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-1">
                    <label className="text-xs font-semibold text-foreground">{t("products.form.category")}</label>
                    <select
                      required
                      value={formData.categoryId}
                      onChange={(event) => setFormData({ ...formData, categoryId: event.target.value })}
                      className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                    >
                      <option value="">{t("products.form.chooseCategory")}</option>
                      {categories.map((category) => (
                        <option key={category.id} value={category.id}>{category.name}</option>
                      ))}
                    </select>
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs font-semibold text-foreground">{t("products.form.brand")}</label>
                    <select
                      value={formData.brandId}
                      onChange={(event) => setFormData({ ...formData, brandId: event.target.value })}
                      className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
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
              <label className="text-xs font-semibold text-foreground">{t("products.form.description")}</label>
              <textarea
                value={formData.description}
                onChange={(event) => setFormData({ ...formData, description: event.target.value })}
                rows={4}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">{t("products.form.productType")}</label>
                <select
                  value={formData.productType}
                  onChange={(event) => setFormData({ ...formData, productType: event.target.value as "simple" | "variable" })}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                >
                  <option value="simple">{t("products.form.simple")}</option>
                  <option value="variable">{t("products.form.variable")}</option>
                </select>
              </div>
              {formData.productType === "variable" && (
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">{t("products.form.variationAxes")}</label>
                  <input
                    value={formData.variationAxes}
                    onChange={(event) => setFormData({ ...formData, variationAxes: event.target.value })}
                    placeholder={t("products.form.variationAxesPlaceholder")}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                  />
                </div>
              )}
            </div>

            {!editingProduct && (
              <div className="space-y-4 rounded-xl border border-border p-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-1">
                    <label className="text-xs font-semibold text-foreground">{t("products.form.skuCode")}</label>
                    <input
                      value={formData.sku}
                      onChange={(event) => setFormData({ ...formData, sku: event.target.value })}
                      className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs font-semibold text-foreground">{t("products.form.sellingPrice")}</label>
                    <input
                      type="number"
                      min="0.01"
                      step="0.01"
                      value={formData.price}
                      onChange={(event) => setFormData({ ...formData, price: event.target.value })}
                      className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                    />
                  </div>
                </div>
                <div className="grid grid-cols-3 gap-4">
                  <div className="space-y-1">
                    <label className="text-xs font-semibold text-foreground">{t("products.form.compareAtPrice")}</label>
                    <input
                      type="number"
                      min="0"
                      step="0.01"
                      value={formData.crossesOutPrice}
                      onChange={(event) => setFormData({ ...formData, crossesOutPrice: event.target.value })}
                      className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs font-semibold text-foreground">{t("products.form.currency")}</label>
                    <input
                      value={formData.currency}
                      onChange={(event) => setFormData({ ...formData, currency: event.target.value.toUpperCase() })}
                      className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs font-semibold text-foreground">{t("products.form.initialStock")}</label>
                    <input
                      type="number"
                      min="0"
                      value={formData.initialStock}
                      onChange={(event) => setFormData({ ...formData, initialStock: event.target.value })}
                      className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                    />
                  </div>
                </div>
                <p className="text-[11px] text-muted-foreground">{t("products.form.initialPurchaseNote")}</p>
              </div>
            )}

            <div className="flex justify-end gap-3 pt-3 border-t border-border">
              <button type="button" onClick={() => setIsModalOpen(false)} className="px-4 py-2 text-xs border border-border rounded-xl">
                {t("actions.cancel")}
              </button>
              <button type="submit" disabled={saving} className="inline-flex items-center gap-2 px-4 py-2 text-xs bg-accent-brand text-white font-semibold rounded-xl disabled:opacity-60">
                {saving && <Loader2 className="size-4 animate-spin" />}
                {t("actions.saveChanges")}
              </button>
            </div>
          </form>
        </div>
      )}

      {selectedProduct && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-3xl w-full p-6 shadow-xl space-y-4 max-h-[92vh] overflow-y-auto">
            <div className="flex justify-between items-start border-b border-border pb-3">
              <div>
                <h3 className="font-bold text-base text-foreground">{t("variants.title")}</h3>
                <p className="text-xs text-muted-foreground mt-0.5">{selectedProduct.title}</p>
              </div>
              <button onClick={closeVariants} className="text-muted-foreground">
                <X className="size-5" />
              </button>
            </div>

            <div className="rounded-xl border border-border bg-secondary/30 p-3">
              <div className="flex items-center justify-between gap-3 text-xs">
                <span className="inline-flex items-center gap-1.5 font-semibold">
                  {commercialLock ? <Link2 className="size-3.5" /> : <Link2Off className="size-3.5" />}
                  {t("variants.sellable")}
                </span>
                <span className="text-muted-foreground">{t("variants.sellableNote")}</span>
              </div>
            </div>

            <div className="overflow-x-auto rounded-xl border border-border">
              <table className="w-full min-w-max text-left text-xs">
                <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
                  <tr>
                    <th className="py-2.5 pl-3">{t("variants.table.title")}</th>
                    <th className="py-2.5">{t("variants.table.sku")}</th>
                    <th className="py-2.5">{t("variants.table.price")}</th>
                    <th className="py-2.5">{t("variants.table.default")}</th>
                    <th className="py-2.5">{t("variants.table.stockAdjustment")}</th>
                    <th className="py-2.5 pr-3 text-right">{t("variants.table.actions")}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/60">
                  {variants.map((variant) => (
                    <tr key={variant.id}>
                      <td className="py-2.5 pl-3 font-semibold">{variant.title}</td>
                      <td className="py-2.5 font-mono text-muted-foreground">{variant.sku}</td>
                      <td className="py-2.5">{variant.price ? `${variant.currency ?? "EGP"} ${(variant.price / 100).toFixed(2)}` : "—"}</td>
                      <td className="py-2.5">{variant.isDefault ? t("variants.defaultYes") : t("variants.defaultNo")}</td>
                      <td className="py-2.5">
                        <div className="flex items-center gap-2">
                          <input
                            type="number"
                            value={stockAmount}
                            onChange={(event) => setStockAmount(event.target.value)}
                            className="w-20 px-2 py-1.5 bg-secondary/40 border border-border rounded-lg text-xs"
                            aria-label={t("variants.stockAmount")}
                          />
                          <select
                            value={stockReason}
                            onChange={(event) => setStockReason(event.target.value as "restock" | "return" | "adjustment")}
                            className="px-2 py-1.5 bg-secondary/40 border border-border rounded-lg text-xs"
                            aria-label={t("variants.stockReason")}
                          >
                            <option value="restock">{t("variants.reasonRestock")}</option>
                            <option value="return">{t("variants.reasonReturn")}</option>
                            <option value="adjustment">{t("variants.reasonAdjustment")}</option>
                          </select>
                          <button
                            onClick={() => void handleAdjustStock(variant)}
                            disabled={savingStock}
                            className="px-2 py-1.5 bg-accent-brand text-white rounded-lg disabled:opacity-50"
                          >
                            {t("variants.applyStock")}
                          </button>
                        </div>
                      </td>
                      <td className="py-2.5 pr-3 text-right">
                        <button onClick={() => handleOpenVariantEdit(variant)} className="p-1.5 text-muted-foreground hover:text-accent-brand">
                          <Edit2 className="size-4" />
                        </button>
                        <button onClick={() => void handleDeleteVariant(variant)} className="p-1.5 text-muted-foreground hover:text-destructive">
                          <Trash2 className="size-4" />
                        </button>
                      </td>
                    </tr>
                  ))}
                  {variants.length === 0 && (
                    <tr>
                      <td colSpan={6} className="py-8 text-center text-muted-foreground">{t("variants.empty")}</td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            <form onSubmit={handleSubmitVariant} className="space-y-4 rounded-xl border border-border p-4">
              <div className="flex items-center justify-between">
                <h4 className="font-bold text-sm">{editingVariant ? t("variants.edit") : t("variants.add")}</h4>
                <button
                  type="button"
                  onClick={handleOpenVariantCreate}
                  className="inline-flex items-center gap-1 text-xs font-semibold text-accent-brand"
                >
                  <Plus className="size-3.5" /> {t("variants.add")}
                </button>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <label className="text-xs font-semibold">{t("variants.form.title")}</label>
                  <input
                    required
                    value={variantForm.title}
                    onChange={(event) => setVariantForm({ ...variantForm, title: event.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-semibold">{t("variants.form.sku")}</label>
                  <input
                    value={variantForm.sku}
                    onChange={(event) => setVariantForm({ ...variantForm, sku: event.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs"
                  />
                </div>
              </div>
              <div className="grid grid-cols-3 gap-4">
                <div className="space-y-1">
                  <label className="text-xs font-semibold">{t("variants.form.price")}</label>
                  <input
                    type="number"
                    min="0.01"
                    step="0.01"
                    value={variantForm.price}
                    onChange={(event) => setVariantForm({ ...variantForm, price: event.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-semibold">{t("variants.form.compareAtPrice")}</label>
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    value={variantForm.crossedOutPrice}
                    onChange={(event) => setVariantForm({ ...variantForm, crossedOutPrice: event.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-semibold">{t("variants.form.currency")}</label>
                  <input
                    value={variantForm.currency}
                    onChange={(event) => setVariantForm({ ...variantForm, currency: event.target.value.toUpperCase() })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs"
                  />
                </div>
              </div>
              <div className="space-y-1">
                <label className="text-xs font-semibold">{t("variants.form.attributes")}</label>
                <input
                  value={variantForm.attributes}
                  onChange={(event) => setVariantForm({ ...variantForm, attributes: event.target.value })}
                  placeholder={t("variants.form.attributesPlaceholder")}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs"
                />
              </div>
              <label className="flex items-center gap-2 text-xs font-semibold">
                <input
                  type="checkbox"
                  checked={variantForm.isDefault}
                  onChange={(event) => setVariantForm({ ...variantForm, isDefault: event.target.checked })}
                />
                {t("variants.form.isDefault")}
              </label>
              <div className="flex justify-end">
                <button type="submit" disabled={savingVariant} className="inline-flex items-center gap-2 px-4 py-2 text-xs bg-accent-brand text-white font-semibold rounded-xl disabled:opacity-60">
                  {savingVariant && <Loader2 className="size-4 animate-spin" />}
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
