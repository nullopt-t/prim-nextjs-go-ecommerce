import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback } from "react";
// Zero-icon standard: no lucide icons
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import CategoryTreeMap from "@/modules/admin/components/CategoryTreeMap";
import CategoryProductsPanel from "@/modules/admin/components/CategoryProductsPanel";
import { catalogService } from "@/modules/catalog/api/catalogService";
import { adminService } from "@/modules/admin/api/adminService";
import type { AdminCategory } from "@/modules/catalog/types";
import { useTranslations } from "next-intl";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";
import { revalidateCatalog } from "@/app/actions/revalidateCatalog";
import { CustomSelect } from "@/shared/ui/customSelect";

interface CategoryItem {
  id: string;
  name: string;
  slug: string;
  parent: string;
  parentId: string | undefined;
  productsCount: number;
  status: string;
  icon: string;
}

interface CategoryAttribute {
  id: string;
  name: string;
  categoryId: string;
  key: string;
  label?: string;
  valueType?: string;
  allowedValues?: string[];
}

type AdminCategoryRow = AdminCategory & { slug?: string; deletedAt?: string };

function mapCategory(c: AdminCategoryRow): CategoryItem {
  return {
    id: c.id,
    name: c.name,
    slug: c.slug || c.name.toLowerCase().replace(/\s+/g, "-"),
    parent: c.parentName || "None",
    parentId: c.parentId,
    productsCount: c.productCount || 0,
    status: "Active",
    icon: "FolderTree",
  };
}

