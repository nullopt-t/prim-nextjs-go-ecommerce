import { useState, useEffect } from "react";
import { Plus, Search, Edit2, Trash2, ExternalLink, X } from "lucide-react";
import { catalogService } from "@/services/catalog";
import { useTranslations } from "next-intl";
import { notify, apiErrorMessage } from "@/lib/notify";
import { ErrorState } from "@/components/ui/errorState";

export function AdminBrands() {
  const t = useTranslations("admin");
  const [brands, setBrands] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingBrand, setEditingBrand] = useState<any>(null);
  const [loadError, setLoadError] = useState(false);

  const loadBrands = async () => {
    try {
      setLoading(true);
      setLoadError(false);
      const res: any = await catalogService.getBrands(100);
      const list = Array.isArray(res) ? res : (Array.isArray(res?.data) ? res.data : []);
      setBrands(list.map((b: any) => ({
        id: b.id,
        name: b.name,
        website: b.link || b.website || "https://prim.com",
        productsCount: b.productCount || 0,
        status: "Active",
        logo: b.logoUrl || "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=100&q=80"
      })));
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadBrands();
  }, []);

  const [formData, setFormData] = useState({ name: "", website: "", logo: "" });

  const handleOpenModal = (brand: any = null) => {
    if (brand) {
      setEditingBrand(brand);
      setFormData({ name: brand.name, website: brand.website, logo: brand.logo });
    } else {
      setEditingBrand(null);
      setFormData({ name: "", website: "https://", logo: "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=100&q=80" });
    }
    setIsModalOpen(true);
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name) return;

    try {
      if (editingBrand) {
        await catalogService.updateBrand(editingBrand.id, {
          name: formData.name,
          link: formData.website || undefined
        });
        notify.success(t("brands.toast.updated"));
      } else {
        await catalogService.createBrand({
          name: formData.name,
          link: formData.website || undefined
        });
        notify.success(t("brands.toast.created"));
      }
      setIsModalOpen(false);
      loadBrands();
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("brands.toast.saveFailed")));
    }
  };

  const handleDelete = async (id: string) => {
    if (confirm(t("brands.confirmDelete"))) {
      try {
        await catalogService.deleteBrand(id);
        notify.success(t("brands.toast.deleted"));
        loadBrands();
      } catch (err: any) {
        notify.error(apiErrorMessage(err, t("brands.toast.deleteFailed")));
      }
    }
  };

  const filtered = brands.filter((b) => b.name.toLowerCase().includes(search.toLowerCase()));

  if (loadError) {
    return <ErrorState onRetry={loadBrands} />;
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("brands.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t("brands.subtitle")}</p>
        </div>
        <button
          onClick={() => handleOpenModal()}
          className="flex items-center gap-2 px-4 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0"
        >
          <Plus className="size-4" />
          <span>{t("brands.add")}</span>
        </button>
      </div>

      <div className="relative max-w-md">
        <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
        <input
          type="text"
          placeholder={t("brands.searchPlaceholder")}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full pl-10 pr-4 py-2.5 bg-background border border-border rounded-xl text-xs text-foreground focus:outline-none focus:ring-1 focus:ring-accent-brand"
        />
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {filtered.map((b) => (
          <div key={b.id} className="bg-background border border-border rounded-2xl p-5 shadow-xs flex flex-col justify-between space-y-4">
            <div className="flex items-start justify-between">
              <div className="flex items-center gap-3">
                <img src={b.logo} alt={b.name} className="size-12 rounded-xl object-cover border border-border" />
                <div>
                  <h3 className="font-bold text-sm text-foreground">{b.name}</h3>
                  <a
                    href={b.website}
                    target="_blank"
                    rel="noreferrer"
                    className="text-[10px] text-accent-brand hover:underline flex items-center gap-1 mt-0.5"
                  >
                    <span>{t("brands.visitWebsite")}</span>
                    <ExternalLink className="size-3" />
                  </a>
                </div>
              </div>
              <span className="px-2 py-0.5 bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400 text-[10px] font-bold rounded-full">
                {t(`status.${String(b.status).toLowerCase()}`) || b.status}
              </span>
            </div>

            <div className="flex items-center justify-between pt-3 border-t border-border text-xs">
              <span className="text-muted-foreground font-medium">{t("brands.associatedProducts", { count: b.productsCount })}</span>
              <div className="flex items-center gap-2">
                <button onClick={() => handleOpenModal(b)} className="p-1 text-muted-foreground hover:text-accent-brand">
                  <Edit2 className="size-4" />
                </button>
                <button onClick={() => handleDelete(b.id)} className="p-1 text-muted-foreground hover:text-destructive">
                  <Trash2 className="size-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-md w-full p-6 shadow-xl space-y-4">
            <div className="flex items-center justify-between border-b border-border pb-3">
              <h3 className="font-bold text-base text-foreground">{editingBrand ? t("brands.editTitle") : t("brands.createTitle")}</h3>
              <button onClick={() => setIsModalOpen(false)} className="text-muted-foreground hover:text-foreground">
                <X className="size-5" />
              </button>
            </div>
            <form onSubmit={handleSave} className="space-y-4">
              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">{t("brands.form.name")}</label>
                <input
                  type="text"
                  required
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                />
              </div>
              <div className="space-y-1">
                <label className="text-xs font-semibold text-foreground">{t("brands.form.website")}</label>
                <input
                  type="text"
                  value={formData.website}
                  onChange={(e) => setFormData({ ...formData, website: e.target.value })}
                  className="w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                />
              </div>
              <div className="flex justify-end gap-3 pt-3 border-t border-border">
                <button type="button" onClick={() => setIsModalOpen(false)} className="px-4 py-2 text-xs font-semibold border border-border rounded-xl">{t("actions.cancel")}</button>
                <button type="submit" className="px-4 py-2 text-xs font-semibold bg-accent-brand text-white rounded-xl">{t("actions.save")}</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
