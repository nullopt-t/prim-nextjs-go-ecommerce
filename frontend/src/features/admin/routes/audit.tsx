"use client";

import { ResponsiveTable } from "@/components/ui/responsiveTable";
import { useState, useEffect, useCallback } from "react";
import { Search, X, ScrollText } from "lucide-react";
import { adminService, type AuditLogEntry } from "@/services/admin";
import { useLocale, useTranslations } from "next-intl";
import { ErrorState } from "@/components/ui/errorState";
import { CursorPager, useCursorPagination } from "@/components/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/lib/pagination";
import { formatDateTime } from "@/lib/datetime";

interface AuditFilters {
  action: string;
  entityType: string;
  actorId: string;
  from: string;
  to: string;
}

const emptyFilters: AuditFilters = {
  action: "",
  entityType: "",
  actorId: "",
  from: "",
  to: "",
};

export function AdminAudit() {
  const t = useTranslations("admin");
  const locale = useLocale();
  const [entries, setEntries] = useState<AuditLogEntry[]>([]);
  const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [draft, setDraft] = useState<AuditFilters>(emptyFilters);
  const [applied, setApplied] = useState<AuditFilters>(emptyFilters);
  const [detail, setDetail] = useState<AuditLogEntry | null>(null);
  const { cursor, hasPrev, next, prev, reset, page } = useCursorPagination();

  const load = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);
      const res: any = await adminService.getAuditLogs({
        cursor: cursor || undefined,
        pageSize: 50,
        action: applied.action || undefined,
        entityType: applied.entityType || undefined,
        actorId: applied.actorId || undefined,
        from: applied.from ? `${applied.from}T00:00:00Z` : undefined,
        to: applied.to ? `${applied.to}T23:59:59Z` : undefined,
      });
      setEntries(Array.isArray(res?.data) ? res.data : []);
      if (res?.meta) setMeta(res.meta);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [cursor, applied]);

  useEffect(() => {
    load();
  }, [load]);

  const applyFilters = () => {
    reset();
    setApplied(draft);
  };

  const clearFilters = () => {
    reset();
    setDraft(emptyFilters);
    setApplied(emptyFilters);
  };

  const statusTone = (code: number) =>
    code >= 500
      ? "text-rose-600 bg-rose-500/10"
      : code >= 400
        ? "text-amber-600 bg-amber-500/10"
        : "text-emerald-600 bg-emerald-500/10";

  if (loadError) {
    return <ErrorState onRetry={load} />;
  }

  const inputClass =
    "w-full px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground";

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("audit.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t("audit.subtitle")}</p>
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs space-y-3">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
          <div className="relative sm:col-span-2">
            <Search className="absolute start-3 top-2.5 size-4 text-muted-foreground" />
            <input
              type="text"
              placeholder={t("audit.filterAction")}
              value={draft.action}
              onChange={(e) => setDraft((d) => ({ ...d, action: e.target.value }))}
              className={`${inputClass} ps-9`}
            />
          </div>
          <input
            type="text"
            placeholder={t("audit.filterEntityType")}
            value={draft.entityType}
            onChange={(e) => setDraft((d) => ({ ...d, entityType: e.target.value }))}
            className={inputClass}
          />
          <input
            type="text"
            placeholder={t("audit.filterActor")}
            value={draft.actorId}
            onChange={(e) => setDraft((d) => ({ ...d, actorId: e.target.value }))}
            className={`${inputClass} font-mono`}
          />
          <div className="flex flex-wrap items-center gap-2 sm:col-span-2 lg:col-span-4">
            <input
              type="date"
              value={draft.from}
              onChange={(e) => setDraft((d) => ({ ...d, from: e.target.value }))}
              className={`${inputClass} sm:w-auto`}
            />
            <span className="text-muted-foreground text-xs">–</span>
            <input
              type="date"
              value={draft.to}
              onChange={(e) => setDraft((d) => ({ ...d, to: e.target.value }))}
              className={`${inputClass} sm:w-auto`}
            />
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={applyFilters}
            className="px-4 py-2 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-colors"
          >
            {t("audit.apply")}
          </button>
          <button
            onClick={clearFilters}
            className="px-4 py-2 border border-border text-muted-foreground hover:text-foreground text-xs font-semibold rounded-xl transition-colors"
          >
            {t("audit.clear")}
          </button>
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full sm:min-w-max text-start text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 ps-4 text-start">{t("audit.colTime")}</th>
                <th className="py-3.5 text-start">{t("audit.colActor")}</th>
                <th className="py-3.5 text-start">{t("audit.colAction")}</th>
                <th className="py-3.5 text-start">{t("audit.colEntity")}</th>
                <th className="py-3.5 text-start">{t("audit.colMethod")}</th>
                <th className="py-3.5 text-start">{t("audit.colStatus")}</th>
                <th className="py-3.5 pe-4 text-start">{t("audit.colIp")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {entries.map((entry) => (
                <tr
                  key={entry.id}
                  onClick={() => setDetail(entry)}
                  className="hover:bg-secondary/20 transition-colors cursor-pointer"
                >
                  <td className="py-3.5 ps-4 text-muted-foreground whitespace-nowrap">
                    {formatDateTime(entry.createdAt, locale)}
                  </td>
                  <td className="py-3.5 font-semibold text-foreground">{entry.actorRole || "—"}</td>
                  <td className="py-3.5 font-mono font-semibold text-foreground">
                    <span dir="ltr">{entry.action}</span>
                  </td>
                  <td className="py-3.5 font-semibold text-foreground">
                    <span dir="ltr">{entry.entityType || "—"}</span>
                  </td>
                  <td className="py-3.5">
                    <span className="font-mono text-[10px] px-1.5 py-0.5 rounded bg-secondary/60 text-muted-foreground" dir="ltr">
                      {entry.method}
                    </span>
                  </td>
                  <td className="py-3.5">
                    <span
                      className={`font-mono font-bold text-[11px] px-1.5 py-0.5 rounded ${statusTone(entry.statusCode)}`}
                      dir="ltr"
                    >
                      {entry.statusCode}
                    </span>
                  </td>
                  <td className="py-3.5 pe-4 font-mono text-[10px] text-muted-foreground">
                    {entry.ipAddress ? <span dir="ltr">{entry.ipAddress}</span> : "—"}
                  </td>
                </tr>
              ))}
              {entries.length === 0 && !loading && (
                <tr>
                  <td colSpan={7} className="py-10 text-center text-muted-foreground">
                    <ScrollText className="size-6 mx-auto mb-2 opacity-50" />
                    {t("audit.empty")}
                  </td>
                </tr>
              )}
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
          shown={entries.length}
        />
      </div>

      {detail && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-background border border-border rounded-2xl max-w-2xl w-full max-h-[85vh] overflow-y-auto p-6 shadow-2xl space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="font-bold text-sm text-foreground font-mono" dir="ltr">{detail.action}</h3>
              <button
                onClick={() => setDetail(null)}
                className="text-muted-foreground hover:text-foreground"
                aria-label={t("actions.close")}
              >
                <X className="size-5" />
              </button>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs text-muted-foreground">
              <div>
                <strong>{t("audit.colTime")}:</strong> {formatDateTime(detail.createdAt, locale)}
              </div>
              <div>
                <strong>{t("audit.colActor")}:</strong> {detail.actorRole}{" "}
                <span dir="ltr">({detail.actorId || "—"})</span>
              </div>
              <div>
                <strong>{t("audit.colEntity")}:</strong>{" "}
                <span dir="ltr">
                  {detail.entityType} / {detail.entityId || "—"}
                </span>
              </div>
              <div>
                <strong>{t("audit.colStatus")}:</strong> <span dir="ltr">{detail.statusCode}</span>
              </div>
              <div className="col-span-2">
                <strong>{t("audit.colMethod")}:</strong>{" "}
                <span dir="ltr">
                  {detail.method} {detail.path}
                </span>
              </div>
              <div className="col-span-2">
                <strong>{t("audit.colIp")}:</strong>{" "}
                <span dir="ltr">
                  {detail.ipAddress} — {detail.userAgent}
                </span>
              </div>
            </div>
            {detail.after && (
              <div>
                <div className="text-xs font-semibold text-foreground mb-1">{t("audit.payload")}</div>
                <pre className="text-[11px] bg-secondary/40 border border-border rounded-xl p-3 overflow-x-auto">
                  {JSON.stringify(detail.after, null, 2)}
                </pre>
              </div>
            )}
            {detail.metadata && (
              <div>
                <div className="text-xs font-semibold text-foreground mb-1">{t("audit.metadata")}</div>
                <pre className="text-[11px] bg-secondary/40 border border-border rounded-xl p-3 overflow-x-auto">
                  {JSON.stringify(detail.metadata, null, 2)}
                </pre>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
