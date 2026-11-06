import { useState, useEffect, useCallback, useRef } from "react";
import { Search, X, Loader2 } from "lucide-react";
import { adminService } from "@/modules/admin/api/adminService";
import type { AdminMedia } from "@/modules/admin/types";
import { useLocale, useTranslations } from "next-intl";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { copyTextToClipboard } from "@/shared/lib/clipboard";
import { formatDateTime } from "@/shared/lib/datetime";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";

const MEDIA_PAGE_SIZE = 50;

interface MediaItem {
  id: string;
  name: string;
  size: string;
  type: string;
  dimensions: string;
  url: string;
  uploadedAt: string;
}

const formatBytes = (bytes?: number) => {
  if (bytes == null) return "—";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
};

export function AdminMedia() {
  const t = useTranslations("admin");
  const tCommon = useTranslations("common");
  const locale = useLocale();
  const { user } = useAuthContext();
  const isSuperadmin = user?.role === "superadmin";
  const [mediaList, setMediaList] = useState<MediaItem[]>([]);
  const [hasMore, setHasMore] = useState(false);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [search, setSearch] = useState("");
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [previewItem, setPreviewItem] = useState<MediaItem | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [loadError, setLoadError] = useState(false);
  const [loadMoreError, setLoadMoreError] = useState(false);

  // Cursor chain is walked server-side, so we only need the next token. Refs
  // keep the observer callback stable and prevent overlapping page fetches.
  const cursorRef = useRef<string | undefined>(undefined);
  const hasMoreRef = useRef(false);
  const loadMoreInFlightRef = useRef(false);
  const requestIdRef = useRef(0);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  const loadPage = useCallback(
    async (mode: "reset" | "more") => {
      const isReset = mode === "reset";
      if (!isReset && (loadMoreInFlightRef.current || !hasMoreRef.current)) return;

      if (!isReset) loadMoreInFlightRef.current = true;
      const requestId = ++requestIdRef.current;

      if (isReset) {
        setLoading(true);
        setLoadError(false);
        setLoadMoreError(false);
      } else {
        setLoadingMore(true);
        setLoadMoreError(false);
      }

      try {
        const res = await adminService.getMedia({
          cursor: isReset ? undefined : cursorRef.current,
          pageSize: MEDIA_PAGE_SIZE,
        });
        if (requestId !== requestIdRef.current) return;

        const rows = res?.data ?? [];
        const mapped: MediaItem[] = rows.map((m: AdminMedia) => ({
          id: m.id,
          name: m.filename || m.objectKey || m.id,
          size: formatBytes(m.fileSize ?? m.size),
          type: (m.contentType || "").includes("image") ? "Image" : (m.contentType || "—"),
          dimensions: "—",
          url: m.url || m.publicUrl || "",
          uploadedAt: formatDateTime(m.createdAt, locale) || "—",
        }));

        setMediaList((previous) => {
          if (isReset) return mapped;
          const seen = new Set(previous.map((item) => item.id));
          return [...previous, ...mapped.filter((item) => !seen.has(item.id))];
        });

        const meta = res?.meta;
        const nextCursor =
          typeof meta?.nextCursor === "string" && meta.nextCursor ? meta.nextCursor : undefined;
        cursorRef.current = nextCursor;
        const more = Boolean(meta?.hasMore && nextCursor);
        hasMoreRef.current = more;
        setHasMore(more);
      } catch {
        if (requestId !== requestIdRef.current) return;
        if (isReset) {
          setLoadError(true);
        } else {
          setLoadMoreError(true);
        }
      } finally {
        if (!isReset) loadMoreInFlightRef.current = false;
        if (requestId === requestIdRef.current) {
          setLoading(false);
          setLoadingMore(false);
        }
      }
    },
    [locale],
  );

  // Reload the first page on mount and when the locale changes. The ref keeps
  // StrictMode's double-effect in development from firing two identical loads.
  const loadedLocaleRef = useRef<string | null>(null);
  useEffect(() => {
    if (loadedLocaleRef.current === locale) return;
    loadedLocaleRef.current = locale;
    cursorRef.current = undefined;
    hasMoreRef.current = false;
    loadMoreInFlightRef.current = false;
    void loadPage("reset");
  }, [locale, loadPage]);

  // Infinite scroll: prefetch well before the sentinel scrolls into view. The
  // effect re-arms as pages append so a short first page still fills the page.
  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel || !hasMore || loadMoreError) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          void loadPage("more");
        }
      },
      { rootMargin: "600px 0px" },
    );

    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [hasMore, loadMoreError, loadPage, mediaList.length]);

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    try {
      setIsUploading(true);
      await adminService.uploadMedia(file);
      notify.success(t("media.toast.uploaded"));
      await loadPage("reset");
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
        setMediaList((previous) => previous.filter((m) => m.id !== id));
        notify.success(t("media.toast.deleted"));
      } catch (err: any) {
        notify.error(apiErrorMessage(err, t("media.toast.deleteFailed")));
      }
    }
  };

  const filtered = mediaList.filter((m) => m.name.toLowerCase().includes(search.toLowerCase()));

  if (loadError) {
    return <ErrorState onRetry={() => loadPage("reset")} />;
  }

  return (
    <div className="space-y-5">
      <PageHeader
        title={t("media.title")}
        actions={
          <label className="inline-flex cursor-pointer items-center rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90">
            {isUploading ? t("media.uploading") : t("media.upload")}
            <input type="file" onChange={handleFileUpload} disabled={isUploading} className="hidden" accept="image/*" />
          </label>
        }
      />

      <div className="relative w-full sm:max-w-xs">
        <Search className="absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <input
          type="text"
          placeholder={t("media.searchPlaceholder")}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full rounded-md border border-border bg-background py-2 pe-3 ps-9 text-sm text-foreground placeholder:text-muted-foreground focus:outline-hidden focus:ring-2 focus:ring-ring/30"
        />
      </div>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
        {loading && mediaList.length === 0
          ? Array.from({ length: 10 }).map((_, index) => (
              <div
                key={`media-skeleton-${index}`}
                aria-hidden
                className="overflow-hidden rounded-lg border border-border bg-background"
              >
                <div className="aspect-square animate-pulse bg-secondary/40" />
                <div className="space-y-2 border-t border-border p-3">
                  <div className="h-3 w-3/4 animate-pulse rounded bg-secondary/60" />
                  <div className="h-2.5 w-1/3 animate-pulse rounded bg-secondary/60" />
                </div>
              </div>
            ))
          : filtered.map((item) => (
              <div key={item.id} className="group overflow-hidden rounded-lg border border-border bg-background">
                <div className="relative aspect-square overflow-hidden bg-secondary/20">
                  {item.url ? (
                    <img src={item.url} alt={item.name} className="h-full w-full object-cover" />
                  ) : (
                    <div className="flex h-full w-full items-center justify-center text-xs text-muted-foreground">—</div>
                  )}
                  <div className="absolute inset-0 flex items-center justify-center gap-1 bg-black/50 opacity-0 transition-opacity group-hover:opacity-100 focus-within:opacity-100">
                    <button
                      onClick={() => setPreviewItem(item)}
                      className="rounded bg-white/90 px-2 py-1 text-xs font-medium text-black hover:bg-white"
                    >
                      {t("media.preview")}
                    </button>
                    <button
                      onClick={() => handleCopyUrl(item.url, item.id)}
                      className="rounded bg-white/90 px-2 py-1 text-xs font-medium text-black hover:bg-white"
                    >
                      {copiedId === item.id ? t("media.copied") : t("media.copyLink")}
                    </button>
                    {isSuperadmin ? (
                      <button
                        onClick={() => handleDelete(item.id)}
                        className="rounded bg-white/90 px-2 py-1 text-xs font-medium text-red-600 hover:bg-white"
                      >
                        {t("actions.delete")}
                      </button>
                    ) : null}
                  </div>
                </div>
                <div className="border-t border-border p-3">
                  <div className="truncate text-xs font-medium text-foreground">{item.name}</div>
                  <div className="mt-0.5 flex justify-between gap-2 text-xs text-muted-foreground">
                    <span>{item.size}</span>
                    <span className="truncate">{item.type}</span>
                  </div>
                </div>
              </div>
            ))}
      </div>

      {!loading && !loadingMore && filtered.length === 0 ? (
        <div className="rounded-lg border border-dashed border-border px-4 py-10 text-center text-sm text-muted-foreground">
          {mediaList.length === 0 ? t("media.empty") : t("media.noResults")}
        </div>
      ) : null}

      <div ref={sentinelRef} aria-hidden className="h-px w-full" />

      <div role="status" aria-live="polite">
        {loadingMore ? (
          <div className="flex items-center justify-center gap-2 py-4 text-sm text-muted-foreground">
            <Loader2 className="size-4 animate-spin" />
            <span>{t("media.loadingMore")}</span>
          </div>
        ) : null}

        {loadMoreError ? (
          <div className="flex items-center justify-center gap-3 py-4 text-sm text-muted-foreground">
            <span>{t("media.loadMoreFailed")}</span>
            <button
              type="button"
              onClick={() => loadPage("more")}
              className="rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground transition-colors hover:bg-secondary"
            >
              {tCommon("errors.retry")}
            </button>
          </div>
        ) : null}

        {!loading && !loadingMore && !hasMore && !loadMoreError && mediaList.length > 0 && filtered.length > 0 ? (
          <div className="py-4 text-center text-xs text-muted-foreground">{t("media.allLoaded")}</div>
        ) : null}
      </div>

      {previewItem ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="max-h-[85vh] w-full max-w-lg space-y-4 overflow-y-auto rounded-lg border border-border bg-background p-5 shadow-xl">
            <div className="flex items-start justify-between gap-4 border-b border-border pb-3">
              <div className="min-w-0">
                <div className="text-xs font-medium text-muted-foreground">{t("media.preview")}</div>
                <h3 className="mt-0.5 truncate text-sm font-semibold text-foreground">{previewItem.name}</h3>
              </div>
              <button
                onClick={() => setPreviewItem(null)}
                className="shrink-0 text-muted-foreground hover:text-foreground"
                aria-label={t("actions.close")}
              >
                <X className="size-4" />
              </button>
            </div>
            <div className="space-y-4">
              <div className="aspect-square overflow-hidden rounded-md border border-border">
                {previewItem.url ? (
                  <img src={previewItem.url} alt={previewItem.name} className="h-full w-full bg-black/5 object-contain" />
                ) : (
                  <div className="flex h-full w-full items-center justify-center text-sm text-muted-foreground">—</div>
                )}
              </div>
              <dl className="grid grid-cols-1 gap-x-6 gap-y-4 text-sm sm:grid-cols-3">
                <div className="space-y-0.5">
                  <dt className="text-xs text-muted-foreground">{t("media.size")}</dt>
                  <dd className="font-mono text-foreground" dir="ltr">{previewItem.size}</dd>
                </div>
                <div className="space-y-0.5">
                  <dt className="text-xs text-muted-foreground">{t("media.dimensions")}</dt>
                  <dd className="font-mono text-foreground" dir="ltr">{previewItem.dimensions}</dd>
                </div>
                <div className="space-y-0.5">
                  <dt className="text-xs text-muted-foreground">{t("media.uploaded")}</dt>
                  <dd className="text-foreground">{previewItem.uploadedAt}</dd>
                </div>
              </dl>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
