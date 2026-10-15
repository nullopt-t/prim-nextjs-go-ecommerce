import { useState, useEffect } from "react";
import { Plus, Search, Edit2, Trash2, FolderTree, X } from "lucide-react";
import { catalogService } from "@/services/catalog";
import { toast } from "sonner";

export function AdminCategories() {
  const [categories, setCategories] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingCategory, setEditingCategory] = useState<any>(null);

  const loadCategories = async () => {
    try {
      setLoading(true);
      const res: any = await catalogService.getCategories(100);
      const list = Array.isArray(res) ? res : (Array.isArray(res?.data) ? res.data : []);
      setCategories(list.map((c: any) => ({
        id: c.id,
        name: c.name,
        slug: c.slug || c.name.toLowerCase().replace(/\s+/g, "-"),
        parent: c.parentName || "None",
        parentId: c.parentId,
        productsCount: c.productCount || 0,
        status: "Active",
        icon: "FolderTree"
      })));
    } catch {
      toast.error("Failed to load categories from server");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadCategories();
  }, []);

  const [formData, setFormData] = useState({
    name: "",
    slug: "",
    parentId: "",
    status: "Active"
  });

  const handleOpenModal = (cat: any = null) => {
    if (cat) {
      setEditingCategory(cat);
      setFormData({ name: cat.name, slug: cat.slug, parentId: cat.parentId || "", status: cat.status });
    } else {
      setEditingCategory(null);
      setFormData({ name: "", slug: "", parentId: "", status: "Active" });
    }
    setIsModalOpen(true);
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name) return;

    try {
      if (editingCategory) {
        await catalogService.updateCategory(editingCategory.id, {
          name: formData.name,
          parentId: formData.parentId ? formData.parentId : undefined,
        });
        toast.success("Category updated!");
      } else {
        await catalogService.createCategory({
          name: formData.name,
          parentId: formData.parentId ? formData.parentId : undefined,
        });
        toast.success("New category added!");
      }
      setIsModalOpen(false);
      loadCategories();
    } catch (err: any) {
      toast.error(err?.response?.data?.message || "Failed to save category");
    }
  };

  const handleDelete = (id: string) => {
    toast.info("Categories with attached products cannot be deleted directly.");
  };

  const filtered = categories.filter((c) =>
    c.name.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Categories</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Organize products into hierarchical categories and subcategories.
          </p>
        </div>
        <button
          onClick={() => handleOpenModal()}
          className="flex items-center gap-2 px-4 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0"
        >
          <Plus className="size-4" />
          <span>Add Category</span>
        </button>
      </div>

      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs">
        <div className="relative max-w-sm">
          <Search className="absolute left-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search categories..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground placeholder:text-muted-foreground focus:outline-hidden"
          />
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
        <table className="w-full text-left text-xs">
          <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
            <tr>
              <th className="py-3.5 pl-4">Category Name</th>
              <th className="py-3.5">Slug</th>
              <th className="py-3.5">Parent Category</th>
              <th className="py-3.5">Products Count</th>
              <th className="py-3.5">Status</th>
              <th className="py-3.5 pr-4 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/60">
            {filtered.map((cat) => (
              <tr key={cat.id} className="hover:bg-secondary/20 transition-colors">
                <td className="py-3.5 pl-4 font-semibold text-foreground flex items-center gap-2">
                  <FolderTree className="size-4 text-accent-brand" />
                  <span>{cat.name}</span>
                </td>
                <td className="py-3.5 font-mono text-muted-foreground">{cat.slug}</td>
                <td className="py-3.5 text-muted-foreground">{cat.parent}</td>
                <td className="py-3.5 font-bold text-foreground">{cat.productsCount} items</td>
                <td className="py-3.5">
                  <span className="px-2 py-0.5 bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400 rounded-full text-[10px] font-bold">
                    {cat.status}
                  </span>
                </td>
                <td className="py-3.5 pr-4 text-right">
                  <div className="flex items-center justify-end gap-2">
                    <button onClick={() => handleOpenModal(cat)} className="p-1.5 text-muted-foreground hover:text-accent-brand rounded-lg">
                      <Edit2 className="size-4" />
                    </button>
                    <button onClick={() => handleDelete(cat.id)} className="p-1.5 text-muted-foreground hover:text-destructive rounded-lg">
                      <Trash2 className="size-4" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-md w-full p-6 shadow-xl space-y-4">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <h3 className="font-bold text-base text-foreground">
                {editingCategory ? "Edit Category" : "Add New Category"}
              </h3>
              <button onClick={() => setIsModalOpen(false)} className="text-muted-foreground hover:text-foreground">
                <X className="size-5" />
              </button>
            </div>
            <form onSubmit={handleSave} className="space-y-4">
              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">Category Name *</label>
                <input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                />
              </div>
              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">Parent Category</label>
                <select
                  value={formData.parentId}
                  onChange={(e) => setFormData({ ...formData, parentId: e.target.value })}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                >
                  <option value="">None (Root Level)</option>
                  {categories.filter(c => !editingCategory || c.id !== editingCategory.id).map((c) => (
                    <option key={c.id} value={c.id}>{c.name}</option>
                  ))}
                </select>
              </div>
              <div className="flex justify-end gap-3 pt-3 border-t border-border">
                <button type="button" onClick={() => setIsModalOpen(false)} className="px-4 py-2 text-xs font-semibold border border-border rounded-xl">Cancel</button>
                <button type="submit" className="px-4 py-2 text-xs font-semibold bg-accent-brand text-white rounded-xl">Save</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
