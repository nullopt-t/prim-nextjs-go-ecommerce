import { ResponsiveTable } from "@/components/ui/responsiveTable";
import { useState, useEffect, useCallback, useMemo } from "react";
import {
  Plus,
  Search,
  Filter,
  Edit2,
  Trash2,
  History,
  X,
  Star,
  Package,
  Loader2
} from "lucide-react";
import { catalogService } from "@/services/catalog";
import { adminService } from "@/services/admin";
import { useCatalogContext } from "@/context/CatalogContext";
import { useTranslations } from "next-intl";
import { notify, apiErrorMessage } from "@/lib/notify";
import { ErrorState } from "@/components/ui/errorState";
import { CursorPager, useCursorPagination } from "@/components/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/lib/pagination";

export function AdminProducts() {
  const t = useTranslations("admin");
  const [products, setProducts] = useState<any[]>([]);
  const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [categories, setCategories] = useState<any[]>([]);
  const [brands, setBrands] = useState<any[]>([]);
  const [attributes, setAttributes] = useState<any[]>([]);
  const [productDetail, setProductDetail] = useState<any>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [variantForm, setVariantForm] = useState<any>(null);
  const [savingVariant, setSavingVariant] = useState(false);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("All");
  const { refreshCatalog } = useCatalogContext() || {};
  const { cursor, hasPrev, next, prev, reset, page } = useCursorPagination();

  const loadAll = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);

      const productQuery = new URLSearchParams({ pageSize: "20", keyset: "true", _t: String(Date.now()) });
      if (cursor) productQuery.set("cursor", cursor);
      if (debouncedSearch) productQuery.set("search", debouncedSearch);
      if (selectedCategory !== "All") productQuery.set("category", selectedCategory);

      const [prodsRes, catsRes, brandsRes, attrsRes]: [any, any, any, any] = await Promise.all([
        catalogService.getProducts(productQuery.toString()),
        catalogService.getCategories(100),
        catalogService.getBrands(100),
        adminService.getAttributes()
      ]);

      const catList = Array.isArray(catsRes) ? catsRes : (Array.isArray(catsRes?.data) ? catsRes.data : []);
      setCategories(catList);

      const brandList = Array.isArray(brandsRes) ? brandsRes : (Array.isArray(brandsRes?.data) ? brandsRes.data : []);
      setBrands(brandList);

      const attrList = Array.isArray(attrsRes) ? attrsRes : (Array.isArray(attrsRes?.data) ? attrsRes.data : []);
      setAttributes(attrList);

      const list = Array.isArray(prodsRes) ? prodsRes : (Array.isArray(prodsRes?.data) ? prodsRes.data : []);
      const formatted = list.map((p: any) => {
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
          id: p.id || p.slug,
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
          status: p.status || (p.inStock || p.stockQuantity > 0 ? "Active" : "Out of Stock"),
          image: p.thumbnail || p.img || p.image || "/placeholder.png",
          variants: p.variants || [],
          salesCount: 0
        };
      });
      setProducts(formatted);
      if (prodsRes?.meta) setMeta(prodsRes.meta);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [cursor, debouncedSearch, selectedCategory]);

  useEffect(() => {
    loadAll();
  }, [loadAll]);

  // Debounce the search box and restart pagination whenever the query changes.
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search);
      reset();
    }, 350);
    return () => clearTimeout(timer);
  }, [search, reset]);

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingProduct, setEditingProduct] = useState(null);
  const [auditProduct, setAuditProduct] = useState<any>(null);
  const [auditLogs, setAuditLogs] = useState<any[]>([]);
  const [loadingAudit, setLoadingAudit] = useState(false);

  const handleViewAudit = async (prod: any) => {
    setAuditProduct(prod);
    setLoadingAudit(true);
    try {
      const res = await adminService.getInventoryLedgers({
        pageSize: 50,
        productId: prod.id
      });
      const list = Array.isArray(res?.data) ? res.data : (Array.isArray(res) ? res : []);
      const mapped = list.map((l: any) => ({
        id: l.id,
        productName: l.productName || prod.name,
        variantTitle: l.variantTitle || "Default Variant",
        sku: l.sku || prod.sku,
        change: l.quantity > 0 ? `+${l.quantity}` : `${l.quantity}`,
        quantity: l.quantity,
        reason: l.reason || "restock",
        date: l.createdAt ? new Date(l.createdAt).toLocaleString() : "Recently",
        ledgerId: l.id ? l.id.slice(0, 8) : ""
      }));
      setAuditLogs(mapped);

      // Always guarantee Current On-Hand Stock is 100% in sync with the ledger balance
      const ledgerSum = mapped.reduce((acc: number, curr: any) => acc + (curr.quantity || 0), 0);
      setAuditProduct((prev: any) => prev ? { ...prev, stock: ledgerSum } : prev);
      // Also update the row in products list state if it differed
      setProducts((prev: any[]) => prev.map((p) => p.id === prod.id ? { ...p, stock: ledgerSum } : p));
    } catch {
      notify.error(t("products.toast.auditFailed"));
    } finally {
      setLoadingAudit(false);
    }
  };

  // Form State
  const [formData, setFormData] = useState({
    name: "",
    sku: "",
    category: categories[0]?.name || "Electronics",
    brand: brands[0]?.name || "SoundMaster",
    price: "",
    compareAtPrice: "",
    cost: "",
    stock: "",
    adjustmentReason: "adjustment",
    status: "Active",
    image: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400&q=80"
  });

  const saveProductsToState = (updated: any[]) => {
    setProducts(updated);
  };

  const handleOpenCreateModal = () => {
    setEditingProduct(null);
    setFormData({
      name: "",
      sku: `SKU-${Math.floor(1000 + Math.random() * 9000)}`,
      category: categories[0]?.name || "Electronics",
      brand: brands[0]?.name || "SoundMaster",
      price: "99.99",
      compareAtPrice: "120.00",
      cost: "45.00",
      stock: "25",
      adjustmentReason: "restock",
      status: "Active",
      image: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400&q=80"
    });
    setIsModalOpen(true);
  };

  const loadProductDetail = useCallback(async (productId: string) => {
    try {
      setDetailLoading(true);
      const res: any = await catalogService.getAdminProduct(productId);
      setProductDetail(res?.data ?? res ?? null);
    } catch {
      setProductDetail(null);
      notify.error(t("products.toast.variantLoadFailed"));
    } finally {
      setDetailLoading(false);
    }
  }, [t]);

  const handleOpenEditModal = (prod) => {
    setEditingProduct(prod);
    setProductDetail(null);
    setVariantForm(null);
    setFormData({
      name: prod.name,
      sku: prod.sku,
      category: prod.category,
      brand: prod.brand,
      price: prod.price.toString(),
      compareAtPrice: prod.compareAtPrice?.toString() || "",
      cost: prod.cost?.toString() || "",
      stock: prod.stock.toString(),
      adjustmentReason: "adjustment",
      status: prod.status,
      image: prod.image
    });
    setIsModalOpen(true);
    loadProductDetail(prod.id);
  };

  const handleSaveProduct = async (e) => {
    e.preventDefault();
    if (!formData.name || !formData.price || !formData.stock) {
      notify.error(t("products.toast.requiredFields"));
      return;
    }

    if (editingProduct) {
      const newStock = parseInt(formData.stock, 10);
      const oldStock = typeof editingProduct.stock === "number" ? editingProduct.stock : 0;
      const stockDiff = newStock - oldStock;

      const updatedProductData = {
        name: formData.name,
        sku: formData.sku,
        category: formData.category,
        brand: formData.brand,
        price: parseFloat(formData.price),
        compareAtPrice: formData.compareAtPrice ? parseFloat(formData.compareAtPrice) : null,
        cost: formData.cost ? parseFloat(formData.cost) : null,
        stock: newStock,
        status: formData.status,
        image: formData.image
      };

      // Optimistically update products table immediately
      setProducts((prev) =>
        prev.map((p) =>
          p.id === editingProduct.id
            ? { ...p, ...updatedProductData }
            : p
        )
      );
      setIsModalOpen(false);

      try {
        await catalogService.updateProduct(editingProduct.id, updatedProductData);
        // If stock changed or was edited, record an audit event in inventory_ledgers
        if (stockDiff !== 0) {
          const reason = formData.adjustmentReason || (stockDiff > 0 ? "restock" : "adjustment");
          if (editingProduct.variantId) {
            await adminService.adjustInventory(editingProduct.variantId, stockDiff, reason);
          } else {
            await adminService.adjustProductInventory(editingProduct.id, stockDiff, reason);
          }
          notify.success(t("products.toast.updatedLedger", { change: stockDiff > 0 ? `+${stockDiff}` : stockDiff }));
        } else {
          notify.success(t("products.toast.updated"));
        }
        await loadAll();
        refreshCatalog?.();
      } catch (err: any) {
        notify.error(apiErrorMessage(err, t("products.toast.updatePartialFailed")));
      }
    } else {
      const newId = `prod-${Date.now()}`;
      const newProd = {
        id: newId,
        name: formData.name,
        sku: formData.sku,
        category: formData.category,
        brand: formData.brand,
        price: parseFloat(formData.price),
        compareAtPrice: formData.compareAtPrice ? parseFloat(formData.compareAtPrice) : null,
        cost: formData.cost ? parseFloat(formData.cost) : null,
        stock: parseInt(formData.stock, 10),
        lowStockThreshold: 10,
        status: formData.status,
        image: formData.image,
        variants: [
          { id: `v-${Date.now()}`, name: "Default Variant", sku: formData.sku, price: parseFloat(formData.price), stock: parseInt(formData.stock, 10) }
        ],
        salesCount: 0
      };
      setProducts((prev) => [newProd, ...prev]);
      setIsModalOpen(false);

      try {
        await catalogService.createProduct(newProd);
        await loadAll();
        refreshCatalog?.();
      } catch {
        // Ignored in mock
      }

      notify.success(t("products.toast.created"));
    }
  };

  const productCategoryAttributes = useMemo(() => {
    const categoryId = productDetail?.categoryId;
    return attributes.filter((attribute: any) => !attribute.categoryId || attribute.categoryId === categoryId);
  }, [attributes, productDetail]);

  const isSimpleProduct = productDetail?.productType === "simple";
  // A simple product may create its initial variant, but afterwards its default is fixed.
  const variantLocked = isSimpleProduct && (productDetail?.variants?.length ?? 0) > 0;
  const canAddVariant = Boolean(productDetail) && !variantLocked;

  const attributeDefaults = useCallback((existing: Record<string, any> = {}) => {
    const values: Record<string, any> = {};
    for (const attribute of productCategoryAttributes) {
      if (existing[attribute.key] !== undefined) {
        values[attribute.key] = existing[attribute.key];
      } else if (attribute.valueType === "boolean") {
        values[attribute.key] = false;
      } else if (attribute.valueType === "enum" && Array.isArray(attribute.allowedValues) && attribute.allowedValues.length > 0) {
        values[attribute.key] = attribute.allowedValues[0];
      } else {
        values[attribute.key] = "";
      }
    }
    return values;
  }, [productCategoryAttributes]);

  const openVariantCreate = () => {
    setVariantForm({
      mode: "create",
      id: null,
      title: "",
      sku: "",
      price: "",
      originalPrice: "",
      currency: "EGP",
      initialStock: "0",
      isDefault: false,
      attributes: attributeDefaults(),
    });
  };

  const openVariantEdit = (variant: any) => {
    setVariantForm({
      mode: "edit",
      id: variant.id,
      title: variant.title || "",
      sku: variant.sku || "",
      price: variant.extractedPrice != null ? String(variant.extractedPrice) : "",
      originalPrice: variant.extractedOriginalPrice != null ? String(variant.extractedOriginalPrice) : "",
      currency: variant.currency || "EGP",
      initialStock: "",
      isDefault: Boolean(variant.isDefault),
      attributes: attributeDefaults(variant.attributes || {}),
    });
  };

  const handleSaveVariant = async () => {
    if (!variantForm || !productDetail) return;

    const price = parseFloat(variantForm.price);
    if (!variantForm.title.trim() || Number.isNaN(price)) {
      notify.error(t("products.variants.requiredFields"));
      return;
    }

    setSavingVariant(true);
    try {
      if (variantForm.mode === "create") {
        await catalogService.createVariant(productDetail.id, {
          title: variantForm.title.trim(),
          price: Math.round(price * 100),
          crossedOutPrice: variantForm.originalPrice ? Math.round(parseFloat(variantForm.originalPrice) * 100) : undefined,
          currency: variantForm.currency || "EGP",
          attributes: variantForm.attributes,
          initialStock: variantForm.initialStock ? parseInt(variantForm.initialStock, 10) : undefined,
          isDefault: Boolean(variantForm.isDefault),
        });
        notify.success(t("products.variants.created"));
      } else {
        await catalogService.updateVariant(variantForm.id, {
          title: variantForm.title.trim(),
          sku: variantForm.sku || undefined,
          price: Math.round(price * 100),
          crossedOutPrice: variantForm.originalPrice ? Math.round(parseFloat(variantForm.originalPrice) * 100) : undefined,
          currency: variantForm.currency || "EGP",
          attributes: variantForm.attributes,
        });
        notify.success(t("products.variants.updated"));
      }

      setVariantForm(null);
      await Promise.all([loadProductDetail(productDetail.id), loadAll()]);
      refreshCatalog?.();
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.variants.saveFailed")));
    } finally {
      setSavingVariant(false);
    }
  };

  const handleDeleteVariant = async (variant: any) => {
    if (!productDetail || !confirm(t("products.variants.confirmDelete", { title: variant.title }))) return;
    try {
      await catalogService.deleteVariant(variant.id);
      notify.success(t("products.variants.deleted"));
      await Promise.all([loadProductDetail(productDetail.id), loadAll()]);
      refreshCatalog?.();
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.variants.deleteFailed")));
    }
  };

  const handleSetDefaultVariant = async (variant: any) => {
    if (!productDetail) return;
    try {
      await catalogService.setDefaultVariant(productDetail.id, variant.id);
      notify.success(t("products.variants.defaultSet"));
      await loadProductDetail(productDetail.id);
      refreshCatalog?.();
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("products.variants.defaultFailed")));
    }
  };

  const handleDeleteProduct = async (id) => {
    if (confirm(t("products.confirmDelete"))) {
      const updated = products.filter((p) => p.id !== id);
      saveProductsToState(updated);
      try {
        await catalogService.deleteProduct(id);
        refreshCatalog?.();
      } catch {
        // Ignored in mock
      }
      notify.success(t("products.toast.deleted"));
    }
  };

  const filteredProducts = products;

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
        <button
          onClick={handleOpenCreateModal}
          className="flex items-center gap-2 px-4 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0"
        >
          <Plus className="size-4" />
          <span>{t("products.addNew")}</span>
        </button>
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
            {categories.map((c) => (
              <option key={c.id} value={c.name}>
                {c.name}
              </option>
            ))}
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
              {filteredProducts.length > 0 ? (
                filteredProducts.map((prod) => (
                  <tr key={prod.id} className="hover:bg-secondary/20 transition-colors">
                    <td className="py-3.5 ps-4">
                      <div className="flex items-center gap-3">
                        <img
                          src={prod.image}
                          alt={prod.name}
                          className="size-11 rounded-xl object-cover border border-border"
                        />
                        <div>
                          <div className="font-semibold text-foreground text-xs line-clamp-1">
                            {prod.name}
                          </div>
                          <div className="text-[10px] text-muted-foreground">
                            {t("products.brandLabel", { name: typeof prod.brand === "object" ? prod.brand?.name : String(prod.brand || "PRIM") })}
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
                          onClick={() => handleOpenEditModal(prod)}
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
          shown={filteredProducts.length}
        />
      </div>

      {/* Create / Edit Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-xl w-full p-6 shadow-xl max-h-[90vh] overflow-y-auto space-y-4">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <h3 className="font-bold text-lg text-foreground">
                {editingProduct ? t("products.editProduct") : t("products.createProduct")}
              </h3>
              <button
                onClick={() => setIsModalOpen(false)}
                className="p-1 rounded-lg text-muted-foreground hover:text-foreground hover:bg-secondary"
              >
                <X className="size-5" />
              </button>
            </div>

            <form onSubmit={handleSaveProduct} className="space-y-4">
              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">{t("products.form.title")}</label>
                <input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder={t("products.form.titlePlaceholder")}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                />
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">{t("products.form.sku")}</label>
                  <input
                    type="text"
                    value={formData.sku}
                    onChange={(e) => setFormData({ ...formData, sku: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground font-mono focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">{t("products.form.category")}</label>
                  <select
                    value={formData.category}
                    onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                  >
                    {categories.map((c) => (
                      <option key={c.id} value={c.name}>
                        {c.name}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">{t("products.form.price")}</label>
                  <input
                    type="number"
                    step="0.01"
                    required
                    value={formData.price}
                    onChange={(e) => setFormData({ ...formData, price: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">{t("products.form.comparePrice")}</label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.compareAtPrice}
                    onChange={(e) => setFormData({ ...formData, compareAtPrice: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">{t("products.form.stockQty")}</label>
                  <input
                    type="number"
                    required
                    value={formData.stock}
                    onChange={(e) => setFormData({ ...formData, stock: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                  />
                </div>
              </div>

              {editingProduct && (
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">{t("products.form.auditReason")}</label>
                  <select
                    value={formData.adjustmentReason || "adjustment"}
                    onChange={(e) => setFormData({ ...formData, adjustmentReason: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                  >
                    <option value="adjustment">{t("products.form.reasonAdjustment")}</option>
                    <option value="restock">{t("products.form.reasonRestock")}</option>
                    <option value="return">{t("products.form.reasonReturn")}</option>
                    <option value="sale">{t("products.form.reasonSale")}</option>
                  </select>
                  <p className="text-[11px] text-muted-foreground">{t("products.form.auditNote")}</p>
                </div>
              )}

              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">{t("products.form.imageUrl")}</label>
                <input
                  type="text"
                  value={formData.image}
                  onChange={(e) => setFormData({ ...formData, image: e.target.value })}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                />
              </div>

              <div className="flex justify-end gap-3 pt-3 border-t border-border">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-2 text-xs font-semibold rounded-xl border border-border hover:bg-secondary text-foreground"
                >
                  {t("actions.cancel")}
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-xs font-semibold rounded-xl bg-accent-brand text-white shadow-xs hover:bg-accent-brand/90"
                >
                  {t("products.form.saveProduct")}
                </button>
              </div>
            </form>

            {editingProduct && (
              <div className="border-t border-border pt-4 space-y-3">
                <div className="flex items-center justify-between gap-3">
                  <div className="flex flex-wrap items-center gap-2">
                    <Package className="size-4 text-accent-brand" />
                    <h4 className="text-sm font-bold text-foreground">{t("products.variants.title")}</h4>
                    {productDetail && (
                      <span className="rounded-full border border-border px-2 py-0.5 text-[10px] font-semibold text-muted-foreground">
                        {isSimpleProduct ? t("products.variants.simpleBadge") : t("products.variants.variableBadge")}
                      </span>
                    )}
                  </div>
                  {canAddVariant && (
                    <button
                      type="button"
                      onClick={openVariantCreate}
                      className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-border text-xs font-semibold text-foreground hover:bg-secondary disabled:opacity-50"
                    >
                      <Plus className="size-3.5" />
                      {t("products.variants.add")}
                    </button>
                  )}
                </div>

                {variantLocked && (
                  <p className="rounded-lg border border-border bg-secondary/30 px-3 py-2 text-[11px] text-muted-foreground">
                    {t("products.variants.simpleLocked")}
                  </p>
                )}

                {detailLoading ? (
                  <div className="flex items-center gap-2 py-4 text-xs text-muted-foreground">
                    <Loader2 className="size-4 animate-spin" />
                    {t("products.variants.loading")}
                  </div>
                ) : !productDetail?.variants?.length ? (
                  <p className="py-3 text-xs text-muted-foreground">{t("products.variants.empty")}</p>
                ) : (
                  <div className="space-y-2">
                    {productDetail.variants.map((variant: any) => (
                      <div key={variant.id} className="rounded-xl border border-border bg-secondary/20 p-3">
                        <div className="flex items-start justify-between gap-3">
                          <div className="min-w-0">
                            <div className="flex flex-wrap items-center gap-2">
                              <span className="truncate text-xs font-semibold text-foreground">{variant.title}</span>
                              {variant.isDefault && (
                                <span className="rounded-full bg-accent-brand/10 px-2 py-0.5 text-[10px] font-semibold text-accent-brand">
                                  {t("products.variants.defaultBadge")}
                                </span>
                              )}
                            </div>
                            <div className="mt-1 font-mono text-[10px] text-muted-foreground">
                              {variant.sku || "—"} · EGP {variant.extractedPrice ?? 0}
                            </div>
                            {variant.attributes && Object.keys(variant.attributes).length > 0 && (
                              <div className="mt-1.5 flex flex-wrap gap-1">
                                {Object.entries(variant.attributes).map(([key, value]) => (
                                  <span key={key} className="rounded-md bg-background px-1.5 py-0.5 text-[10px] text-muted-foreground border border-border">
                                    {key}: {String(value)}
                                  </span>
                                ))}
                              </div>
                            )}
                          </div>
                          <div className="flex shrink-0 items-center gap-1">
                            <span className="me-1 text-[10px] text-muted-foreground">
                              {t("products.units", { count: variant.stock?.availableQuantity ?? variant.stock?.onHandQuantity ?? 0 })}
                            </span>
                            {!variantLocked && !variant.isDefault && (
                              <button
                                type="button"
                                onClick={() => handleSetDefaultVariant(variant)}
                                className="rounded-lg p-1.5 text-muted-foreground hover:bg-secondary hover:text-accent-brand"
                                title={t("products.variants.setDefault")}
                              >
                                <Star className="size-4" />
                              </button>
                            )}
                            <button
                              type="button"
                              onClick={() => openVariantEdit(variant)}
                              className="rounded-lg p-1.5 text-muted-foreground hover:bg-secondary hover:text-accent-brand"
                              title={t("products.variants.edit")}
                            >
                              <Edit2 className="size-4" />
                            </button>
                            {!variantLocked && (
                              <button
                                type="button"
                                onClick={() => handleDeleteVariant(variant)}
                                className="rounded-lg p-1.5 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                                title={t("products.variants.delete")}
                              >
                                <Trash2 className="size-4" />
                              </button>
                            )}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}

                {variantForm && (
                  <div className="space-y-3 rounded-xl border border-accent-brand/40 bg-secondary/20 p-3">
                    <h5 className="text-xs font-bold text-foreground">
                      {variantForm.mode === "create" ? t("products.variants.newTitle") : t("products.variants.editTitle")}
                    </h5>

                    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                      <div className="space-y-1">
                        <label className="text-[11px] font-semibold text-foreground">{t("products.variants.fields.title")}</label>
                        <input
                          type="text"
                          value={variantForm.title}
                          onChange={(e) => setVariantForm({ ...variantForm, title: e.target.value })}
                          placeholder={t("products.variants.fields.titlePlaceholder")}
                          className="w-full rounded-lg border border-border bg-background px-2.5 py-1.5 text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                        />
                      </div>
                      {variantForm.mode === "edit" && (
                        <div className="space-y-1">
                          <label className="text-[11px] font-semibold text-foreground">{t("products.variants.fields.sku")}</label>
                          <input
                            type="text"
                            value={variantForm.sku}
                            onChange={(e) => setVariantForm({ ...variantForm, sku: e.target.value })}
                            className="w-full rounded-lg border border-border bg-background px-2.5 py-1.5 font-mono text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                          />
                        </div>
                      )}
                      <div className="space-y-1">
                        <label className="text-[11px] font-semibold text-foreground">{t("products.variants.fields.price")}</label>
                        <input
                          type="number"
                          step="0.01"
                          value={variantForm.price}
                          onChange={(e) => setVariantForm({ ...variantForm, price: e.target.value })}
                          className="w-full rounded-lg border border-border bg-background px-2.5 py-1.5 text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                        />
                      </div>
                      <div className="space-y-1">
                        <label className="text-[11px] font-semibold text-foreground">{t("products.variants.fields.originalPrice")}</label>
                        <input
                          type="number"
                          step="0.01"
                          value={variantForm.originalPrice}
                          onChange={(e) => setVariantForm({ ...variantForm, originalPrice: e.target.value })}
                          className="w-full rounded-lg border border-border bg-background px-2.5 py-1.5 text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                        />
                      </div>
                      <div className="space-y-1">
                        <label className="text-[11px] font-semibold text-foreground">{t("products.variants.fields.currency")}</label>
                        <input
                          type="text"
                          value={variantForm.currency}
                          onChange={(e) => setVariantForm({ ...variantForm, currency: e.target.value.toUpperCase() })}
                          className="w-full rounded-lg border border-border bg-background px-2.5 py-1.5 text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                        />
                      </div>
                      {variantForm.mode === "create" && (
                        <div className="space-y-1">
                          <label className="text-[11px] font-semibold text-foreground">{t("products.variants.fields.initialStock")}</label>
                          <input
                            type="number"
                            min="0"
                            value={variantForm.initialStock}
                            onChange={(e) => setVariantForm({ ...variantForm, initialStock: e.target.value })}
                            className="w-full rounded-lg border border-border bg-background px-2.5 py-1.5 text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                          />
                        </div>
                      )}
                    </div>

                    {productCategoryAttributes.length > 0 ? (
                      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                        {productCategoryAttributes.map((attribute: any) => {
                          const value = variantForm.attributes?.[attribute.key] ?? "";
                          const setValue = (next: any) =>
                            setVariantForm({
                              ...variantForm,
                              attributes: { ...variantForm.attributes, [attribute.key]: next },
                            });
                          return (
                            <div key={attribute.id} className="space-y-1">
                              <label className="text-[11px] font-semibold text-foreground">{attribute.label}</label>
                              {attribute.valueType === "boolean" ? (
                                <label className="flex items-center gap-2 rounded-lg border border-border bg-background px-2.5 py-2 text-xs text-foreground">
                                  <input
                                    type="checkbox"
                                    checked={Boolean(value)}
                                    onChange={(e) => setValue(e.target.checked)}
                                    className="size-4 accent-accent-brand"
                                  />
                                  {attribute.label}
                                </label>
                              ) : attribute.valueType === "enum" && Array.isArray(attribute.allowedValues) ? (
                                <select
                                  value={String(value)}
                                  onChange={(e) => setValue(e.target.value)}
                                  className="w-full rounded-lg border border-border bg-background px-2.5 py-1.5 text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                                >
                                  <option value="">—</option>
                                  {attribute.allowedValues.map((option: string) => (
                                    <option key={option} value={option}>{option}</option>
                                  ))}
                                </select>
                              ) : (
                                <input
                                  type="text"
                                  value={String(value)}
                                  onChange={(e) => setValue(e.target.value)}
                                  className="w-full rounded-lg border border-border bg-background px-2.5 py-1.5 text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                                />
                              )}
                            </div>
                          );
                        })}
                      </div>
                    ) : (
                      <p className="text-[11px] text-muted-foreground">{t("products.variants.noAttributes")}</p>
                    )}

                    <div className="flex justify-end gap-2">
                      <button
                        type="button"
                        onClick={() => setVariantForm(null)}
                        className="rounded-lg border border-border px-3 py-1.5 text-xs font-semibold text-foreground hover:bg-secondary"
                      >
                        {t("actions.cancel")}
                      </button>
                      <button
                        type="button"
                        onClick={handleSaveVariant}
                        disabled={savingVariant}
                        className="flex items-center gap-1.5 rounded-lg bg-accent-brand px-3 py-1.5 text-xs font-semibold text-white hover:bg-accent-brand/90 disabled:opacity-60"
                      >
                        {savingVariant && <Loader2 className="size-3.5 animate-spin" />}
                        {t("products.variants.save")}
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      )}
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
                  <p className="text-xs text-muted-foreground">{auditProduct.name} &bull; <span className="font-mono">{auditProduct.sku}</span></p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="text-end">
                  <div className="text-[10px] text-muted-foreground uppercase font-semibold">{t("products.audit.currentStock")}</div>
                  <div className="text-xs font-bold text-accent-brand">{t("products.units", { count: auditProduct.stock })}</div>
                </div>
                <button
                  onClick={() => setAuditProduct(null)}
                  className="text-muted-foreground hover:text-foreground text-sm font-semibold p-1"
                >
                  ✕
                </button>
              </div>
            </div>

            <div className="flex-1 overflow-y-auto space-y-2 pe-1 min-h-[220px]">
              {loadingAudit ? (
                <div className="py-12 text-center text-xs text-muted-foreground">{t("products.audit.loading")}</div>
              ) : auditLogs.length === 0 ? (
                <div className="py-12 text-center text-xs text-muted-foreground">
                  {t("products.audit.empty")}
                </div>
              ) : (
                auditLogs.map((entry) => (
                  <div key={entry.id} className="p-3 border border-border rounded-xl bg-secondary/15 flex items-center justify-between text-xs">
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
