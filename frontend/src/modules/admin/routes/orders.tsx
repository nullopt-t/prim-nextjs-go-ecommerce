import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { useState, useEffect, useCallback } from "react";
import { useRouter } from "@/shared/i18n/navigation";
import { useLocale, useTranslations } from "next-intl";
import { Search, Eye } from "lucide-react";
import { adminService } from "@/modules/admin/api/adminService";
import type { Order, OrderItem } from "@/modules/orders/types";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { formatDateTime } from "@/shared/lib/datetime";
import { ErrorState } from "@/shared/ui/errorState";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { CursorPager, useCursorPagination } from "@/shared/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/shared/lib/pagination";

interface OrderView {
  id: string;
  customer: string;
  email: string;
  date: string;
  total: string;
  discount: string | null;
  couponCode: string | null;
  status: string;
  fulfillmentStatus: string;
  paymentStatus: string;
  items: OrderItem[];
  history: { date: string; note: string }[];
}

export function AdminOrders() {
  const navigate = useRouter();
  const t = useTranslations("admin");
  const locale = useLocale();
  const [orders, setOrders] = useState<OrderView[]>([]);
  const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [dateFrom, setDateFrom] = useState("");
  const [dateTo, setDateTo] = useState("");
  const [exporting, setExporting] = useState(false);
  const [loadError, setLoadError] = useState(false);
  const { cursor, hasPrev, next, prev, reset, page } = useCursorPagination();

  const loadOrders = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);
      const res = await adminService.getOrders({
        cursor: cursor || undefined,
        pageSize: 50,
        status: statusFilter !== "all" ? statusFilter.toLowerCase() : undefined,
        dateFrom: dateFrom || undefined,
        dateTo: dateTo || undefined,
      });
      const list = res?.data ?? [];
      const formatted = list.map((so: Order) => ({
        id: so.id,
        customer: so.customerEmail ? so.customerEmail.split("@")[0] : "—",
        email: so.customerEmail || "—",
        date: formatDateTime(so.createdAt, locale),
        total: `EGP ${((so.totalAmount || 0) / 100).toFixed(2)}`,
        discount: so.discountAmount ? `EGP ${((so.discountAmount || 0) / 100).toFixed(2)}` : null,
        couponCode: so.couponCode || null,
        status: so.status ? so.status.toLowerCase() : "pending",
        fulfillmentStatus: so.status === "shipped" || so.status === "delivered" ? "Fulfilled" : "Unfulfilled",
        paymentStatus: so.paymentStatus
          ? String(so.paymentStatus).toLowerCase()
          : "unknown",
        shippingAddress: so.shippingAddress ? `${so.shippingAddress.street || ""}, ${so.shippingAddress.city || ""}` : "Online Order",
        paymentMethod: so.paymentMethod || "—",
        trackingNumber: so.trackingNumber || "",
        items: so.items || [],
        history: [
          { date: formatDateTime(so.createdAt, locale), note: "Order placed successfully." }
        ]
      }));
      setOrders(formatted);
      if (res?.meta) setMeta(res.meta);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [cursor, statusFilter, dateFrom, dateTo]);

  useEffect(() => {
    loadOrders();
  }, [loadOrders]);

  // On-spot refresh when a new order event is received in the admin console
  useEffect(() => {
    const handleNewOrder = () => {
      loadOrders();
    };

    window.addEventListener("admin-order-placed", handleNewOrder);
    return () => {
      window.removeEventListener("admin-order-placed", handleNewOrder);
    };
  }, [loadOrders]);

  const handleUpdateStatus = async (id: string, newStatus: string) => {
    try {
      await adminService.updateOrderStatus(id, newStatus.toLowerCase());
      notify.success(t("orders.toast.statusUpdated", { status: newStatus }));
      window.dispatchEvent(new CustomEvent("refresh-admin-stats"));
      loadOrders();
    } catch (err) {
      notify.error(apiErrorMessage(err, t("orders.toast.statusUpdateFailed")));
      loadOrders();
    }
  };

  const handleExport = async () => {
    try {
      setExporting(true);
      await adminService.exportReport("orders", {
        dateFrom: dateFrom || undefined,
        dateTo: dateTo || undefined,
        status: statusFilter !== "all" ? statusFilter : undefined,
      });
      notify.success(t("reports.exportStarted", { type: t("reports.types.orders") }));
    } catch (err) {
      notify.error(apiErrorMessage(err, t("reports.exportFailed")));
    } finally {
      setExporting(false);
    }
  };

  const filtered = orders.filter((o) => {
    const matchesSearch =
      o.id.toLowerCase().includes(search.toLowerCase()) ||
      o.customer.toLowerCase().includes(search.toLowerCase()) ||
      o.email.toLowerCase().includes(search.toLowerCase());
    const matchesStatus =
      statusFilter === "all" || o.status === statusFilter.toLowerCase();
    return matchesSearch && matchesStatus;
  });

  if (loadError) {
    return <ErrorState onRetry={loadOrders} />;
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("orders.title")}
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

      <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div className="relative w-full lg:max-w-xs">
          <Search className="absolute start-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder={t("orders.searchPlaceholder")}
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
          <select
            value={statusFilter}
            onChange={(e) => {
              setStatusFilter(e.target.value);
              reset();
            }}
            className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
          >
            <option value="all">{t("status.all")}</option>
            <option value="pending">{t("status.pending")}</option>
            <option value="processing">{t("status.processing")}</option>
            <option value="shipped">{t("status.shipped")}</option>
            <option value="delivered">{t("status.delivered")}</option>
            <option value="cancelled">{t("status.cancelled")}</option>
          </select>
        </div>
      </div>

      <div className="overflow-hidden rounded-lg border border-border bg-background">
        <div className="overflow-x-auto">
          <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
            <thead className="text-xs text-muted-foreground">
              <tr className="border-b border-border">
                <th className="py-2.5 ps-4 text-start font-medium">{t("orders.colOrderId")}</th>
                <th className="py-2.5 text-start font-medium">{t("orders.colDate")}</th>
                <th className="py-2.5 text-start font-medium">{t("orders.colCustomer")}</th>
                <th className="py-2.5 text-start font-medium">{t("orders.colTotal")}</th>
                <th className="py-2.5 text-start font-medium">{t("orders.colPayment")}</th>
                <th className="py-2.5 text-start font-medium">{t("orders.colFulfillment")}</th>
                <th className="py-2.5 pe-4 text-end font-medium">{t("orders.colAction")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {filtered.map((order) => (
                <tr key={order.id} className="hover:bg-secondary/20 transition-colors">
                  <td className="max-w-36 truncate py-3 ps-4 font-mono text-xs text-foreground">{order.id}</td>
                  <td className="py-3 text-xs text-muted-foreground">{order.date}</td>
                  <td className="py-3">
                    <div className="font-medium text-foreground">{order.customer}</div>
                    <div className="text-xs text-muted-foreground">{order.email}</div>
                  </td>
                  <td className="py-3">
                    <div className="font-medium text-foreground tabular-nums">{order.total}</div>
                    {order.couponCode ? (
                      <div className="mt-0.5 flex items-center gap-1">
                        <span className="rounded border border-border bg-secondary/60 px-1.5 py-0.5 font-mono text-xs text-muted-foreground">
                          {order.couponCode}
                        </span>
                        {order.discount ? (
                          <span className="text-xs text-muted-foreground line-through">
                            -{order.discount}
                          </span>
                        ) : null}
                      </div>
                    ) : null}
                  </td>
                  <td className="py-3">
                    <StatusPill tone={statusTone(order.paymentStatus)}>
                      {t(`status.${order.paymentStatus}`)}
                    </StatusPill>
                  </td>
                  <td className="py-3">
                    <select
                      value={order.status.charAt(0).toUpperCase() + order.status.slice(1)}
                      onChange={(e) => handleUpdateStatus(order.id, e.target.value.toLowerCase())}
                      disabled={order.status === "cancelled" || order.status === "delivered"}
                      className="px-2.5 py-1 bg-secondary/50 border border-border rounded-lg text-xs font-semibold text-foreground focus:outline-hidden disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {order.status === "pending" && (
                        <>
                          <option value="pending" disabled>{t("status.pending")}</option>
                          <option value="processing">{t("status.processing")}</option>
                          <option value="cancelled">{t("status.cancelled")}</option>
                        </>
                      )}
                      {order.status === "processing" && (
                        <>
                          <option value="processing" disabled>{t("status.processing")}</option>
                          <option value="shipped">{t("status.shipped")}</option>
                          <option value="cancelled">{t("status.cancelled")}</option>
                        </>
                      )}
                      {order.status === "shipped" && (
                        <>
                          <option value="shipped" disabled>{t("status.shipped")}</option>
                          <option value="delivered">{t("status.delivered")}</option>
                          <option value="cancelled">{t("status.cancelled")}</option>
                        </>
                      )}
                      {order.status === "delivered" && (
                        <option value="delivered">{t("status.deliveredFinal")}</option>
                      )}
                      {(order.status === "cancelled") && (
                        <option value="cancelled">{t("status.cancelledFinal")}</option>
                      )}
                    </select>
                  </td>
                  <td className="py-3 pe-4 text-end">
                    <button
                      onClick={() => navigate.push(`/admin/orders/${order.id}`)}
                      className="p-1.5 text-muted-foreground hover:text-accent-brand hover:bg-secondary rounded-lg transition-colors"
                      title={t("orders.manage")}
                    >
                      <Eye className="size-4" />
                    </button>
                  </td>
                </tr>
              ))}
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
