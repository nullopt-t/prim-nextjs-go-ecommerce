"use client";

import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback } from "react";
import { useTranslations } from "next-intl";
import { adminService, type SalesReport, type ReportExportType } from "@/modules/admin/api/adminService";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";

const isoDate = (d: Date) => d.toISOString().slice(0, 10);

const today = new Date();
const defaultTo = isoDate(today);
const defaultFrom = isoDate(new Date(today.getTime() - 29 * 24 * 60 * 60 * 1000));

const exportTypes: ReportExportType[] = ["orders", "payments", "customers", "inventory"];

function money(minor: number) {
  return `EGP ${(minor / 100).toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`;
}

export function AdminReports() {
  const t = useTranslations("admin");
  const [dateFrom, setDateFrom] = useState(defaultFrom);
  const [dateTo, setDateTo] = useState(defaultTo);
  const [groupBy, setGroupBy] = useState<"day" | "month">("day");
  const [report, setReport] = useState<SalesReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [exporting, setExporting] = useState<ReportExportType | null>(null);

  const load = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);
      const res = await adminService.getSalesReport({ dateFrom, dateTo, groupBy });
      setReport(res?.data ?? null);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [dateFrom, dateTo, groupBy]);

  useEffect(() => {
    load();
  }, [load]);

  const handleExport = async (type: ReportExportType) => {
    try {
      setExporting(type);
      await adminService.exportReport(type, { dateFrom, dateTo });
      notify.success(t("reports.exportStarted", { type: t(`reports.types.${type}`) }));
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("reports.exportFailed")));
    } finally {
      setExporting(null);
    }
  };

  if (loadError) {
    return <ErrorState onRetry={load} />;
  }

  const totals = report?.totals;
  const inputClass =
    "rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground";

  const metrics = [
    { label: t("reports.revenue"), value: loading || !totals ? "—" : money(totals.revenue) },
    { label: t("reports.orders"), value: loading || !totals ? "—" : totals.orders.toLocaleString() },
    { label: t("reports.aov"), value: loading || !totals ? "—" : money(Math.round(totals.averageOrderValue)) },
    { label: t("reports.cancelled"), value: loading || !totals ? "—" : totals.cancelledOrders.toLocaleString() },
  ];

  return (
    <div className="space-y-5">
      <PageHeader title={t("reports.title")} />

      <div className="flex flex-wrap items-center gap-2">
        <input
          type="date"
          value={dateFrom}
          max={dateTo}
          onChange={(e) => setDateFrom(e.target.value)}
          className={inputClass}
          aria-label={t("reports.from")}
        />
        <span className="text-sm text-muted-foreground">–</span>
        <input
          type="date"
          value={dateTo}
          min={dateFrom}
          onChange={(e) => setDateTo(e.target.value)}
          className={inputClass}
          aria-label={t("reports.to")}
        />
        <select
          value={groupBy}
          onChange={(e) => setGroupBy(e.target.value as "day" | "month")}
          className={inputClass}
          aria-label={t("reports.groupBy")}
        >
          <option value="day">{t("reports.groupByDay")}</option>
          <option value="month">{t("reports.groupByMonth")}</option>
        </select>
      </div>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
        {metrics.map((metric) => (
          <div key={metric.label} className="rounded-lg border border-border bg-background p-4">
            <div className="text-xs font-medium text-muted-foreground">{metric.label}</div>
            <div className="mt-2 text-2xl font-semibold tracking-tight text-foreground tabular-nums">
              {metric.value}
            </div>
          </div>
        ))}
      </div>

      <section className="rounded-lg border border-border bg-background p-4">
        <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-sm font-semibold text-foreground">{t("reports.export")}</h2>
          <span className="text-xs text-muted-foreground">{t("reports.exportHint")}</span>
        </div>
        <div className="flex flex-wrap gap-2">
          {exportTypes.map((type) => (
            <button
              key={type}
              onClick={() => handleExport(type)}
              disabled={exporting !== null}
              className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground transition-colors hover:bg-secondary disabled:opacity-50"
            >
              {exporting === type ? t("reports.exporting") : t(`reports.types.${type}`)}
            </button>
          ))}
        </div>
      </section>

      <section className="overflow-hidden rounded-lg border border-border bg-background">
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
            <thead className="text-xs text-muted-foreground">
              <tr className="border-b border-border">
                <th className="py-2.5 ps-4 text-start font-medium">{t("reports.period")}</th>
                <th className="py-2.5 text-start font-medium">{t("reports.orders")}</th>
                <th className="py-2.5 pe-4 text-end font-medium">{t("reports.revenue")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {report?.series?.map((point) => (
                <tr key={point.period} className="transition-colors hover:bg-secondary/30">
                  <td className="py-3 ps-4 font-mono text-xs text-foreground">{point.period}</td>
                  <td className="py-3 text-muted-foreground tabular-nums">{point.orders.toLocaleString()}</td>
                  <td className="py-3 pe-4 text-end font-medium text-foreground tabular-nums">{money(point.revenue)}</td>
                </tr>
              ))}
              {!loading && (report?.series?.length ?? 0) === 0 ? (
                <tr>
                  <td colSpan={3} className="py-10 text-center text-sm text-muted-foreground">
                    {t("reports.empty")}
                  </td>
                </tr>
              ) : null}
            </tbody>
          </ResponsiveTable>
        </div>
      </section>
    </div>
  );
}
