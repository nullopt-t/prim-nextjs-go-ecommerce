import { useState, useEffect, useCallback } from "react";
import { CheckCircle2, RefreshCcw, Search } from "lucide-react";
import { adminService } from "@/services/admin";
import { useTranslations } from "next-intl";
import { notify, apiErrorMessage } from "@/lib/notify";
import { ErrorState } from "@/components/ui/errorState";
import { CursorPager, useCursorPagination } from "@/components/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/lib/pagination";

export function AdminPayments() {
  const t = useTranslations("admin");
  const [payments, setPayments] = useState<any[]>([]);
  const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [loadError, setLoadError] = useState(false);
  const { cursor, hasPrev, next, prev, page } = useCursorPagination();

  const loadPayments = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);
      const res: any = await adminService.getPayments({ cursor: cursor || undefined, pageSize: 50 });
      const list = Array.isArray(res?.data) ? res.data : (Array.isArray(res) ? res : []);
      setPayments(list.map((p: any) => ({
        id: p.id,
        orderId: p.orderId,
        customer: p.customerEmail ? p.customerEmail.split("@")[0] : "Customer",
        email: p.customerEmail || "customer@example.com",
        gateway: p.provider ? p.provider.toUpperCase() : "STRIPE",
        amount: `EGP ${((p.amount || 0) / 100).toFixed(2)}`,
        fee: `EGP ${(((p.amount || 0) * 0.029 + 30) / 100).toFixed(2)}`,
        status: p.status === "captured" ? "Succeeded" : (p.status ? p.status.charAt(0).toUpperCase() + p.status.slice(1) : "Pending")
      })));
      if (res?.meta) setMeta(res.meta);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [cursor]);

  useEffect(() => {
    loadPayments();
  }, [loadPayments]);

  const handleIssueRefund = async (pay: any) => {
    if (confirm(t("payments.confirmRefund", { id: pay.id }))) {
      try {
        await adminService.refundPayment(pay.id);
        notify.success(t("payments.toast.refundIssued", { customer: pay.customer }));
        loadPayments();
      } catch (err: any) {
        notify.error(apiErrorMessage(err, t("payments.toast.refundFailed")));
      }
    }
  };

  const handleMarkPaid = async (pay: any) => {
    if (confirm(t("payments.confirmMarkPaid", { id: pay.id }))) {
      try {
        await adminService.updatePaymentStatus(pay.id, "captured");
        notify.success(t("payments.toast.markedPaid", { customer: pay.customer }));
        loadPayments();
      } catch (err: any) {
        notify.error(apiErrorMessage(err, t("payments.toast.statusFailed")));
      }
    }
  };

  const filtered = payments.filter(
    (p) =>
      p.id.toLowerCase().includes(search.toLowerCase()) ||
      p.orderId.toLowerCase().includes(search.toLowerCase()) ||
      p.customer.toLowerCase().includes(search.toLowerCase())
  );

  if (loadError) {
    return <ErrorState onRetry={loadPayments} />;
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("payments.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t("payments.subtitle")}
          </p>
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs">
        <div className="relative max-w-sm">
          <Search className="absolute start-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder={t("payments.searchPlaceholder")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full ps-10 pe-4 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
          />
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <table className="w-full text-start text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 ps-4">{t("payments.colTxn")}</th>
                <th className="py-3.5">{t("payments.colOrder")}</th>
                <th className="py-3.5">{t("payments.colCustomer")}</th>
                <th className="py-3.5">{t("payments.colMethod")}</th>
                <th className="py-3.5">{t("payments.colGross")}</th>
                <th className="py-3.5">{t("payments.colFee")}</th>
                <th className="py-3.5">{t("payments.colStatus")}</th>
                <th className="py-3.5 pe-4 text-end">{t("payments.colActions")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {filtered.map((pay) => (
                <tr key={pay.id} className="hover:bg-secondary/20 transition-colors">
                  <td className="py-3.5 ps-4 font-mono font-bold text-foreground">{pay.id}</td>
                  <td className="py-3.5 font-mono text-muted-foreground">{pay.orderId}</td>
                  <td className="py-3.5 font-semibold text-foreground">{pay.customer}</td>
                  <td className="py-3.5 text-muted-foreground">{pay.gateway}</td>
                  <td className="py-3.5 font-bold text-foreground">{pay.amount}</td>
                  <td className="py-3.5 text-muted-foreground">{pay.fee}</td>
                  <td className="py-3.5 text-foreground">
                    {t(`status.${String(pay.status).toLowerCase()}`) || pay.status}
                  </td>
                  <td className="py-3.5 pe-4 text-end space-x-2">
                    {pay.status === "Pending" && (
                      <button
                        onClick={() => handleMarkPaid(pay)}
                        className="px-2.5 py-1 text-[11px] font-semibold text-emerald-600 bg-emerald-500/10 hover:bg-emerald-500/20 rounded-lg transition-colors inline-flex items-center gap-1"
                      >
                        <CheckCircle2 className="size-3" />
                        <span>{t("payments.markPaid")}</span>
                      </button>
                    )}
                    {pay.status === "Succeeded" && (
                      <button
                        onClick={() => handleIssueRefund(pay)}
                        className="px-2.5 py-1 text-[11px] font-semibold text-destructive bg-destructive/10 hover:bg-destructive/20 rounded-lg transition-colors inline-flex items-center gap-1"
                      >
                        <RefreshCcw className="size-3" />
                        <span>{t("payments.issueRefund")}</span>
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
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
