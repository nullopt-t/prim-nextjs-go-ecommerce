import { useState } from "react";
import { Plus, SlidersHorizontal, Trash2, X } from "lucide-react";
import { INITIAL_ATTRIBUTES, getLocalAdminData, setLocalAdminData } from "../data/mockAdminData";
import { toast } from "sonner";

export function AdminAttributes() {
  const [attributes, setAttributes] = useState(() => getLocalAdminData("attributes", INITIAL_ATTRIBUTES));
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [newAttrName, setNewAttrName] = useState("");
  const [newAttrValue, setNewAttrValue] = useState("");
  const [activeAttrId, setActiveAttrId] = useState(null);

  const saveAttributes = (updated) => {
    setAttributes(updated);
    setLocalAdminData("attributes", updated);
  };

  const handleCreateAttribute = (e) => {
    e.preventDefault();
    if (!newAttrName) return;
    const newAttr = {
      id: `attr-${Date.now()}`,
      name: newAttrName,
      code: newAttrName.toLowerCase().replace(/\s+/g, "_"),
      values: []
    };
    saveAttributes([...attributes, newAttr]);
    setNewAttrName("");
    setIsModalOpen(false);
    toast.success("Attribute created!");
  };

  const handleAddValue = (attrId) => {
    if (!newAttrValue) return;
    const updated = attributes.map((a) =>
      a.id === attrId && !a.values.includes(newAttrValue)
        ? { ...a, values: [...a.values, newAttrValue] }
        : a
    );
    saveAttributes(updated);
    setNewAttrValue("");
    setActiveAttrId(null);
    toast.success("Option value added!");
  };

  const handleRemoveValue = (attrId, valToRemove) => {
    const updated = attributes.map((a) =>
      a.id === attrId
        ? { ...a, values: a.values.filter((v) => v !== valToRemove) }
        : a
    );
    saveAttributes(updated);
  };

  const handleDeleteAttribute = (id) => {
    if (confirm("Delete attribute?")) {
      saveAttributes(attributes.filter((a) => a.id !== id));
      toast.success("Attribute deleted");
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Product Attributes</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Define variant attributes like Color, Size, and Technical Specs for your products.
          </p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="flex items-center gap-2 px-4 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0"
        >
          <Plus className="size-4" />
          <span>New Attribute</span>
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {attributes.map((attr) => (
          <div key={attr.id} className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2.5">
                <SlidersHorizontal className="size-5 text-accent-brand" />
                <div>
                  <h3 className="font-bold text-base text-foreground">{attr.name}</h3>
                  <span className="text-[10px] text-muted-foreground font-mono">Code: {attr.code}</span>
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
          <div className="bg-background border border-border rounded-2xl max-w-sm w-full p-6 shadow-xl space-y-4">
            <h3 className="font-bold text-base text-foreground">Create Attribute</h3>
            <form onSubmit={handleCreateAttribute} className="space-y-4">
              <input
                type="text"
                required
                placeholder="Attribute Name (e.g. Material)"
                value={newAttrName}
                onChange={(e) => setNewAttrName(e.target.value)}
                className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
              <div className="flex justify-end gap-3">
                <button type="button" onClick={() => setIsModalOpen(false)} className="px-4 py-2 text-xs border border-border rounded-xl">Cancel</button>
                <button type="submit" className="px-4 py-2 text-xs bg-accent-brand text-white rounded-xl">Create</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
