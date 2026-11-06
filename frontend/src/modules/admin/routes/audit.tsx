"use client";

import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback } from "react";
import { Search, X } from "lucide-react";
import { adminService, type AuditLogEntry } from "@/modules/admin/api/adminService";
import { useLocale, useTranslations } from "next-intl";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, type StatusTone } from "@/shared/ui/statusPill";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";
import { formatDateTime } from "@/shared/lib/datetime";

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

  const httpTone = (code: number): StatusTone =>
    code >= 500 ? "danger" : code >= 400 ? "warning" : "success";

  if (loadError) {
    return <ErrorState onRetry={load} />;
  }

  const inputClass =
    "w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-hidden focus:ring-2 focus:ring-ring/30";

  return (
    <div className="space-y-5">
      <PageHeader title={t("audit.title")} />

      <div className="space-y-3 rounded-lg border border-border bg-background p-4">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <div className="relative sm:col-span-2">
            <Search className="absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
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
              aria-label={t("reports.from")}
            />
            <span className="text-sm text-muted-foreground">–</span>
            <input
              type="date"
              value={draft.to}
              onChange={(e) => setDraft((d) => ({ ...d, to: e.target.value }))}
              className={`${inputClass} sm:w-auto`}
              aria-label={t("reports.to")}
            />
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={applyFilters}
            className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
          >
            {t("audit.apply")}
          </button>
          <button
            onClick={clearFilters}
            className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground transition-colors hover:bg-secondary"
          >
            {t("audit.clear")}
          </button>
        </div>
      </div>

      <div className="overflow-hidden rounded-lg border border-border bg-background">
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
            <thead className="text-xs text-muted-foreground">
              <tr className="border-b border-border">
                <th className="py-2.5 ps-4 text-start font-medium">{t("audit.colTime")}</th>
                <th className="py-2.5 text-start font-medium">{t("audit.colActor")}</th>
                <th className="py-2.5 text-start font-medium">{t("audit.colAction")}</th>
                <th className="py-2.5 text-start font-medium">{t("audit.colEntity")}</th>
                <th className="py-2.5 text-start font-medium">{t("audit.colMethod")}</th>
                <th className="py-2.5 text-start font-medium">{t("audit.colStatus")}</th>
                <th className="py-2.5 pe-4 text-start font-medium">{t("audit.colIp")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {entries.map((entry) => (
                <tr
                  key={entry.id}
                  onClick={() => setDetail(entry)}
                  className="cursor-pointer transition-colors hover:bg-secondary/30"
                >
                  <td className="whitespace-nowrap py-3 ps-4 text-xs text-muted-foreground">
                    {formatDateTime(entry.createdAt, locale) || "—"}
                  </td>
                  <td className="py-3 text-sm font-medium text-foreground">{entry.actorRole || "—"}</td>
                  <td className="py-3 font-mono text-xs font-medium text-foreground">
                    <span dir="ltr">{entry.action}</span>
                  </td>
                  <td className="py-3 text-sm text-foreground">
                    <span dir="ltr">{entry.entityType || "—"}</span>
                  </td>
                  <td className="py-3">
                    <span className="rounded border border-border bg-secondary/50 px-1.5 py-0.5 font-mono text-xs text-muted-foreground" dir="ltr">
                      {entry.method}
                    </span>
                  </td>
                  <td className="py-3">
                    <StatusPill tone={httpTone(entry.statusCode)}>
                      <span dir="ltr">{entry.statusCode}</span>
                    </StatusPill>
                  </td>
                  <td className="py-3 pe-4 font-mono text-xs text-muted-foreground">
                    {entry.ipAddress ? <span dir="ltr">{entry.ipAddress}</span> : "—"}
                  </td>
                </tr>
              ))}
              {entries.length === 0 && !loading ? (
                <tr>
                  <td colSpan={7} className="py-10 text-center text-sm text-muted-foreground">
                    {t("audit.empty")}
                  </td>
                </tr>
              ) : null}
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

      {detail ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="max-h-[85vh] w-full max-w-2xl space-y-4 overflow-y-auto rounded-lg border border-border bg-background p-5 shadow-xl">
            <div className="flex items-start justify-between gap-4 border-b border-border pb-3">
              <div className="min-w-0">
                <div className="text-xs font-medium text-muted-foreground">{t("audit.title")}</div>
                <h3 className="mt-0.5 truncate font-mono text-sm font-semibold text-foreground" dir="ltr">
                  {detail.action}
                </h3>
              </div>
              <button
                onClick={() => setDetail(null)}
                className="shrink-0 text-muted-foreground hover:text-foreground"
                aria-label={t("actions.close")}
              >
                <X className="size-4" />
              </button>
            </div>

            <dl className="grid grid-cols-1 gap-x-6 gap-y-4 text-sm sm:grid-cols-2">
              <div className="space-y-0.5">
                <dt className="text-xs text-muted-foreground">{t("audit.colTime")}</dt>
                <dd className="text-foreground">{formatDateTime(detail.createdAt, locale) || "—"}</dd>
              </div>
              <div className="space-y-0.5">
                <dt className="text-xs text-muted-foreground">{t("audit.colActor")}</dt>
                <dd className="text-foreground" dir="ltr">
                  {detail.actorRole}{" "}
                  <span className="font-mono text-muted-foreground">({detail.actorId || "—"})</span>
                </dd>
              </div>
              <div className="space-y-0.5">
                <dt className="text-xs text-muted-foreground">{t("audit.colEntity")}</dt>
                <dd className="font-mono text-foreground" dir="ltr">
                  {detail.entityType} / {detail.entityId || "—"}
                </dd>
              </div>
              <div className="space-y-0.5">
                <dt className="text-xs text-muted-foreground">{t("audit.colStatus")}</dt>
                <dd>
                  <StatusPill tone={httpTone(detail.statusCode)}>
                    <span dir="ltr">{detail.statusCode}</span>
                  </StatusPill>
                </dd>
              </div>
              <div className="space-y-0.5 sm:col-span-2">
                <dt className="text-xs text-muted-foreground">{t("audit.colMethod")}</dt>
                <dd className="break-all font-mono text-foreground" dir="ltr">
                  {detail.method} {detail.path}
                </dd>
              </div>
              <div className="space-y-0.5 sm:col-span-2">
                <dt className="text-xs text-muted-foreground">{t("audit.colIp")}</dt>
                <dd className="break-all font-mono text-foreground" dir="ltr">
                  {detail.ipAddress} — {detail.userAgent}
                </dd>
              </div>
            </dl>

            {detail.after || detail.metadata ? (
              <div className="space-y-4">
                {detail.after ? (
                  <div>
                    <div className="mb-1 text-xs font-medium text-foreground">{t("audit.payload")}</div>
                    <pre className="overflow-x-auto rounded-md border border-border bg-secondary/40 p-3 font-mono text-xs" dir="ltr">
                      {JSON.stringify(detail.after, null, 2)}
                    </pre>
                  </div>
                ) : null}
                {detail.metadata ? (
                  <div>
                    <div className="mb-1 text-xs font-medium text-foreground">{t("audit.metadata")}</div>
                    <pre className="overflow-x-auto rounded-md border border-border bg-secondary/40 p-3 font-mono text-xs" dir="ltr">
                      {JSON.stringify(detail.metadata, null, 2)}
                    </pre>
                  </div>
                ) : null}
              </div>
            ) : null}
          </div>
        </div>
      ) : null}
    </div>
  );
}
