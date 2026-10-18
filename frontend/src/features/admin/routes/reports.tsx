"use client";

import { useState, useEffect, useCallback } from "react";
import { Download, TrendingUp, ShoppingBag, Receipt, Ban } from "lucide-react";
import { useTranslations } from "next-intl";
import { adminService, type SalesReport, type ReportExportType } from "@/services/admin";
import { notify, apiErrorMessage } from "@/lib/notify";
import { ErrorState } from "@/components/ui/errorState";

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
    "px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground";

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("reports.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">{t("reports.subtitle")}</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <input
            type="date"
            value={dateFrom}
            max={dateTo}
            onChange={(e) => setDateFrom(e.target.value)}
            className={inputClass}
            aria-label={t("reports.from")}
          />
          <span className="text-muted-foreground text-xs">–</span>
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
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs">
          <div className="flex items-center gap-2 text-muted-foreground text-xs font-semibold uppercase tracking-wider">
            <TrendingUp className="size-4" /> {t("reports.revenue")}
          </div>
          <div className="mt-2 text-xl font-bold text-foreground">
            {loading || !totals ? "—" : money(totals.revenue)}
          </div>
        </div>
        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs">
          <div className="flex items-center gap-2 text-muted-foreground text-xs font-semibold uppercase tracking-wider">
            <ShoppingBag className="size-4" /> {t("reports.orders")}
          </div>
          <div className="mt-2 text-xl font-bold text-foreground">
            {loading || !totals ? "—" : totals.orders.toLocaleString()}
          </div>
        </div>
        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs">
          <div className="flex items-center gap-2 text-muted-foreground text-xs font-semibold uppercase tracking-wider">
            <Receipt className="size-4" /> {t("reports.aov")}
          </div>
          <div className="mt-2 text-xl font-bold text-foreground">
            {loading || !totals ? "—" : money(Math.round(totals.averageOrderValue))}
          </div>
        </div>
        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs">
          <div className="flex items-center gap-2 text-muted-foreground text-xs font-semibold uppercase tracking-wider">
            <Ban className="size-4" /> {t("reports.cancelled")}
          </div>
          <div className="mt-2 text-xl font-bold text-foreground">
            {loading || !totals ? "—" : totals.cancelledOrders.toLocaleString()}
          </div>
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs">
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-sm font-bold text-foreground">{t("reports.export")}</h2>
          <span className="text-[11px] text-muted-foreground">{t("reports.exportHint")}</span>
        </div>
        <div className="flex flex-wrap gap-2">
          {exportTypes.map((type) => (
            <button
              key={type}
              onClick={() => handleExport(type)}
              disabled={exporting !== null}
              className="inline-flex items-center gap-1.5 px-3 py-2 border border-border rounded-xl text-xs font-semibold text-foreground hover:bg-secondary/40 disabled:opacity-50 transition-colors"
            >
              <Download className="size-3.5" />
              {exporting === type ? t("reports.exporting") : t(`reports.types.${type}`)}
            </button>
          ))}
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <table className="w-full min-w-max text-start text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 ps-4 text-start">{t("reports.period")}</th>
                <th className="py-3.5 text-start">{t("reports.orders")}</th>
                <th className="py-3.5 pe-4 text-start">{t("reports.revenue")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {report?.series?.map((point) => (
                <tr key={point.period} className="hover:bg-secondary/20 transition-colors">
                  <td className="py-3 ps-4 font-mono text-foreground">{point.period}</td>
                  <td className="py-3 text-muted-foreground">{point.orders.toLocaleString()}</td>
                  <td className="py-3 pe-4 font-semibold text-foreground">{money(point.revenue)}</td>
                </tr>
              ))}
              {!loading && (report?.series?.length ?? 0) === 0 && (
                <tr>
                  <td colSpan={3} className="py-10 text-center text-muted-foreground">
                    {t("reports.empty")}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
