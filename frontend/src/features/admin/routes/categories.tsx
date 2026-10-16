import { useState, useEffect } from "react";
import { Plus, Search, Edit2, Trash2, FolderTree, X, SlidersHorizontal, Tag } from "lucide-react";
import { catalogService } from "@/services/catalog";
import { adminService } from "@/services/admin";
import { toast } from "sonner";

export function AdminCategories() {
  const [categories, setCategories] = useState<any[]>([]);
  const [allAttributes, setAllAttributes] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingCategory, setEditingCategory] = useState<any>(null);
  const [viewingCategoryAttrs, setViewingCategoryAttrs] = useState<any>(null);
  const [newLinkedAttrName, setNewLinkedAttrName] = useState("");
  const [creatingAttr, setCreatingAttr] = useState(false);

  const loadCategories = async () => {
    try {
      setLoading(true);
      const [res, attrRes]: [any, any] = await Promise.all([
        catalogService.getCategories(100),
        adminService.getAttributes()
      ]);
      const list = Array.isArray(res) ? res : (Array.isArray(res?.data) ? res.data : []);
      const attrs = Array.isArray(attrRes?.data) ? attrRes.data : (Array.isArray(attrRes) ? attrRes : []);
      setAllAttributes(attrs);
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

  const handleDelete = async (id: string) => {
    if (confirm("Categories with attached products cannot be deleted directly.")) {
      try {
        await catalogService.deleteCategory(id);
        toast.success("Category deleted");
        loadCategories();
      } catch (err: any) {
        toast.error(err?.response?.data?.message || "Failed to delete category");
      }
    }
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
                  <div className="flex items-center justify-end gap-1.5">
                    <button
                      onClick={() => setViewingCategoryAttrs(cat)}
                      className="px-2.5 py-1 text-xs font-semibold rounded-lg bg-secondary text-secondary-foreground hover:bg-secondary/80 border border-border flex items-center gap-1.5 transition-colors"
                      title="Manage Category Attributes"
                    >
                      <SlidersHorizontal className="size-3.5 text-accent-brand" />
                      <span>Attributes</span>
                    </button>
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

      {/* Linked Category Attributes Modal */}
      {viewingCategoryAttrs && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-lg w-full p-6 shadow-xl space-y-4 max-h-[85vh] flex flex-col">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2.5">
                <div className="p-2 bg-accent-brand/10 text-accent-brand rounded-xl">
                  <SlidersHorizontal className="size-5" />
                </div>
                <div>
                  <h3 className="font-bold text-base text-foreground">Linked Attributes</h3>
                  <p className="text-xs text-muted-foreground">{viewingCategoryAttrs.name} category specifications</p>
                </div>
              </div>
              <button
                onClick={() => setViewingCategoryAttrs(null)}
                className="text-muted-foreground hover:text-foreground text-sm font-semibold p-1"
              >
                ✕
              </button>
            </div>

            <div className="space-y-2">
              <label className="text-xs font-semibold text-foreground">Link New Attribute to this Category</label>
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  placeholder="Attribute name (e.g. Screen Size, Material)"
                  value={newLinkedAttrName}
                  onChange={(e) => setNewLinkedAttrName(e.target.value)}
                  className="flex-1 px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden"
                />
                <button
                  disabled={creatingAttr || !newLinkedAttrName.trim()}
                  onClick={async () => {
                    if (!newLinkedAttrName.trim()) return;
                    setCreatingAttr(true);
                    try {
                      await adminService.createAttribute({
                        categoryId: viewingCategoryAttrs.id,
                        key: newLinkedAttrName.toLowerCase().replace(/\s+/g, "_"),
                        label: newLinkedAttrName.trim(),
                        valueType: "text",
                        allowedValues: [],
                        isFilterable: true,
                        sortOrder: 0
                      });
                      toast.success(`Attribute "${newLinkedAttrName}" linked to ${viewingCategoryAttrs.name}!`);
                      setNewLinkedAttrName("");
                      loadCategories();
                    } catch {
                      toast.error("Failed to link attribute to category");
                    } finally {
                      setCreatingAttr(false);
                    }
                  }}
                  className="px-4 py-2 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 disabled:opacity-50"
                >
                  {creatingAttr ? "Linking..." : "Link"}
                </button>
              </div>
            </div>

            <div className="flex-1 overflow-y-auto space-y-2 pt-2 min-h-[160px]">
              <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                Currently Linked Attributes
              </span>
              {allAttributes.filter(a => a.categoryId === viewingCategoryAttrs.id).length === 0 ? (
                <div className="py-8 text-center text-xs text-muted-foreground">
                  No attributes linked to this category yet. Type an attribute name above to create and link one.
                </div>
              ) : (
                allAttributes
                  .filter(a => a.categoryId === viewingCategoryAttrs.id)
                  .map((attr) => (
                    <div key={attr.id} className="p-3 border border-border rounded-xl bg-secondary/15 flex items-center justify-between text-xs">
                      <div>
                        <div className="font-semibold text-foreground">{attr.label || attr.key}</div>
                        <div className="text-[10px] text-muted-foreground font-mono">key: {attr.key} &bull; type: {attr.valueType}</div>
                      </div>
                      <button
                        onClick={async () => {
                          if (confirm(`Unlink and delete attribute "${attr.label || attr.key}"?`)) {
                            try {
                              await adminService.deleteAttribute(attr.id);
                              toast.success("Attribute unlinked");
                              loadCategories();
                            } catch {
                              toast.error("Failed to unlink attribute");
                            }
                          }
                        }}
                        className="p-1.5 text-muted-foreground hover:text-destructive rounded-lg transition-colors"
                        title="Unlink Attribute"
                      >
                        <Trash2 className="size-4" />
                      </button>
                    </div>
                  ))
              )}
            </div>

            <div className="pt-3 border-t border-border flex justify-end">
              <button
                onClick={() => setViewingCategoryAttrs(null)}
                className="px-4 py-2 text-xs font-semibold rounded-xl border border-border hover:bg-secondary text-foreground"
              >
                Done
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
