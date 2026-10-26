import { useState, useEffect } from "react";

import { Search, Plus, Edit2, Trash2, X } from "lucide-react";
import {
  INITIAL_VENDOR_PRODUCTS,
  getLocalVendorData,
  setLocalVendorData,
  type VendorProduct
} from "../data/mockVendorData";
import { catalogService } from "@/modules/catalog/api/catalogService";
import { useOptionalCatalogContext } from "@/modules/catalog/model/catalogContext";
import { notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";

import { useVendorContext } from "./vendorContext";

interface VendorProductRaw {
  id: string;
  vendorId?: string;
  vendorName?: string;
  sku?: string;
  name?: string;
  product?: { en?: string; ar?: string } | string;
  title?: string;
  category?: string | { name?: string };
  price?: string | number;
  cost?: number;
  stock?: number;
  stockCount?: number;
  status?: string;
  salesCount?: number;
  img?: string;
  image?: string;
}

export function VendorProducts() {
  const t = useTranslations("vendor");
  const { vendor } = useVendorContext();
  const { refreshCatalog } = useOptionalCatalogContext() || {};

  const [products, setProducts] = useState<VendorProduct[]>(() =>
    getLocalVendorData("products", INITIAL_VENDOR_PRODUCTS)
  );
  const [search, setSearch] = useState("");
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingProduct, setEditingProduct] = useState<VendorProduct | null>(null);

  useEffect(() => {
    async function loadCatalog() {
      try {
        const res = await catalogService.getProducts();
        const list: VendorProductRaw[] = res?.data ?? [];
        if (list.length > 0) {
          // Map backend products to vendor products format if needed
          const serverProducts = list
            .filter((p: VendorProductRaw) => !p.vendorId || p.vendorId === (vendor?.id || "v-1"))
            .map((p: VendorProductRaw) => ({
              id: p.id,
              vendorId: p.vendorId || vendor?.id || "v-1",
              vendorName: p.vendorName || vendor?.name || "SoundMaster Studio",
              name: (typeof p.product === "object" ? p.product?.en : undefined) || p.name || p.title || "Product",
              sku: p.sku || `SKU-${p.id}`,
              category: typeof p.category === "string" ? p.category : "Electronics",
              price: typeof p.price === "number" ? p.price : parseFloat(String(p.price).replace(/[^0-9.]/g, "")) || 0,
              cost: p.cost ? parseFloat(String(p.cost)) : 45.00,
              stock: p.stock !== undefined ? p.stock : (p.stockCount ?? 20),
              status: p.status || ((p.stock ?? 0) > 0 ? "Active" : "Out of Stock"),
              salesCount: p.salesCount || Math.floor(Math.random() * 50 + 10),
              image: p.img || p.image || "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400&q=80"
            }));

          if (serverProducts.length > 0) {
            setProducts(serverProducts);
            setLocalVendorData("products", serverProducts);
          }
        }
      } catch {
        // Retain local state
      }
    }
    loadCatalog();
  }, [vendor?.id, vendor?.name]);

  const [formData, setFormData] = useState({
    name: "",
    sku: "",
    category: "Electronics",
    price: "",
    cost: "",
    stock: "",
    image: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400&q=80"
  });

  const saveProducts = (updated: VendorProduct[]) => {
    setProducts(updated);
    setLocalVendorData("products", updated);
  };

  const statusLabel = (status: string) => {
    const map: Record<string, string> = {
      Active: "status.active",
      "Low Stock": "status.lowStock",
      "Out of Stock": "status.outOfStock"
    };
    return map[status] ? t(map[status]) : status;
  };

  const handleOpenAdd = () => {
    setEditingProduct(null);
    setFormData({
      name: "",
      sku: `SM-SKU-${Math.floor(Math.random() * 900 + 100)}`,
      category: "Audio",
      price: "129.99",
      cost: "60.00",
      stock: "25",
      image: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400&q=80"
    });
    setIsModalOpen(true);
  };

  const handleOpenEdit = (prod: VendorProduct) => {
    setEditingProduct(prod);
    setFormData({
      name: prod.name,
      sku: prod.sku,
      category: prod.category,
      price: prod.price.toString(),
      cost: prod.cost ? prod.cost.toString() : "50",
      stock: prod.stock.toString(),
      image: prod.image
    });
    setIsModalOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name.trim() || !formData.price) return;

    const priceNum = parseFloat(formData.price) || 0;
    const stockNum = parseInt(formData.stock, 10) || 0;
    const costNum = parseFloat(formData.cost) || 0;

    const statusVal = stockNum === 0 ? "Out of Stock" : stockNum < 10 ? "Low Stock" : "Active";

    if (editingProduct) {
      const updatedProductData = {
        name: formData.name,
        sku: formData.sku,
        category: formData.category,
        price: priceNum,
        cost: costNum,
        stock: stockNum,
        status: statusVal,
        image: formData.image
      };

      const updated = products.map((p: VendorProduct) =>
        p.id === editingProduct.id
          ? {
              ...p,
              ...updatedProductData
            }
          : p
      );
      saveProducts(updated);

      try {
        await catalogService.updateProduct(editingProduct.id, updatedProductData);
        refreshCatalog?.();
      } catch {
        // Mock fallback
      }

      notify.success(t("toasts.productUpdated"));
    } else {
      const newProd = {
        id: `vprod-${Date.now()}`,
        vendorId: vendor?.id || "v-1",
        vendorName: vendor?.name || "SoundMaster Studio",
        name: formData.name,
        sku: formData.sku,
        category: formData.category,
        price: priceNum,
        cost: costNum,
        stock: stockNum,
        status: statusVal,
        salesCount: 0,
        image: formData.image
      };
      saveProducts([newProd, ...products]);

      try {
        await catalogService.createProduct({
          ...newProd,
          product: { en: newProd.name, ar: newProd.name }
        });
        refreshCatalog?.();
      } catch {
        // Mock fallback
      }

      notify.success(t("toasts.productAdded"));
    }

    setIsModalOpen(false);
  };

  const handleDelete = async (id: string) => {
    if (confirm(t("products.confirmDelete"))) {
      saveProducts(products.filter((p: VendorProduct) => p.id !== id));
      try {
        await catalogService.deleteProduct(id);
        refreshCatalog?.();
      } catch {
        // Mock fallback
      }
      notify.success(t("toasts.productRemoved"));
    }
  };

  const filtered = products.filter(
    (p: VendorProduct) =>
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.sku.toLowerCase().includes(search.toLowerCase()) ||
      p.category.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-6 w-full">
      {/* Top Banner */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("products.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t("products.subtitle")}
          </p>
        </div>

        <button
          onClick={handleOpenAdd}
          className="flex items-center gap-2 px-4 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0"
        >
          <Plus className="size-4" />
          <span>{t("actions.addNewProduct")}</span>
        </button>
      </div>

      {/* Search Bar */}
      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs">
        <div className="relative max-w-sm">
          <Search className="absolute left-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder={t("products.searchPlaceholder")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-accent-brand"
          />
        </div>
      </div>

      {/* Product Table */}
      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <table className="w-full min-w-max text-left text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 pl-4">{t("products.table.details")}</th>
                <th className="py-3.5">{t("products.table.sku")}</th>
                <th className="py-3.5">{t("products.table.category")}</th>
                <th className="py-3.5">{t("products.table.priceCost")}</th>
                <th className="py-3.5">{t("products.table.stock")}</th>
                <th className="py-3.5">{t("products.table.status")}</th>
                <th className="py-3.5 pr-4 text-right">{t("products.table.actions")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {filtered.map((prod: VendorProduct) => (
                <tr key={prod.id} className="hover:bg-secondary/20 transition-colors">
                  <td className="py-3.5 pl-4">
                    <div className="flex items-center gap-3">
                      <img src={prod.image} alt={prod.name} className="size-10 rounded-xl object-cover border border-border shrink-0" />
                      <div>
                        <div className="font-semibold text-foreground text-xs">{prod.name}</div>
                        <div className="text-[10px] text-muted-foreground">{t("products.totalSales", { count: prod.salesCount })}</div>
                      </div>
                    </div>
                  </td>
                  <td className="py-3.5 font-mono text-muted-foreground">{prod.sku}</td>
                  <td className="py-3.5 text-muted-foreground">{prod.category}</td>
                  <td className="py-3.5">
                    <div className="font-bold text-foreground">EGP {prod.price.toFixed(2)}</div>
                    <div className="text-[10px] text-muted-foreground">{t("products.cost", { amount: prod.cost ? prod.cost.toFixed(2) : "0.00" })}</div>
                  </td>
                  <td className="py-3.5 font-bold text-foreground">{t("products.units", { count: prod.stock })}</td>
                  <td className="py-3.5">
                    <span
                      className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${
                        prod.status === "Active"
                          ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400"
                          : prod.status === "Low Stock"
                          ? "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-400"
                          : "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-400"
                      }`}
                    >
                      {statusLabel(prod.status)}
                    </span>
                  </td>
                  <td className="py-3.5 pr-4 text-right space-x-1">
                    <button onClick={() => handleOpenEdit(prod)} className="p-1.5 text-muted-foreground hover:text-accent-brand">
                      <Edit2 className="size-4" />
                    </button>
                    <button onClick={() => handleDelete(prod.id)} className="p-1.5 text-muted-foreground hover:text-destructive">
                      <Trash2 className="size-4" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Product Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-md w-full p-6 shadow-xl space-y-4">
            <div className="flex justify-between items-center border-b border-border pb-3">
              <h3 className="font-bold text-base text-foreground">
                {editingProduct ? t("products.editTitle") : t("products.addTitle")}
              </h3>
              <button onClick={() => setIsModalOpen(false)} className="text-muted-foreground">
                <X className="size-5" />
              </button>
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">{t("products.form.title")}</label>
                <input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e: React.ChangeEvent<HTMLInputElement>) => setFormData({ ...formData, name: e.target.value })}
                  placeholder={t("products.form.titlePlaceholder")}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">{t("products.form.skuCode")}</label>
                  <input
                    type="text"
                    value={formData.sku}
                    onChange={(e) => setFormData({ ...formData, sku: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs font-mono text-foreground"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">{t("products.form.category")}</label>
                  <select
                    value={formData.category}
                    onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                  >
                    <option value="Audio">{t("products.categories.audio")}</option>
                    <option value="Electronics">{t("products.categories.electronics")}</option>
                    <option value="Accessories">{t("products.categories.accessories")}</option>
                    <option value="Keyboards">{t("products.categories.keyboards")}</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-3 gap-3">
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">{t("products.form.sellingPrice")}</label>
                  <input
                    type="number"
                    step="0.01"
                    required
                    value={formData.price}
                    onChange={(e) => setFormData({ ...formData, price: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">{t("products.form.unitCost")}</label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.cost}
                    onChange={(e) => setFormData({ ...formData, cost: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">{t("products.form.initialStock")}</label>
                  <input
                    type="number"
                    required
                    value={formData.stock}
                    onChange={(e) => setFormData({ ...formData, stock: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                  />
                </div>
              </div>

              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">{t("products.form.imageUrl")}</label>
                <input
                  type="text"
                  value={formData.image}
                  onChange={(e) => setFormData({ ...formData, image: e.target.value })}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                />
              </div>

              <div className="flex justify-end gap-3 pt-3 border-t border-border">
                <button type="button" onClick={() => setIsModalOpen(false)} className="px-4 py-2 text-xs border border-border rounded-xl">
                  {t("actions.cancel")}
                </button>
                <button type="submit" className="px-4 py-2 text-xs bg-accent-brand text-white font-semibold rounded-xl">
                  {editingProduct ? t("actions.saveChanges") : t("actions.createProduct")}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
