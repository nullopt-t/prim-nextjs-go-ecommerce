"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter } from "@/shared/i18n/navigation";
import { useTranslations, useLocale } from "next-intl";
import { ArrowLeft, Plus, Loader2 } from "lucide-react";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill } from "@/shared/ui/statusPill";
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
    <div className="mx-auto w-full max-w-5xl space-y-5 pb-16">
      <div>
        <button
          type="button"
          onClick={() => router.push("/admin/products")}
          className="mb-2 inline-flex items-center gap-1.5 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-3.5 rtl:rotate-180" />
          {t("products.backToProducts")}
        </button>
        <PageHeader
          title={t("products.draftsTitle")}
          description={t("products.draftsSubtitle")}
          actions={
            <>
              <span className="text-xs text-muted-foreground">{drafts.length}</span>
              <button
                onClick={() => router.push("/admin/products/new")}
                className="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
              >
                <Plus className="size-4" />
                {t("products.addNew")}
              </button>
            </>
          }
        />
      </div>

      {/* Content */}
      {loading ? (
        <div className="flex min-h-[40vh] flex-col items-center justify-center gap-3">
          <Loader2 className="size-6 animate-spin text-muted-foreground" />
          <p className="text-sm text-muted-foreground">{t("products.audit.loading")}</p>
        </div>
      ) : drafts.length === 0 ? (
        <div className="mx-auto max-w-lg space-y-3 rounded-lg border border-dashed border-border p-12 text-center">
          <div className="space-y-1">
            <h2 className="text-sm font-semibold text-foreground">{t("products.noDrafts")}</h2>
            <p className="text-sm text-muted-foreground">{t("products.wizard.draftNotice")}</p>
          </div>
          <button
            type="button"
            onClick={() => router.push("/admin/products/new")}
            className="inline-flex items-center gap-1.5 rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
          >
            <Plus className="size-4" />
            {t("products.addNew")}
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
                className="flex flex-col justify-between gap-3 rounded-lg border border-border bg-background p-4 transition-colors hover:bg-secondary/20 sm:flex-row sm:items-center"
              >
                <div className="min-w-0 space-y-1.5">
                  <div className="flex flex-wrap items-center gap-2">
                    <h2 className="max-w-md truncate text-sm font-semibold text-foreground">
                      {draftTitle || t("products.createProduct")}
                    </h2>
                    <StatusPill tone="warning">{t("status.draft")}</StatusPill>
                  </div>

                  <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
                    <span>{categoryName}</span>
                    {brandName ? <span>{t("products.brandLabel", { name: brandName })}</span> : null}
                    {updatedAtFormatted ? (
                      <span>{t("products.lastUpdated", { date: updatedAtFormatted })}</span>
                    ) : null}
                  </div>
                </div>

                <div className="flex shrink-0 items-center gap-1 self-end sm:self-center">
                  <button
                    type="button"
                    onClick={() => handleDeleteDraft(draftId)}
                    className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                  >
                    {t("products.deleteProduct")}
                  </button>
                  <button
                    type="button"
                    onClick={() => router.push(`/admin/products/${draftId}`)}
                    className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                  >
                    {t("products.resumeDraft")}
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
