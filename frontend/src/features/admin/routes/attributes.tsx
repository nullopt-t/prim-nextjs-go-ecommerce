import { useState, useEffect } from "react";
import { Plus, SlidersHorizontal, Trash2, X, FolderTree, Tag } from "lucide-react";
import { adminService } from "@/services/admin";
import { catalogService } from "@/services/catalog";
import { toast } from "sonner";

export function AdminAttributes() {
  const [attributes, setAttributes] = useState<any[]>([]);
  const [categories, setCategories] = useState<any[]>([]);
  const [selectedCategoryFilter, setSelectedCategoryFilter] = useState("all");
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [newAttrName, setNewAttrName] = useState("");
  const [newAttrCategoryId, setNewAttrCategoryId] = useState("");
  const [newAttrValueType, setNewAttrValueType] = useState("text");
  const [newAttrValue, setNewAttrValue] = useState("");
  const [activeAttrId, setActiveAttrId] = useState<string | null>(null);

  const loadData = async () => {
    try {
      setLoading(true);
      const [attrRes, catRes]: [any, any] = await Promise.all([
        adminService.getAttributes(),
        catalogService.getCategories(100)
      ]);
      const list = Array.isArray(attrRes?.data) ? attrRes.data : (Array.isArray(attrRes) ? attrRes : []);
      setAttributes(list.map((a: any) => ({
        id: a.id,
        categoryId: a.categoryId,
        name: a.label || a.key,
        code: a.key,
        valueType: a.valueType || "text",
        categoryName: a.categoryName || "Global / Catalog",
        values: Array.isArray(a.allowedValues) ? a.allowedValues : []
      })));

      const cats = Array.isArray(catRes?.data) ? catRes.data : (Array.isArray(catRes) ? catRes : []);
      setCategories(cats);
    } catch {
      toast.error("Failed to load attributes or categories from server");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleCreateAttribute = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newAttrName) return;

    const payload = {
      categoryId: newAttrCategoryId || undefined,
      key: newAttrName.toLowerCase().replace(/\s+/g, "_"),
      label: newAttrName,
      valueType: newAttrValueType || "text",
      allowedValues: [],
      isFilterable: true,
      sortOrder: 0
    };

    try {
      await adminService.createAttribute(payload);
      toast.success("Attribute created and linked to catalog category!");
      setNewAttrName("");
      setNewAttrCategoryId("");
      setNewAttrValueType("text");
      setIsModalOpen(false);
      loadData();
    } catch (err: any) {
      toast.error(err?.response?.data?.message || "Failed to create attribute");
    }
  };

  const handleAddValue = (attrId: string) => {
    if (!newAttrValue) return;
    const updated = attributes.map((a) =>
      a.id === attrId && !a.values.includes(newAttrValue)
        ? { ...a, values: [...a.values, newAttrValue] }
        : a
    );
    setAttributes(updated);
    setNewAttrValue("");
    setActiveAttrId(null);
    toast.success("Option value added!");
  };

  const handleRemoveValue = (attrId: string, valToRemove: string) => {
    const updated = attributes.map((a) =>
      a.id === attrId
        ? { ...a, values: a.values.filter((v: string) => v !== valToRemove) }
        : a
    );
    setAttributes(updated);
  };

  const handleDeleteAttribute = async (id: string) => {
    if (confirm("Delete attribute?")) {
      try {
        await adminService.deleteAttribute(id);
        toast.success("Attribute deleted");
        loadData();
      } catch (err: any) {
        toast.error(err?.response?.data?.message || "Failed to delete attribute");
      }
    }
  };

  const filteredAttributes = selectedCategoryFilter === "all"
    ? attributes
    : attributes.filter((a) => a.categoryId === selectedCategoryFilter || (selectedCategoryFilter === "global" && !a.categoryId));

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Product Attributes</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Define variant attributes like Color, Size, Storage, and Specs linked to catalog categories.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={selectedCategoryFilter}
            onChange={(e) => setSelectedCategoryFilter(e.target.value)}
            className="px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden"
          >
            <option value="all">All Catalog Categories</option>
            <option value="global">Global / Unassigned</option>
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                Category: {c.name}
              </option>
            ))}
          </select>
          <button
            onClick={() => setIsModalOpen(true)}
            className="flex items-center gap-2 px-4 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0"
          >
            <Plus className="size-4" />
            <span>New Attribute</span>
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {filteredAttributes.map((attr) => (
          <div key={attr.id} className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2.5">
                <SlidersHorizontal className="size-5 text-accent-brand" />
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="font-bold text-base text-foreground">{attr.name}</h3>
                    <span className="px-2 py-0.5 rounded-full text-[10px] font-semibold bg-accent-brand/10 text-accent-brand border border-accent-brand/20">
                      {attr.categoryName}
                    </span>
                  </div>
                  <span className="text-[10px] text-muted-foreground font-mono">Code: {attr.code} &bull; Type: {attr.valueType}</span>
                </div>
              </div>
              <button onClick={() => handleDeleteAttribute(attr.id)} className="p-1.5 text-muted-foreground hover:text-destructive rounded-lg">
                <Trash2 className="size-4" />
              </button>
            </div>

            <div className="space-y-2">
              <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Option Values ({attr.values.length})</span>
              <div className="flex flex-wrap gap-2">
                {attr.values.map((val) => (
                  <span key={val} className="inline-flex items-center gap-1.5 px-3 py-1 bg-secondary text-secondary-foreground text-xs font-medium rounded-lg border border-border">
                    <span>{val}</span>
                    <button onClick={() => handleRemoveValue(attr.id, val)} className="text-muted-foreground hover:text-destructive">
                      <X className="size-3" />
                    </button>
                  </span>
                ))}
              </div>
            </div>

            {/* Quick Add Option Value */}
            <div className="pt-2">
              {activeAttrId === attr.id ? (
                <div className="flex items-center gap-2">
                  <input
                    type="text"
                    placeholder="e.g. XL, 256GB..."
                    value={newAttrValue}
                    onChange={(e) => setNewAttrValue(e.target.value)}
                    className="flex-1 px-3 py-1.5 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden"
                  />
                  <button onClick={() => handleAddValue(attr.id)} className="px-3 py-1.5 bg-accent-brand text-white text-xs font-semibold rounded-xl">Add</button>
                  <button onClick={() => setActiveAttrId(null)} className="px-3 py-1.5 text-xs text-muted-foreground">Cancel</button>
                </div>
              ) : (
                <button
                  onClick={() => setActiveAttrId(attr.id)}
                  className="text-xs font-semibold text-accent-brand hover:underline flex items-center gap-1"
                >
                  <Plus className="size-3.5" />
                  <span>Add Option Value</span>
                </button>
              )}
            </div>
          </div>
        ))}
      </div>

      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-md w-full p-6 shadow-xl space-y-4">
            <h3 className="font-bold text-base text-foreground">Create & Link Attribute</h3>
            <p className="text-xs text-muted-foreground">Link this attribute to a specific category in your catalog.</p>
            <form onSubmit={handleCreateAttribute} className="space-y-4">
              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">Attribute Name *</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Screen Size, Color, Storage"
                  value={newAttrName}
                  onChange={(e) => setNewAttrName(e.target.value)}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden"
                />
              </div>

              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">Catalog Category *</label>
                <select
                  required
                  value={newAttrCategoryId}
                  onChange={(e) => setNewAttrCategoryId(e.target.value)}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden"
                >
                  <option value="">-- Select Catalog Category --</option>
                  {categories.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">Value Type</label>
                <select
                  value={newAttrValueType}
                  onChange={(e) => setNewAttrValueType(e.target.value)}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden"
                >
                  <option value="text">Text (e.g. Colors, Sizes)</option>
                  <option value="number">Number (e.g. Weight, Capacity)</option>
                  <option value="boolean">Boolean (True / False)</option>
                </select>
              </div>

              <div className="flex justify-end gap-3 pt-2">
                <button type="button" onClick={() => setIsModalOpen(false)} className="px-4 py-2 text-xs border border-border rounded-xl">Cancel</button>
                <button type="submit" className="px-4 py-2 text-xs bg-accent-brand text-white font-semibold rounded-xl">Create & Link</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
