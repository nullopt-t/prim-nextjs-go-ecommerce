import { useState, useEffect, useCallback } from "react";
import { UploadCloud, Search, Trash2, Copy, Check, Eye, X } from "lucide-react";
import { adminService } from "@/services/admin";
import { useLocale, useTranslations } from "next-intl";
import { useAuthContext } from "@/context/AuthContext";
import { notify, apiErrorMessage } from "@/lib/notify";
import { copyTextToClipboard } from "@/lib/clipboard";
import { formatDateTime } from "@/lib/datetime";
import { ErrorState } from "@/components/ui/errorState";
import { CursorPager, useCursorPagination } from "@/components/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/lib/pagination";

export function AdminMedia() {
  const t = useTranslations("admin");
  const tCommon = useTranslations("common");
  const locale = useLocale();
  const { user } = useAuthContext();
  const isSuperadmin = user?.role === "superadmin";
  const [mediaList, setMediaList] = useState<any[]>([]);
  const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [previewItem, setPreviewItem] = useState<any>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [loadError, setLoadError] = useState(false);
  const { cursor, hasPrev, next, prev, page } = useCursorPagination();

  const loadMedia = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);
      const res: any = await adminService.getMedia({ cursor: cursor || undefined, pageSize: 50 });
      const list = Array.isArray(res?.data) ? res.data : (Array.isArray(res) ? res : []);
      setMediaList(list.map((m: any) => ({
        id: m.id,
        name: m.objectKey || `media-${m.id.slice(0, 8)}`,
        size: `${((m.fileSize || 500000) / (1024 * 1024)).toFixed(1)} MB`,
        type: (m.contentType || "").includes("image") ? "Image" : "Media",
        dimensions: "1200x1200",
        url: m.url || `http://localhost:9000/${m.bucket || "product-media"}/${m.objectKey}`,
        uploadedAt: formatDateTime(m.createdAt, locale)
      })));
      if (res?.meta) setMeta(res.meta);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [cursor, locale]);

  useEffect(() => {
    loadMedia();
  }, [loadMedia]);

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    try {
      setIsUploading(true);
      await adminService.uploadMedia(file);
      notify.success(t("media.toast.uploaded"));
      await loadMedia();
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("media.toast.uploadFailed")));
    } finally {
      setIsUploading(false);
      // Reset input value so same file can be uploaded again if needed
      e.target.value = "";
    }
  };

  const handleCopyUrl = async (url: string, id: string) => {
    const copied = await copyTextToClipboard(url);
    if (!copied) {
      notify.error(tCommon("errors.copyFailed"));
      return;
    }
    setCopiedId(id);
    notify.success(t("media.toast.copied"));
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleDelete = async (id: string) => {
    if (confirm(t("media.confirmDelete"))) {
      try {
        await adminService.deleteMedia(id);
        setMediaList(mediaList.filter((m) => m.id !== id));
        notify.success(t("media.toast.deleted"));
      } catch (err: any) {
        notify.error(apiErrorMessage(err, t("media.toast.deleteFailed")));
      }
    }
  };

  const filtered = mediaList.filter((m) => m.name.toLowerCase().includes(search.toLowerCase()));

  if (loadError) {
    return <ErrorState onRetry={loadMedia} />;
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("media.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t("media.subtitle")}</p>
        </div>
        <label className="flex items-center gap-2 px-4 py-2.5 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 cursor-pointer transition-all shrink-0">
          <UploadCloud className="size-4" />
          <span>{isUploading ? t("media.uploading") : t("media.upload")}</span>
          <input type="file" onChange={handleFileUpload} disabled={isUploading} className="hidden" accept="image/*" />
        </label>
      </div>

      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs">
        <div className="relative max-w-sm">
          <Search className="absolute start-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder={t("media.searchPlaceholder")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full ps-10 pe-4 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
          />
        </div>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
        {filtered.map((item) => (
          <div key={item.id} className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs group relative flex flex-col justify-between">
            <div className="aspect-square relative overflow-hidden bg-secondary/20">
              <img src={item.url} alt={item.name} className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
              <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                <button onClick={() => setPreviewItem(item)} className="p-2 bg-white/90 text-black rounded-lg hover:bg-white" title={t("media.preview")}>
                  <Eye className="size-4" />
                </button>
                <button onClick={() => handleCopyUrl(item.url, item.id)} className="p-2 bg-white/90 text-black rounded-lg hover:bg-white" title={t("media.copyLink")}>
                  {copiedId === item.id ? <Check className="size-4 text-emerald-600" /> : <Copy className="size-4" />}
                </button>
                {isSuperadmin && (
                  <button onClick={() => handleDelete(item.id)} className="p-2 bg-white/90 text-red-600 rounded-lg hover:bg-white" title={t("actions.delete")}>
                    <Trash2 className="size-4" />
                  </button>
                )}
              </div>
            </div>
            <div className="p-3 border-t border-border">
              <div className="font-semibold text-xs text-foreground truncate">{item.name}</div>
              <div className="text-[10px] text-muted-foreground mt-0.5 flex justify-between">
                <span>{item.size}</span>
                <span>{item.dimensions}</span>
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
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

      {previewItem && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-lg w-full max-h-[85vh] overflow-y-auto shadow-2xl">
            <div className="flex items-start justify-between gap-4 px-5 py-4 border-b border-border">
              <div className="min-w-0">
                <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
                  {t("media.preview")}
                </div>
                <h3 className="font-bold text-base text-foreground truncate">{previewItem.name}</h3>
              </div>
              <button
                onClick={() => setPreviewItem(null)}
                className="-me-1 shrink-0 rounded-lg p-1.5 text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors"
                aria-label={t("actions.close")}
              >
                <X className="size-5" />
              </button>
            </div>
            <div className="p-5 space-y-4">
              <div className="aspect-square rounded-xl overflow-hidden border border-border">
                <img src={previewItem.url} alt={previewItem.name} className="w-full h-full object-contain bg-black/5" />
              </div>
              <dl className="grid grid-cols-1 sm:grid-cols-3 gap-x-6 gap-y-4 text-xs">
                <div className="space-y-0.5">
                  <dt className="font-medium text-muted-foreground">{t("media.size")}</dt>
                  <dd className="text-foreground font-mono" dir="ltr">
                    {previewItem.size}
                  </dd>
                </div>
                <div className="space-y-0.5">
                  <dt className="font-medium text-muted-foreground">{t("media.dimensions")}</dt>
                  <dd className="text-foreground font-mono" dir="ltr">
                    {previewItem.dimensions}
                  </dd>
                </div>
                <div className="space-y-0.5">
                  <dt className="font-medium text-muted-foreground">{t("media.uploaded")}</dt>
                  <dd className="text-foreground">{previewItem.uploadedAt}</dd>
                </div>
              </dl>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
