"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "@/shared/i18n/navigation";
import { useTranslations, useLocale } from "next-intl";
import {
  ArrowLeft,
  Plus,
  Edit2,
  Trash2,
  FileText,
  Loader2,
  Layers,
  Calendar,
} from "lucide-react";
import { catalogService } from "@/modules/catalog/api/catalogService";
import type { AdminCategory, Brand } from "@/modules/catalog/types";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { formatDateTime } from "@/shared/lib/datetime";

interface DraftProductRaw {
  id: string;
  ID?: string;
  title?: string;
  Title?: string;
  status?: string;
  Status?: string;
  categoryId?: string;
  category?: string | { id?: string; name?: string } | null;
  CategoryID?: string;
  brandId?: string;
  brand?: string | { id?: string; name?: string } | null;
  BrandID?: string;
  updatedAt?: string;
  UpdatedAt?: string;
  createdAt?: string;
  CreatedAt?: string;
}

export function AdminProductDrafts() {
  const t = useTranslations("admin");
  const router = useRouter();
  const locale = useLocale();

  const [drafts, setDrafts] = useState<DraftProductRaw[]>([]);
  const [categories, setCategories] = useState<AdminCategory[]>([]);
  const [brands, setBrands] = useState<Brand[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);

  const loadData = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);

      const [prodsRes, catsRes, brandsRes] = await Promise.all([
        catalogService.getAdminProducts({ pageSize: 100, status: "draft" }),
        catalogService.getAdminCategories(100),
        catalogService.getBrands(100),
      ]);

      const catList = (Array.isArray(catsRes) ? catsRes : (Array.isArray(catsRes?.data) ? catsRes.data : []))
        .filter((c: AdminCategory & { deletedAt?: string }) => !c.deletedAt);
      setCategories(catList);

      const brandList = Array.isArray(brandsRes) ? brandsRes : (Array.isArray(brandsRes?.data) ? brandsRes.data : []);
      setBrands(brandList);

      const rawItems = Array.isArray(prodsRes) ? prodsRes : (Array.isArray(prodsRes?.data) ? prodsRes.data : []);
      // Filter strictly to draft items (handling both camelCase and Go PascalCase JSON)
      const draftList = rawItems.filter((p: DraftProductRaw) => {
        const s = String(p.status || p.Status || "").toLowerCase();
        return s === "draft";
      });

      setDrafts(draftList);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleDeleteDraft = async (id: string) => {
    if (!confirm(t("products.confirmDelete"))) return;
    try {
      await catalogService.deleteProduct(id);
      setDrafts((prev) => prev.filter((d) => (d.id || d.ID) !== id));
      notify.success(t("products.toast.deleted"));
    } catch (err) {
      notify.error(apiErrorMessage(err, t("products.toast.deleteFailed")));
    }
  };

  if (loadError) {
    return <ErrorState onRetry={loadData} />;
  }

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-16">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-background p-5 rounded-2xl border border-border shadow-xs">
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => router.push("/admin/products")}
            className="p-2 rounded-xl border border-border hover:bg-secondary text-muted-foreground hover:text-foreground transition-all shrink-0"
            title={t("products.backToProducts")}
          >
            <ArrowLeft className="size-4 rtl:rotate-180" />
          </button>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-foreground">{t("products.draftsTitle")}</h1>
              <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-amber-500/10 text-amber-600 border border-amber-500/30">
                {drafts.length}
              </span>
            </div>
            <p className="text-xs text-muted-foreground mt-0.5">
              {t("products.draftsSubtitle")}
            </p>
          </div>
        </div>

        <button
          onClick={() => router.push("/admin/products/new")}
          className="flex items-center gap-2 px-4 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all shrink-0"
        >
          <Plus className="size-4" />
          <span>{t("products.addNew")}</span>
        </button>
      </div>

      {/* Content */}
      {loading ? (
        <div className="flex flex-col items-center justify-center min-h-[40vh] gap-3">
          <Loader2 className="size-8 animate-spin text-accent-brand" />
          <p className="text-xs text-muted-foreground">{t("products.audit.loading")}</p>
        </div>
      ) : drafts.length === 0 ? (
        <div className="p-12 text-center border border-dashed border-border rounded-2xl bg-secondary/10 space-y-3 max-w-lg mx-auto">
          <FileText className="size-10 mx-auto text-muted-foreground opacity-40" />
          <div className="space-y-1">
            <h3 className="text-sm font-bold text-foreground">{t("products.noDrafts")}</h3>
            <p className="text-xs text-muted-foreground">
              {t("products.wizard.draftNotice")}
            </p>
          </div>
          <button
            type="button"
            onClick={() => router.push("/admin/products/new")}
            className="px-4 py-2 bg-accent-brand text-white rounded-xl text-xs font-semibold hover:bg-accent-brand/90 transition-all inline-flex items-center gap-1.5 shadow-xs"
          >
            <Plus className="size-3.5" />
            <span>{t("products.addNew")}</span>
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-3">
          {drafts.map((draft, idx) => {
            const draftId = draft.id || draft.ID || '';
            const draftTitle = draft.title || draft.Title;
            const catId = draft.categoryId || (typeof draft.category === "object" ? draft.category?.id : undefined) || draft.CategoryID;
            const brandId = draft.brandId || (typeof draft.brand === "object" ? draft.brand?.id : undefined) || draft.BrandID;
            const cat = categories.find((c) => c.id === catId);
            const categoryName =
              (typeof draft.category === "object" ? draft.category?.name : draft.category) ||
              cat?.name ||
              t("products.colCategory");
            const brand = brands.find((b) => b.id === brandId);
            const brandName =
              (typeof draft.brand === "object" ? draft.brand?.name : draft.brand) ||
              brand?.name ||
              "";
            const rawDate = draft.updatedAt || draft.UpdatedAt || draft.createdAt || draft.CreatedAt;
            const updatedAtFormatted = rawDate ? formatDateTime(rawDate, locale) : "";

            return (
              <div
                key={draftId ? `draft-${draftId}` : `draft-idx-${idx}`}
                className="bg-background border border-border rounded-2xl p-4 sm:p-5 shadow-xs flex flex-col sm:flex-row sm:items-center justify-between gap-4 hover:border-accent-brand/40 transition-all"
              >
                <div className="space-y-1.5 min-w-0">
                  <div className="flex items-center gap-2">
                    <h3 className="text-sm font-bold text-foreground truncate max-w-md">
                      {draftTitle || t("products.createProduct")}
                    </h3>
                    <span className="px-2 py-0.5 rounded-full text-[9px] font-bold uppercase tracking-wider bg-amber-500/10 text-amber-600 border border-amber-500/30 shrink-0">
                      {t("status.draft")}
                    </span>
                  </div>

                  <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
                    <span className="flex items-center gap-1">
                      <Layers className="size-3.5 text-accent-brand" />
                      <span>{categoryName}</span>
                    </span>
                    {brandName && (
                      <>
                        <span>•</span>
                        <span>{t("products.brandLabel", { name: brandName })}</span>
                      </>
                    )}
                    {updatedAtFormatted && (
                      <>
                        <span>•</span>
                        <span className="flex items-center gap-1">
                          <Calendar className="size-3" />
                          <span>{t("products.lastUpdated", { date: updatedAtFormatted })}</span>
                        </span>
                      </>
                    )}
                  </div>
                </div>

                <div className="flex items-center gap-2 shrink-0 self-end sm:self-center">
                  <button
                    type="button"
                    onClick={() => handleDeleteDraft(draftId)}
                    className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-xl transition-colors"
                    title={t("products.deleteProduct")}
                  >
                    <Trash2 className="size-4" />
                  </button>

                  <button
                    type="button"
                    onClick={() => router.push(`/admin/products/${draftId}`)}
                    className="flex items-center gap-1.5 px-4 py-2 bg-accent-brand text-white text-xs font-semibold rounded-xl hover:bg-accent-brand/90 transition-all shadow-xs"
                  >
                    <Edit2 className="size-3.5" />
                    <span>{t("products.resumeDraft")}</span>
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
