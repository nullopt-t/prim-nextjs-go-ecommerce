import { useState, useEffect, useCallback } from "react";
import { useRouter } from "@/i18n/navigation";
import { useTranslations } from "next-intl";
import { Search, Filter, Eye } from "lucide-react";
import { adminService } from "@/services/admin";
import { notify, apiErrorMessage } from "@/lib/notify";
import { formatDateTime } from "@/lib/datetime";
import { ErrorState } from "@/components/ui/errorState";
import { CursorPager, useCursorPagination } from "@/components/ui/cursorPager";
import { emptyCursorMeta, type CursorMeta } from "@/lib/pagination";

export function AdminOrders() {
  const navigate = useRouter();
  const t = useTranslations("admin");
  const [orders, setOrders] = useState<any[]>([]);
  const [meta, setMeta] = useState<CursorMeta>(emptyCursorMeta);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("All");
  const [loadError, setLoadError] = useState(false);
  const { cursor, hasPrev, next, prev, reset, page } = useCursorPagination();

  const loadOrders = useCallback(async () => {
    try {
      setLoading(true);
      setLoadError(false);
      const res = await adminService.getOrders({
        cursor: cursor || undefined,
        pageSize: 50,
        status: statusFilter !== "All" ? statusFilter.toLowerCase() : undefined
      });
      const list = Array.isArray(res?.data) ? res.data : (Array.isArray(res) ? res : []);
      const formatted = list.map((so: any) => ({
        id: so.id,
        customer: so.customerEmail ? so.customerEmail.split("@")[0] : "Customer",
        email: so.customerEmail || "customer@example.com",
        date: formatDateTime(so.createdAt),
        total: `EGP ${((so.totalAmount || 0) / 100).toFixed(2)}`,
        discount: so.discountAmount ? `EGP ${((so.discountAmount || 0) / 100).toFixed(2)}` : null,
        couponCode: so.couponCode || null,
        status: so.status ? so.status.toLowerCase() : "pending",
        fulfillmentStatus: so.status === "shipped" || so.status === "delivered" ? "Fulfilled" : "Unfulfilled",
        paymentStatus: so.status === "cancelled" ? "Canceled" : "Paid",
        items: Array.isArray(so.items) ? so.items.length : 0,
        shippingAddress: so.shippingAddress ? `${so.shippingAddress.street || ""}, ${so.shippingAddress.city || ""}` : "Online Order",
        paymentMethod: so.paymentMethod || "—",
        trackingNumber: so.trackingNumber || "",
        itemsList: so.items || [],
        history: [
          { date: formatDateTime(so.createdAt), note: "Order placed successfully." }
        ]
      }));
      setOrders(formatted);
      if (res?.meta) setMeta(res.meta);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [cursor, statusFilter]);

  useEffect(() => {
    loadOrders();
  }, [loadOrders]);

  const handleUpdateStatus = async (id: string, newStatus: string) => {
    try {
      await adminService.updateOrderStatus(id, newStatus.toLowerCase());
      notify.success(t("orders.toast.statusUpdated", { status: newStatus }));
      loadOrders();
    } catch (err: any) {
      notify.error(apiErrorMessage(err, t("orders.toast.statusUpdateFailed")));
      loadOrders();
    }
  };

  const filtered = orders.filter((o) => {
    const matchesSearch =
      o.id.toLowerCase().includes(search.toLowerCase()) ||
      o.customer.toLowerCase().includes(search.toLowerCase()) ||
      o.email.toLowerCase().includes(search.toLowerCase());
    const matchesStatus =
      statusFilter === "All" || o.status === statusFilter.toLowerCase();
    return matchesSearch && matchesStatus;
  });

  if (loadError) {
    return <ErrorState onRetry={loadOrders} />;
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("orders.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t("orders.subtitle")}
          </p>
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs flex flex-col md:flex-row gap-4 justify-between items-center">
        <div className="relative w-full md:w-80">
          <Search className="absolute start-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder={t("orders.searchPlaceholder")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full ps-10 pe-4 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
          />
        </div>

        <div className="flex items-center gap-3 w-full md:w-auto">
          <Filter className="size-4 text-muted-foreground" />
          <select
            value={statusFilter}
            onChange={(e) => {
              setStatusFilter(e.target.value);
              reset();
            }}
            className="px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
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

      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <table className="w-full text-start text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 ps-4">{t("orders.colOrderId")}</th>
                <th className="py-3.5">{t("orders.colDate")}</th>
                <th className="py-3.5">{t("orders.colCustomer")}</th>
                <th className="py-3.5">{t("orders.colTotal")}</th>
                <th className="py-3.5">{t("orders.colPayment")}</th>
                <th className="py-3.5">{t("orders.colFulfillment")}</th>
                <th className="py-3.5 pe-4 text-end">{t("orders.colAction")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {filtered.map((order) => (
                <tr key={order.id} className="hover:bg-secondary/20 transition-colors">
                  <td className="py-3.5 ps-4 font-mono font-bold text-foreground">{order.id}</td>
                  <td className="py-3.5 text-muted-foreground">{order.date}</td>
                  <td className="py-3.5">
                    <div className="font-semibold text-foreground">{order.customer}</div>
                    <div className="text-[10px] text-muted-foreground">{order.email}</div>
                  </td>
                  <td className="py-3.5">
                    <div className="font-bold text-foreground">{order.total}</div>
                    {order.couponCode && (
                      <div className="flex items-center gap-1 mt-0.5">
                        <span className="px-1.5 py-0.5 rounded text-[10px] font-mono font-semibold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
                          {order.couponCode}
                        </span>
                        {order.discount && (
                          <span className="text-[10px] text-muted-foreground line-through">
                            -{order.discount}
                          </span>
                        )}
                      </div>
                    )}
                  </td>
                  <td className="py-3.5 text-foreground">
                    {order.paymentStatus}
                  </td>
                  <td className="py-3.5">
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
                  <td className="py-3.5 pe-4 text-end">
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
