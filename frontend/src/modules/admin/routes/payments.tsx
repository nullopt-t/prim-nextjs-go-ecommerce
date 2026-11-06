import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback } from "react";
import { useRouter } from "@/shared/i18n/navigation";
import { Search } from "lucide-react";
import { adminService } from "@/modules/admin/api/adminService";
import type { AdminPayment } from "@/modules/admin/types";
import { useTranslations } from "next-intl";
import { useAuthContext } from "@/modules/auth/model/authContext";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";

export function AdminPayments() {
  const navigate = useRouter();
  const t = useTranslations("admin");
  const { user } = useAuthContext();
  const isSuperadmin = user?.role === "superadmin";

  const [payments, setPayments] = useState<AdminPayment[]>([]);
  const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [exporting, setExporting] = useState(false);
  const [loadError, setLoadError] = useState(false);
  const { cursor, hasPrev, next, prev, reset, page } = useCursorPagination();

  const loadPayments = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);
      const res = await adminService.getPayments({
        cursor: cursor || undefined,
        pageSize: 50,
        dateFrom: dateFrom || undefined,
        dateTo: dateTo || undefined,
      });
      const list = res?.data ?? [];
      setPayments(list.map((p: AdminPayment) => ({
        id: p.id,
        orderId: p.orderId,
        customerEmail: p.customerEmail ? p.customerEmail.split("@")[0] : "—",
        email: p.customerEmail || "—",
        provider: p.provider ? p.provider.toUpperCase() : "—",
        amount: (p.amount || 0) / 100,
        fee: p.fee != null ? (Number(p.fee) || 0) / 100 : undefined,
        status: p.status === "captured" ? "Succeeded" : (p.status ? p.status.charAt(0).toUpperCase() + p.status.slice(1) : "Unknown"),
        currency: p.currency,
        createdAt: p.createdAt,
      })));
      if (res?.meta) setMeta(res.meta as CursorMeta);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [cursor, dateFrom, dateTo]);

  useEffect(() => {
    loadPayments();
  }, [loadPayments]);

  const handleExport = async () => {
    try {
      setExporting(true);
      await adminService.exportReport("payments", {
        dateFrom: dateFrom || undefined,
        dateTo: dateTo || undefined,
      });
      notify.success(t("reports.exportStarted", { type: t("reports.types.payments") }));
    } catch (err) {
      notify.error(apiErrorMessage(err, t("reports.exportFailed")));
    } finally {
      setExporting(false);
    }
  };

  const handleIssueRefund = async (pay: AdminPayment) => {
    if (confirm(t("payments.confirmRefund", { id: pay.id }))) {
      try {
        await adminService.refundPayment(pay.id);
        notify.success(t("payments.toast.refundIssued", { customer: pay.customerEmail || "—" }));
        loadPayments();
      } catch (err) {
        notify.error(apiErrorMessage(err, t("payments.toast.refundFailed")));
      }
    }
  };

  const handleMarkPaid = async (pay: AdminPayment) => {
    if (confirm(t("payments.confirmMarkPaid", { id: pay.id }))) {
      try {
        await adminService.updatePaymentStatus(pay.id, "captured");
        notify.success(t("payments.toast.markedPaid", { customer: pay.customerEmail || "—" }));
        loadPayments();
      } catch (err) {
        notify.error(apiErrorMessage(err, t("payments.toast.statusFailed")));
      }
    }
  };

  const filtered = payments.filter(
    (p) =>
      p.id.toLowerCase().includes(search.toLowerCase()) ||
      p.orderId.toLowerCase().includes(search.toLowerCase()) ||
      p.customerEmail?.toLowerCase().includes(search.toLowerCase())
  );

  if (loadError) {
    return <ErrorState onRetry={loadPayments} />;
  }

  return (
    <div className="space-y-5">
      <PageHeader
        title={t("payments.title")}
        actions={
          <button
            onClick={handleExport}
            disabled={exporting}
            className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground transition-colors hover:bg-secondary disabled:opacity-50"
          >
            {t("actions.export")}
          </button>
        }
      />

      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div className="relative w-full sm:max-w-xs">
          <Search className="absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder={t("payments.searchPlaceholder")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-md border border-border bg-background py-2 pe-3 ps-9 text-sm text-foreground placeholder:text-muted-foreground focus:outline-hidden focus:ring-2 focus:ring-ring/30"
          />
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <input
            type="date"
            value={dateFrom}
            max={dateTo || undefined}
            onChange={(e) => {
              setDateFrom(e.target.value);
              reset();
            }}
            className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
            aria-label={t("reports.from")}
          />
          <span className="text-sm text-muted-foreground">–</span>
          <input
            type="date"
            value={dateTo}
            min={dateFrom || undefined}
            onChange={(e) => {
              setDateTo(e.target.value);
              reset();
            }}
            className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
            aria-label={t("reports.to")}
          />
        </div>
      </div>

      <div className="overflow-hidden rounded-lg border border-border bg-background">
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
            <thead className="text-xs text-muted-foreground">
              <tr className="border-b border-border">
                <th className="py-2.5 ps-4 text-start font-medium">{t("payments.colTxn")}</th>
                <th className="py-2.5 text-start font-medium">{t("payments.colCustomer")}</th>
                <th className="py-2.5 text-start font-medium">{t("payments.colMethod")}</th>
                <th className="py-2.5 text-start font-medium">{t("payments.colGross")}</th>
                <th className="py-2.5 text-start font-medium">{t("payments.colFee")}</th>
                <th className="py-2.5 text-start font-medium">{t("payments.colStatus")}</th>
                <th className="py-2.5 pe-4 text-end font-medium">{t("payments.colActions")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {loading && filtered.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-10 text-center text-sm text-muted-foreground">…</td>
                </tr>
              ) : filtered.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-10 text-center text-sm text-muted-foreground">
                    {t("payments.empty")}
                  </td>
                </tr>
              ) : (
                filtered.map((pay) => (
                  <tr key={pay.id} className="transition-colors hover:bg-secondary/30">
                    <td className="py-3 ps-4 font-mono text-xs text-muted-foreground">{pay.id}</td>
                    <td className="py-3 font-medium text-foreground">{pay.customerEmail}</td>
                    <td className="py-3 text-xs text-muted-foreground">{pay.provider || "—"}</td>
                    <td className="py-3 font-medium text-foreground tabular-nums">
                      {pay.amount.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                      {pay.currency ? ` ${pay.currency}` : ""}
                    </td>
                    <td className="py-3 text-xs text-muted-foreground tabular-nums">
                      {pay.fee == null
                        ? "—"
                        : `${pay.fee.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}${pay.currency ? ` ${pay.currency}` : ""}`}
                    </td>
                    <td className="py-3">
                      <StatusPill tone={statusTone(pay.status)}>
                        {t(`status.${String(pay.status).toLowerCase()}`) || pay.status}
                      </StatusPill>
                    </td>
                    <td className="py-3 pe-4 text-end">
                      <div className="inline-flex items-center gap-1">
                        {pay.orderId ? (
                          <button
                            onClick={() => navigate.push(`/admin/orders/${pay.orderId}`)}
                            className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                          >
                            {t("payments.viewOrder")}
                          </button>
                        ) : null}
                        {isSuperadmin && pay.status === "Pending" ? (
                          <button
                            onClick={() => handleMarkPaid(pay)}
                            className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                          >
                            {t("payments.markPaid")}
                          </button>
                        ) : null}
                        {isSuperadmin && pay.status === "Succeeded" ? (
                          <button
                            onClick={() => handleIssueRefund(pay)}
                            className="rounded px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                          >
                            {t("payments.issueRefund")}
                          </button>
                        ) : null}
                      </div>
                    </td>
                  </tr>
                ))
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
          shown={filtered.length}
        />
      </div>
    </div>
  );
}