export function AdminCategories() {
  const t = useTranslations("admin");
  const [categories, setCategories] = useState<CategoryItem[]>([]);
  // Full (unpaginated) category list backing the parent selector in the modal.
  const [allCategories, setAllCategories] = useState<CategoryItem[]>([]);
  const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [allAttributes, setAllAttributes] = useState<CategoryAttribute[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingCategory, setEditingCategory] = useState<CategoryItem | null>(null);
  const [viewingCategoryAttrs, setViewingCategoryAttrs] = useState<CategoryItem | null>(null);
  const [newLinkedAttrName, setNewLinkedAttrName] = useState("");
  const [newLinkedAttrType, setNewLinkedAttrType] = useState("text");
  const [newLinkedAttrValues, setNewLinkedAttrValues] = useState("");
  const [creatingAttr, setCreatingAttr] = useState(false);
  const [valueDrafts, setValueDrafts] = useState<Record<string, string>>({});
  const [busyValueId, setBusyValueId] = useState<string | null>(null);
  const [loadError, setLoadError] = useState(false);
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [viewMode, setViewMode] = useState<"table" | "map">("table");
  // Category highlighted in the map; its products are listed in the side panel.
  const [selectedCategory, setSelectedCategory] = useState<CategoryItem | null>(null);
  const { cursor, hasPrev, next, prev, reset, page } = useCursorPagination();

  const loadCategories = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);
      const [res, attrRes] = await Promise.all([
        catalogService.getAdminCategories({ cursor: cursor || undefined, pageSize: 20, keyset: true, search: debouncedSearch || undefined }),
        adminService.getAttributes()
      ]);
      const list: AdminCategoryRow[] = res.data ?? [];
      const attrs = (attrRes.data ?? []) as unknown as CategoryAttribute[];
      setAllAttributes(attrs);
      // The admin endpoint includes soft-deleted rows; this table only renders
      // active taxonomy nodes (status is not modelled separately in the UI).
      setCategories(list.filter((c: AdminCategoryRow) => !c.deletedAt).map(mapCategory));
      if ("meta" in res && res.meta) setMeta(res.meta as CursorMeta);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [cursor, debouncedSearch]);

  const loadAllCategories = useCallback(async () => {
    try {
      // Admin listing so empty top-level parents are selectable as parents.
      const res = await catalogService.getAdminCategories(100);
      const list: AdminCategoryRow[] = res.data ?? [];
      const mapped = list.filter((c: AdminCategoryRow) => !c.deletedAt).map(mapCategory);
      setAllCategories(mapped);
      // Keep the map selection pointing at the refreshed row, or clear it if the
      // category was deleted.
      setSelectedCategory((current) =>
        current ? mapped.find((c: CategoryItem) => c.id === current.id) ?? null : null
      );
    } catch {
      // The parent selector simply stays empty if this auxiliary fetch fails.
    }
  }, []);

  useEffect(() => {
    loadAllCategories();
  }, [loadAllCategories]);

  useEffect(() => {
    loadCategories();
  }, [loadCategories]);

  // Debounce the search box and restart pagination whenever the query changes.
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search.trim());
      reset();
    }, 300);
    return () => clearTimeout(timer);
  }, [search, reset]);

  const [formData, setFormData] = useState({
    name: "",
    slug: "",
    parentId: "",
    status: "Active"
  });

  const handleOpenModal = (cat: CategoryItem | null = null) => {
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
        notify.success(t("categories.toast.updated"));
      } else {
        await catalogService.createCategory({
          name: formData.name,
          parentId: formData.parentId ? formData.parentId : undefined,
        });
        notify.success(t("categories.toast.added"));
      }
      setIsModalOpen(false);
      void revalidateCatalog().catch(() => false);
      loadCategories();
      loadAllCategories();
    } catch (err) {
      notify.error(apiErrorMessage(err, t("categories.toast.saveFailed")));
    }
  };

  const handleDelete = async (id: string) => {
    if (confirm(t("categories.confirmDelete"))) {
      try {
        await catalogService.deleteCategory(id);
        notify.success(t("categories.toast.deleted"));
        void revalidateCatalog().catch(() => false);
        loadCategories();
        loadAllCategories();
      } catch (err) {
        notify.error(apiErrorMessage(err, t("categories.toast.deleteFailed")));
      }
    }
  };

  if (loadError) {
    return <ErrorState onRetry={loadCategories} />;
  }

  const filtered = categories;

  const categoryAttributes = (categoryId: string) =>
    allAttributes.filter((a) => a.categoryId === categoryId);

  const parseAttributeValues = (raw: string) =>
    raw.split(",").map((v) => v.trim()).filter(Boolean);

  const resetAttributeForm = () => {
    setNewLinkedAttrName("");
    setNewLinkedAttrType("text");
    setNewLinkedAttrValues("");
  };

  const handleCreateLinkedAttribute = async () => {
    const label = newLinkedAttrName.trim();
    if (!label || !viewingCategoryAttrs) return;
    setCreatingAttr(true);
    try {
      await adminService.createAttribute({
        categoryId: viewingCategoryAttrs.id,
        key: label.toLowerCase().replace(/\s+/g, "_"),
        label,
        valueType: newLinkedAttrType,
        allowedValues: parseAttributeValues(newLinkedAttrValues),
        isFilterable: true,
        sortOrder: 0
      });
      notify.success(t("categories.toast.attributeLinked", { name: label, category: viewingCategoryAttrs.name }));
      resetAttributeForm();
      loadCategories();
    } catch {
      notify.error(t("categories.toast.linkFailed"));
    } finally {
      setCreatingAttr(false);
    }
  };

  const handleAddAttributeValue = async (attr: CategoryAttribute) => {
    const value = (valueDrafts[attr.id] || "").trim();
    if (!value) return;
    const existing: string[] = Array.isArray(attr.allowedValues) ? attr.allowedValues : [];
    if (existing.includes(value)) {
      setValueDrafts((d) => ({ ...d, [attr.id]: "" }));
      return;
    }
    setBusyValueId(attr.id);
    try {
      await adminService.updateAttribute(attr.id, { allowedValues: [...existing, value] });
      notify.success(t("categories.toast.valueAdded"));
      setValueDrafts((d) => ({ ...d, [attr.id]: "" }));
      loadCategories();
    } catch {
      notify.error(t("categories.toast.valueUpdateFailed"));
    } finally {
      setBusyValueId(null);
    }
  };

  const handleRemoveAttributeValue = async (attr: CategoryAttribute, value: string) => {
    const existing: string[] = Array.isArray(attr.allowedValues) ? attr.allowedValues : [];
    setBusyValueId(attr.id);
    try {
      await adminService.updateAttribute(attr.id, { allowedValues: existing.filter((v) => v !== value) });
      notify.success(t("categories.toast.valueRemoved"));
      loadCategories();
    } catch {
      notify.error(t("categories.toast.valueUpdateFailed"));
    } finally {
      setBusyValueId(null);
    }
  };

  const handleDeleteAttribute = async (attr: CategoryAttribute) => {
    if (!confirm(t("categories.attrs.confirmUnlink", { name: attr.label || attr.key }))) return;
    try {
      await adminService.deleteAttribute(attr.id);
      notify.success(t("categories.toast.unlinked"));
      loadCategories();
    } catch {
      notify.error(t("categories.toast.unlinkFailed"));
    }
  };

  return (
    <div className="space-y-5">
      <PageHeader
        title={t("categories.title")}
        actions={
          <>
            <div className="inline-flex items-center gap-1 rounded-md border border-border p-1">
              <button
                type="button"
                onClick={() => setViewMode("table")}
                aria-pressed={viewMode === "table"}
                className={`rounded px-2 py-1 text-xs font-medium transition-colors ${
                  viewMode === "table" ? "bg-secondary text-foreground" : "text-muted-foreground hover:text-foreground"
                }`}
              >
                {t("categories.viewTable")}
              </button>
              <button
                type="button"
                onClick={() => setViewMode("map")}
                aria-pressed={viewMode === "map"}
                className={`rounded px-2 py-1 text-xs font-medium transition-colors ${
                  viewMode === "map" ? "bg-secondary text-foreground" : "text-muted-foreground hover:text-foreground"
                }`}
              >
                {t("categories.viewMap")}
              </button>
            </div>
            <button
              onClick={() => handleOpenModal()}
              className="rounded-xl border border-border/70 bg-card px-3 py-2 text-sm font-medium text-foreground hover:bg-muted/30 active:scale-[0.99] select-none transition-all"
            >
              {t("categories.add")}
            </button>
          </>
        }
      />

      {viewMode === "table" ? (
        <>
      <input
        type="text"
        placeholder={t("categories.searchPlaceholder")}
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            setDebouncedSearch(search.trim());
            reset();
          }
        }}
        className="w-full sm:max-w-xs rounded-xl border border-border/70 bg-card px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground hover:bg-muted/30 active:scale-[0.99] select-none transition-all focus:outline-none"
      />

      <div className="overflow-hidden rounded-xl border border-border/70 bg-card">
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
            <thead className="text-xs text-muted-foreground">
              <tr className="border-b border-border">
                <th className="py-2.5 ps-4 text-start font-medium">{t("categories.colName")}</th>
                <th className="py-2.5 text-start font-medium">{t("categories.colSlug")}</th>
                <th className="py-2.5 text-start font-medium">{t("categories.colParent")}</th>
                <th className="py-2.5 text-start font-medium">{t("categories.colProducts")}</th>
                <th className="py-2.5 text-start font-medium">{t("categories.colStatus")}</th>
                <th className="py-2.5 pe-4 text-end font-medium">{t("categories.colActions")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {filtered.map((cat) => (
                <tr key={cat.id} className="transition-colors hover:bg-secondary/30">
                  <td className="py-3 ps-4 font-medium text-foreground">{cat.name}</td>
                  <td className="py-3 font-mono text-xs text-muted-foreground">{cat.slug}</td>
                  <td className="py-3 text-xs text-muted-foreground">{cat.parent}</td>
                  <td className="py-3 font-medium text-foreground tabular-nums">{t("categories.items", { count: cat.productsCount })}</td>
                  <td className="py-3">
                    <StatusPill tone={statusTone(cat.status)}>{cat.status}</StatusPill>
                  </td>
                  <td className="py-3 pe-4 text-end">
                    <div className="inline-flex items-center gap-1">
                      <button
                        onClick={() => setViewingCategoryAttrs(cat)}
                        className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                      >
                        {t("categories.attributes")} ({categoryAttributes(cat.id).length})
                      </button>
                      <button
                        onClick={() => handleOpenModal(cat)}
                        className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                      >
                        {t("actions.edit")}
                      </button>
                      <button
                        onClick={() => handleDelete(cat.id)}
                        className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                      >
                        {t("actions.delete")}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
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
          shown={filtered.length}
        />
      </div>
        </>
      ) : (
        <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_360px]">
          <CategoryTreeMap
            categories={allCategories}
            selectedId={selectedCategory?.id}
            onSelect={(category) => setSelectedCategory(category as CategoryItem)}
            emptyLabel={t("categories.mapEmpty")}
            emptyNodeLabel={t("categories.mapEmptyNode")}
            itemsLabel={(count) => t("categories.items", { count })}
            hint={t("categories.mapHint")}
            collapseLabel={t("categories.collapseNode")}
            expandLabel={t("categories.expandNode")}
            collapseAllLabel={t("categories.collapseAll")}
            expandAllLabel={t("categories.expandAll")}
          />
          <CategoryProductsPanel
            category={selectedCategory}
            categories={allCategories}
            onEdit={() => selectedCategory && handleOpenModal(selectedCategory)}
            onClear={() => setSelectedCategory(null)}
          />
        </div>
      )}

      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-md space-y-4 rounded-xl border border-border/70 bg-card p-5 shadow-xl">
            <div className="flex items-center justify-between border-b border-border/70 pb-3">
              <h2 className="text-sm font-semibold text-foreground">
                {editingCategory ? t("categories.editTitle") : t("categories.addTitle")}
              </h2>
              <button
                onClick={() => setIsModalOpen(false)}
                className="rounded-lg px-2 py-1 text-xs font-medium text-muted-foreground hover:bg-muted/30 hover:text-foreground active:scale-[0.99] select-none transition-all"
                aria-label={t("actions.cancel")}
              >
                {t("actions.close")}
              </button>
            </div>
            <form onSubmit={handleSave} className="space-y-4">
              <div className="space-y-1">
                <label className="text-sm font-medium text-foreground">{t("categories.form.name")}</label>
                <input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                />
              </div>
              <div className="space-y-1">
                <CustomSelect
                	value={formData.parentId}
                	onChange={(value) => setFormData({ ...formData, parentId: value })}
                	options={[
                		{ value: "", label: t("categories.form.root") },
                		...allCategories
                		.filter((c) => !editingCategory || c.id !== editingCategory.id)
                		.map((c) => ({ value: c.id, label: c.name })),
                	]}
/>
              </div>
              <div className="flex justify-end gap-2 border-t border-border/70 pt-3">
                <button type="button" onClick={() => setIsModalOpen(false)} className="rounded-xl border border-border/70 bg-card px-3 py-2 text-sm font-medium text-foreground hover:bg-muted/30 active:scale-[0.99] select-none transition-all">{t("actions.cancel")}</button>
                <button type="submit" className="rounded-xl bg-foreground px-3 py-2 text-sm font-medium text-background active:scale-[0.99] select-none transition-all">{t("actions.save")}</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Linked Category Attributes Modal */}
      {viewingCategoryAttrs && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="flex max-h-[85vh] w-full max-w-lg flex-col space-y-4 rounded-xl border border-border/70 bg-card p-5 shadow-xl">
            <div className="flex items-center justify-between border-b border-border/70 pb-3">
              <div>
                <h2 className="text-sm font-semibold text-foreground">{t("categories.attrs.title")}</h2>
                <p className="text-xs text-muted-foreground">{t("categories.attrs.subtitle", { name: viewingCategoryAttrs.name })}</p>
              </div>
              <button
                onClick={() => setViewingCategoryAttrs(null)}
                className="rounded-lg px-2 py-1 text-xs font-medium text-muted-foreground hover:bg-muted/30 hover:text-foreground active:scale-[0.99] select-none transition-all"
              >
                {t("actions.close")}
              </button>
            </div>

            <div className="space-y-3">
              <label className="text-xs font-semibold text-foreground">{t("categories.attrs.linkLabel")}</label>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                <input
                  type="text"
                  placeholder={t("categories.attrs.linkPlaceholder")}
                  value={newLinkedAttrName}
                  onChange={(e) => setNewLinkedAttrName(e.target.value)}
                  className="px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden"
                />
                <CustomSelect
                	value={newLinkedAttrType}
                	onChange={(value) => setNewLinkedAttrType(value)}
                	options={[
                		{ value: "text", label: t("categories.attrs.types.text") },
                		{ value: "number", label: t("categories.attrs.types.number") },
                		{ value: "boolean", label: t("categories.attrs.types.boolean") },
                	]}
                	size="sm"
/>
              </div>
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  placeholder={t("categories.attrs.valuesPlaceholder")}
                  value={newLinkedAttrValues}
                  onChange={(e) => setNewLinkedAttrValues(e.target.value)}
                  className="flex-1 px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground focus:outline-hidden"
                />
                <button
                  disabled={creatingAttr || !newLinkedAttrName.trim()}
                  onClick={handleCreateLinkedAttribute}
                  className="px-4 py-2 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 disabled:opacity-50 shrink-0"
                >
                  {creatingAttr ? t("categories.attrs.linking") : t("actions.add")}
                </button>
              </div>
            </div>

            <div className="flex-1 overflow-y-auto space-y-2 pt-2 min-h-[160px]">
              <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                {t("categories.attrs.current")}
              </span>
              {categoryAttributes(viewingCategoryAttrs.id).length === 0 ? (
                <div className="py-8 text-center text-xs text-muted-foreground">
                  {t("categories.attrs.empty")}
                </div>
              ) : (
                categoryAttributes(viewingCategoryAttrs.id).map((attr) => {
                  const values: string[] = Array.isArray(attr.allowedValues) ? attr.allowedValues : [];
                  const isBusy = busyValueId === attr.id;
                  return (
                    <div key={attr.id} className="p-3 border border-border rounded-xl bg-secondary/15 space-y-2.5 text-xs">
                      <div className="flex items-start justify-between gap-2">
                        <div>
                          <div className="font-semibold text-foreground">{attr.label || attr.key}</div>
                          <div className="text-[10px] text-muted-foreground font-mono">key: {attr.key} &bull; type: {attr.valueType}</div>
                        </div>
                        <button
                          onClick={() => handleDeleteAttribute(attr)}
                          className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                          title={t("categories.attrs.unlink")}
                        >
                          {t("actions.delete")}
                        </button>
                      </div>

                      {values.length > 0 && (
                        <div className="flex flex-wrap gap-1.5">
                          {values.map((value) => (
                            <span key={value} className="inline-flex items-center gap-1 px-2 py-0.5 bg-secondary text-secondary-foreground rounded-lg border border-border">
                              <span>{value}</span>
                              <button
                                disabled={isBusy}
                                onClick={() => handleRemoveAttributeValue(attr, value)}
                                className="rounded-lg px-1.5 text-xs text-muted-foreground hover:bg-muted/30 hover:text-destructive active:scale-[0.99] select-none transition-all disabled:opacity-50"
                                title={t("categories.attrs.removeValue")}
                              >
                                {t("actions.close")}
                              </button>
                            </span>
                          ))}
                        </div>
                      )}

                      <div className="flex items-center gap-2">
                        <input
                          type="text"
                          placeholder={t("categories.attrs.valuePlaceholder")}
                          value={valueDrafts[attr.id] || ""}
                          onChange={(e) => setValueDrafts((d) => ({ ...d, [attr.id]: e.target.value }))}
                          onKeyDown={(e) => {
                            if (e.key === "Enter") {
                              e.preventDefault();
                              handleAddAttributeValue(attr);
                            }
                          }}
                          className="flex-1 px-2.5 py-1.5 bg-background border border-border rounded-lg text-xs text-foreground focus:outline-hidden"
                        />
                        <button
                          disabled={isBusy || !(valueDrafts[attr.id] || "").trim()}
                          onClick={() => handleAddAttributeValue(attr)}
                          className="px-3 py-1.5 bg-secondary text-secondary-foreground border border-border font-semibold rounded-lg hover:bg-secondary/80 disabled:opacity-50 shrink-0"
                        >
                          {t("categories.attrs.addValue")}
                        </button>
                      </div>
                    </div>
                  );
                })
              )}
            </div>

            <div className="pt-3 border-t border-border flex justify-end">
              <button
                onClick={() => setViewingCategoryAttrs(null)}
                className="px-4 py-2 text-xs font-semibold rounded-xl border border-border hover:bg-secondary text-foreground"
              >
                {t("actions.done")}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
