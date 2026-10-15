import { useState, useEffect } from "react";
import {
  Plus,
  Search,
  Filter,
  Edit2,
  Trash2,
  History,
  X
} from "lucide-react";
import { catalogService } from "@/services/catalog";
import { adminService } from "@/services/admin";
import { useCatalogContext } from "@/context/CatalogContext";
import { toast } from "sonner";

export function AdminProducts() {
  const [products, setProducts] = useState<any[]>([]);
  const [categories, setCategories] = useState<any[]>([]);
  const [brands, setBrands] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const { refreshCatalog } = useCatalogContext() || {};

  const loadAll = async () => {
    try {
      setLoading(true);
      const [prodsRes, catsRes, brandsRes]: [any, any, any] = await Promise.all([
        catalogService.getProducts(`pageSize=100&_t=${Date.now()}`),
        catalogService.getCategories(100),
        catalogService.getBrands(100)
      ]);

      const catList = Array.isArray(catsRes) ? catsRes : (Array.isArray(catsRes?.data) ? catsRes.data : []);
      setCategories(catList);

      const brandList = Array.isArray(brandsRes) ? brandsRes : (Array.isArray(brandsRes?.data) ? brandsRes.data : []);
      setBrands(brandList);

      const list = Array.isArray(prodsRes) ? prodsRes : (Array.isArray(prodsRes?.data) ? prodsRes.data : []);
      if (list.length > 0) {
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
      }
    } catch {
      toast.error("Failed to load catalog products");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAll();
  }, []);

  const [search, setSearch] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("All");
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
        limit: 50,
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
      toast.error("Failed to load audit history for product");
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

  const handleOpenEditModal = (prod) => {
    setEditingProduct(prod);
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
  };

  const handleSaveProduct = async (e) => {
    e.preventDefault();
    if (!formData.name || !formData.price || !formData.stock) {
      toast.error("Please fill in required fields");
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
          toast.success(`Product updated & ledger recorded (${stockDiff > 0 ? `+${stockDiff}` : stockDiff} units)`);
        } else {
          toast.success("Product updated successfully!");
        }
        await loadAll();
        refreshCatalog?.();
      } catch (err: any) {
        toast.error(err?.message || "Product update partially failed");
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

      toast.success("New product created and added to catalog!");
    }
  };

  const handleDeleteProduct = async (id) => {
    if (confirm("Are you sure you want to delete this product?")) {
      const updated = products.filter((p) => p.id !== id);
      saveProductsToState(updated);
      try {
        await catalogService.deleteProduct(id);
        refreshCatalog?.();
      } catch {
        // Ignored in mock
      }
      toast.success("Product removed from catalog");
    }
  };

  const filteredProducts = products.filter((prod) => {
    const matchesSearch =
      prod.name.toLowerCase().includes(search.toLowerCase()) ||
      prod.sku.toLowerCase().includes(search.toLowerCase());
    const matchesCategory =
      selectedCategory === "All" || prod.category === selectedCategory;
    return matchesSearch && matchesCategory;
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Products & Variants</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Manage your store catalog, pricing, inventory stock, and product variants.
          </p>
        </div>
        <button
          onClick={handleOpenCreateModal}
          className="flex items-center gap-2 px-4 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0"
        >
          <Plus className="size-4" />
          <span>Add New Product</span>
        </button>
      </div>

      {/* Filters Bar */}
      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs flex flex-col md:flex-row gap-4 justify-between items-center">
        <div className="relative w-full md:w-80">
          <Search className="absolute left-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search by product name or SKU..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground placeholder:text-muted-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
          />
        </div>

        <div className="flex items-center gap-3 w-full md:w-auto">
          <Filter className="size-4 text-muted-foreground" />
          <select
            value={selectedCategory}
            onChange={(e) => setSelectedCategory(e.target.value)}
            className="px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
          >
            <option value="All">All Categories</option>
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
          <table className="w-full text-left text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 pl-4">Product</th>
                <th className="py-3.5">SKU</th>
                <th className="py-3.5">Category</th>
                <th className="py-3.5">Price</th>
                <th className="py-3.5">Stock</th>
                <th className="py-3.5">Status</th>
                <th className="py-3.5 pr-4 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {filteredProducts.length > 0 ? (
                filteredProducts.map((prod) => (
                  <tr key={prod.id} className="hover:bg-secondary/20 transition-colors">
                    <td className="py-3.5 pl-4">
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
                            Brand: {typeof prod.brand === "object" ? prod.brand?.name : String(prod.brand || "PRIM")}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="py-3.5 font-mono text-muted-foreground font-medium">{prod.sku}</td>
                    <td className="py-3.5 font-medium text-foreground">
                      {typeof prod.category === "object" ? prod.category?.name : String(prod.category || "General")}
                    </td>
                    <td className="py-3.5">
                      <div className="font-bold text-foreground">${prod.price.toFixed(2)}</div>
                      {prod.compareAtPrice && (
                        <div className="text-[10px] text-muted-foreground line-through">
                          ${prod.compareAtPrice.toFixed(2)}
                        </div>
                      )}
                    </td>
                    <td className="py-3.5">
                      <span className={`font-bold ${prod.stock <= 10 ? "text-amber-600 dark:text-amber-400" : "text-foreground"}`}>
                        {prod.stock} units
                      </span>
                    </td>
                    <td className="py-3.5">
                      <span
                        className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${
                          prod.status === "Active"
                            ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400"
                            : "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-400"
                        }`}
                      >
                        {prod.status}
                      </span>
                    </td>
                    <td className="py-3.5 pr-4 text-right">
                      <div className="flex items-center justify-end gap-1.5">
                        <button
                          onClick={() => handleViewAudit(prod)}
                          className="p-1.5 text-muted-foreground hover:text-accent-brand hover:bg-secondary rounded-lg transition-colors"
                          title="View Audit Trail"
                        >
                          <History className="size-4" />
                        </button>
                        <button
                          onClick={() => handleOpenEditModal(prod)}
                          className="p-1.5 text-muted-foreground hover:text-accent-brand hover:bg-secondary rounded-lg transition-colors"
                          title="Edit Product"
                        >
                          <Edit2 className="size-4" />
                        </button>
                        <button
                          onClick={() => handleDeleteProduct(prod.id)}
                          className="p-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
                          title="Delete Product"
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
                    No products found matching criteria.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {/* Create / Edit Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-xl w-full p-6 shadow-xl max-h-[90vh] overflow-y-auto space-y-4">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <h3 className="font-bold text-lg text-foreground">
                {editingProduct ? "Edit Product" : "Create New Product"}
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
                <label className="text-xs font-semibold text-foreground">Product Title *</label>
                <input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="e.g. Wireless Noise-Canceling Headphones"
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">SKU Code</label>
                  <input
                    type="text"
                    value={formData.sku}
                    onChange={(e) => setFormData({ ...formData, sku: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground font-mono focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">Category</label>
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

              <div className="grid grid-cols-3 gap-4">
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">Price ($) *</label>
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
                  <label className="text-xs font-semibold text-foreground">Compare Price ($)</label>
                  <input
                    type="number"
                    step="0.01"
                    value={formData.compareAtPrice}
                    onChange={(e) => setFormData({ ...formData, compareAtPrice: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-foreground">Stock Qty *</label>
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
                  <label className="text-xs font-semibold text-foreground">Stock Ledger Audit Reason</label>
                  <select
                    value={formData.adjustmentReason || "adjustment"}
                    onChange={(e) => setFormData({ ...formData, adjustmentReason: e.target.value })}
                    className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30"
                  >
                    <option value="adjustment">Manual Adjustment (Correction / Audit)</option>
                    <option value="restock">Restock (Received Inventory)</option>
                    <option value="return">Customer Return</option>
                    <option value="sale">Sale Deduction</option>
                  </select>
                  <p className="text-[11px] text-muted-foreground">Changes to stock quantity will automatically append a new record to the product's inventory audit ledger.</p>
                </div>
              )}

              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">Image URL</label>
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
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-xs font-semibold rounded-xl bg-accent-brand text-white shadow-xs hover:bg-accent-brand/90"
                >
                  Save Product
                </button>
              </div>
            </form>
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
                  <h3 className="font-bold text-base text-foreground">Product Audit Trail</h3>
                  <p className="text-xs text-muted-foreground">{auditProduct.name} &bull; <span className="font-mono">{auditProduct.sku}</span></p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="text-right">
                  <div className="text-[10px] text-muted-foreground uppercase font-semibold">Current On-Hand Stock</div>
                  <div className="text-xs font-bold text-accent-brand">{auditProduct.stock} units</div>
                </div>
                <button
                  onClick={() => setAuditProduct(null)}
                  className="text-muted-foreground hover:text-foreground text-sm font-semibold p-1"
                >
                  ✕
                </button>
              </div>
            </div>

            <div className="flex-1 overflow-y-auto space-y-2 pr-1 min-h-[220px]">
              {loadingAudit ? (
                <div className="py-12 text-center text-xs text-muted-foreground">Loading audit records...</div>
              ) : auditLogs.length === 0 ? (
                <div className="py-12 text-center text-xs text-muted-foreground">
                  No inventory ledger events recorded for this product yet.
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
                    <div className="text-right">
                      <span className={`text-sm font-bold ${entry.quantity > 0 ? "text-emerald-600 dark:text-emerald-400" : "text-amber-600 dark:text-amber-400"}`}>
                        {entry.change} units
                      </span>
                    </div>
                  </div>
                ))
              )}
            </div>

            <div className="flex items-center justify-between pt-3 border-t border-border">
              <div className="text-xs text-muted-foreground">
                Total Ledger Balance:{" "}
                <span className="font-bold text-foreground">
                  {auditLogs.reduce((acc, curr) => acc + (curr.quantity || 0), 0)} units
                </span>
              </div>
              <button
                onClick={() => setAuditProduct(null)}
                className="px-4 py-2 text-xs font-semibold bg-secondary hover:bg-secondary/80 text-foreground border border-border rounded-xl"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
