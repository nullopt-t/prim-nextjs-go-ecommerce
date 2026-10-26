import { useState, useEffect } from "react";

import { Plus, Minus, Search, AlertTriangle } from "lucide-react";
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

export function VendorInventory() {
  const t = useTranslations("vendor");
  const { vendor } = useVendorContext();
  const { refreshCatalog } = useOptionalCatalogContext() || {};

  const [products, setProducts] = useState<VendorProduct[]>(() =>
    getLocalVendorData("products", INITIAL_VENDOR_PRODUCTS)
  );
  const [search, setSearch] = useState("");
  const [filterLowStock, setFilterLowStock] = useState(false);

  useEffect(() => {
    async function loadCatalog() {
      try {
        const res = await catalogService.getProducts();
        const list: VendorProductRaw[] = res?.data ?? [];
        if (list.length > 0) {
          const serverProducts = list
            .filter((p: VendorProductRaw) => !p.vendorId || p.vendorId === (vendor?.id || "v-1"))
            .map((p: VendorProductRaw) => ({
              id: p.id,
              vendorId: p.vendorId || vendor?.id || "v-1",
              vendorName: p.vendorName || vendor?.name || "SoundMaster Studio",
              name: (typeof p.product === "object" ? p.product?.en : undefined) || p.name || p.title || "Product",
              sku: p.sku || `SKU-${p.id}`,
              category: typeof p.category === "string" ? p.category : "Electronics",
              price: typeof p.price === "number" ? p.price : parseFloat(String(p.price || "0").replace(/[^0-9.]/g, "")) || 0,
              cost: p.cost ? parseFloat(String(p.cost)) : 45.00,
              stock: p.stock !== undefined ? p.stock : (p.stockCount ?? 20),
              status: p.status || ((p.stock ?? 20) > 0 ? "Active" : "Out of Stock"),
              salesCount: p.salesCount || Math.floor(Math.random() * 50 + 10),
              image: p.img || p.image || "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400&q=80"
            }));

          if (serverProducts.length > 0) {
            setProducts(serverProducts);
            setLocalVendorData("products", serverProducts);
          }
        }
      } catch {
        // Fallback to local
      }
    }
    loadCatalog();
  }, [vendor?.id]);

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

  const syncStockToApi = async (id: string, newStock: number, newStatus: string) => {
    try {
      await catalogService.updateProduct(id, { stock: newStock, status: newStatus });
      refreshCatalog?.();
    } catch {
      // Ignored for mock
    }
  };

  const handleAdjustStock = (id: string, delta: number) => {
    let updatedStock = 0;
    let updatedStatus = "Active";

    const updated = products.map((p: VendorProduct) => {
      if (p.id === id) {
        const newStock = Math.max(0, (p.stock || 0) + delta);
        const newStatus = newStock === 0 ? "Out of Stock" : newStock < 10 ? "Low Stock" : "Active";
        updatedStock = newStock;
        updatedStatus = newStatus;
        return { ...p, stock: newStock, status: newStatus };
      }
      return p;
    });
    saveProducts(updated);
    syncStockToApi(id, updatedStock, updatedStatus);
    notify.success(t("toasts.inventoryUpdated"));
  };

  const handleSetExactStock = (id: string, newQty: string) => {
    const parsed = parseInt(newQty, 10);
    if (isNaN(parsed) || parsed < 0) return;

    const newStatus = parsed === 0 ? "Out of Stock" : parsed < 10 ? "Low Stock" : "Active";

    const updated = products.map((p: VendorProduct) => {
      if (p.id === id) {
        return { ...p, stock: parsed, status: newStatus };
      }
      return p;
    });
    saveProducts(updated);
    syncStockToApi(id, parsed, newStatus);
  };

  const filtered = products.filter((p: VendorProduct) => {
    const matchesSearch =
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.sku.toLowerCase().includes(search.toLowerCase());
    const matchesLowStock = filterLowStock ? p.stock < 10 : true;
    return matchesSearch && matchesLowStock;
  });

  const lowStockCount = products.filter((p: VendorProduct) => p.stock < 10).length;

  return (
    <div className="space-y-6 w-full">
      {/* Top Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("inventory.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t("inventory.subtitle")}
          </p>
        </div>

        {lowStockCount > 0 && (
          <div className="flex items-center gap-2 px-3.5 py-2 bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20 rounded-xl text-xs font-semibold">
            <AlertTriangle className="size-4 shrink-0" />
            <span>{t("inventory.lowStockWarning", { count: lowStockCount })}</span>
          </div>
        )}
      </div>

      {/* Controls */}
      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="relative max-w-sm flex-1">
          <Search className="absolute left-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder={t("inventory.searchPlaceholder")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
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

      {/* Inventory Table */}
      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <table className="w-full min-w-max text-left text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 pl-4">{t("inventory.table.productName")}</th>
                <th className="py-3.5">{t("inventory.table.sku")}</th>
                <th className="py-3.5">{t("inventory.table.reorderPoint")}</th>
                <th className="py-3.5">{t("inventory.table.currentStock")}</th>
                <th className="py-3.5">{t("inventory.table.status")}</th>
                <th className="py-3.5 pr-4 text-right">{t("inventory.table.adjustment")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {filtered.map((prod: VendorProduct) => (
                <tr key={prod.id} className="hover:bg-secondary/20 transition-colors">
                  <td className="py-3.5 pl-4">
                    <div className="font-semibold text-foreground">{prod.name}</div>
                    <div className="text-[10px] text-muted-foreground">{prod.category}</div>
                  </td>
                  <td className="py-3.5 font-mono text-muted-foreground">{prod.sku}</td>
                  <td className="py-3.5 text-muted-foreground">{t("inventory.reorderPointValue", { count: 10 })}</td>
                  <td className="py-3.5 font-bold text-base text-foreground">{prod.stock}</td>
                  <td className="py-3.5">
                    <span
                      className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${
                        prod.stock === 0
                          ? "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-400"
                          : prod.stock < 10
                          ? "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-400"
                          : "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400"
                      }`}
                    >
                      {statusLabel(prod.status)}
                    </span>
                  </td>
                  <td className="py-3.5 pr-4 text-right">
                    <div className="inline-flex items-center gap-1.5 bg-secondary/40 border border-border p-1 rounded-xl">
                      <button
                        onClick={() => handleAdjustStock(prod.id, -1)}
                        className="p-1 hover:bg-background rounded-lg text-foreground transition-colors"
                        title={t("inventory.decrease")}
                      >
                        <Minus className="size-3.5" />
                      </button>

                      <input
                        type="number"
                        value={prod.stock}
                        onChange={(e) => handleSetExactStock(prod.id, e.target.value)}
                        className="w-12 text-center bg-transparent font-bold text-xs text-foreground focus:outline-none"
                      />

                      <button
                        onClick={() => handleAdjustStock(prod.id, 1)}
                        className="p-1 hover:bg-background rounded-lg text-foreground transition-colors"
                        title={t("inventory.increase")}
                      >
                        <Plus className="size-3.5" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
