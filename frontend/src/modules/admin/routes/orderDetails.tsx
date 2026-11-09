import { useState, useEffect, useMemo } from "react";
import { useRouter } from "@/shared/i18n/navigation";
import { useParams } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { ArrowLeft } from "lucide-react";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { usePublicSettings } from "@/modules/settings/model/publicSettingsContext";
import { formatMoney } from "@/shared/lib/money";
import { adminService } from "@/modules/admin/api/adminService";
import type { AdminOrderDetail } from "@/modules/admin/api/adminService";
import type { OrderItem } from "@/modules/orders/types";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { formatDateTime } from "@/shared/lib/datetime";
import { ErrorState } from "@/shared/ui/errorState";
import { CustomSelect } from "@/shared/ui/customSelect";

interface OrderViewItem {
  id: string;
  name: string;
  price: string;
  quantity: number;
  sellerName?: string;
}

interface OrderViewHistory {
  id?: string;
  status: string;
  date: string;
  note: string | null;
}

interface OrderView {
  id: string;
  customer: string;
  email: string;
  date: string;
  total: string;
  discount: string | null;
  discountAmount: number;
  couponCode: string | null;
  couponId: string | null;
  status: string;
  fulfillmentStatus: string;
  paymentStatus: string;
  shippingAddress: string;
  paymentMethod: string;
  trackingNumber: string;
  carrier: string;
  trackingUrl: string;
  items: OrderViewItem[];
  history: OrderViewHistory[];
}

export function AdminOrderDetails() {
  const { id } = useParams();
  const navigate = useRouter();
  const t = useTranslations("admin");
  const { settings } = usePublicSettings();
  const currency = settings?.currency?.trim() ?? "";

  const locale = useLocale();
  const [order, setOrder] = useState<OrderView | null>(null);
  const [loading, setLoading] = useState(true);
  const [trackingNumber, setTrackingNumber] = useState("");
  const [carrier, setCarrier] = useState("");
  const [trackingUrl, setTrackingUrl] = useState("");
  const [noteInput, setNoteInput] = useState("");
  const [targetStatus, setTargetStatus] = useState("");
  const [updatingStatus, setUpdatingStatus] = useState(false);
  const [loadError, setLoadError] = useState(false);

  const fetchOrderFromApi = async () => {
    if (!id) return;
    try {
      setLoading(true);
      setLoadError(false);
      const res = await adminService.getOrderById(String(id));
      const remoteOrder: AdminOrderDetail = res?.data || res;
      if (remoteOrder) {
      const rawHistory = remoteOrder.statusHistory || remoteOrder.status_history || [];
      const formatted = {
          id: remoteOrder.id,
          customer: remoteOrder.customerEmail ? remoteOrder.customerEmail.split("@")[0] : "—",
          email: remoteOrder.customerEmail || "—",
          date: formatDateTime(remoteOrder.createdAt, locale),
          total: formatMoney((remoteOrder.totalAmount || 0) / 100, currency),
          discount: remoteOrder.discountAmount ? formatMoney((remoteOrder.discountAmount || 0) / 100, currency) : null,
          discountAmount: remoteOrder.discountAmount || 0,
          couponCode: remoteOrder.couponCode || null,
          couponId: remoteOrder.couponId || null,
          status: remoteOrder.status ? remoteOrder.status.toLowerCase() : "pending",
          fulfillmentStatus: remoteOrder.status === "shipped" || remoteOrder.status === "delivered" ? "Fulfilled" : "Unfulfilled",
          paymentStatus: remoteOrder.paymentStatus
            ? String(remoteOrder.paymentStatus).toLowerCase()
            : "unknown",
          shippingAddress: remoteOrder.shippingAddress ? `${remoteOrder.shippingAddress.street || ""}, ${remoteOrder.shippingAddress.city || ""}, ${remoteOrder.shippingAddress.state || ""} ${remoteOrder.shippingAddress.postalCode || ""}` : t("orderDetails.noShippingAddress"),
          paymentMethod: remoteOrder.paymentMethod || "—",
          trackingNumber: remoteOrder.trackingNumber || "",
          carrier: remoteOrder.carrier || "",
          trackingUrl: remoteOrder.trackingUrl || "",
          items: (Array.isArray(remoteOrder.items) ? remoteOrder.items : []).map((it: OrderItem) => {
            let itemName = t("orderDetails.orderItem");
            let sellerName: string | undefined;
            const snapshot: unknown = it.productSnapshot;
            if (typeof snapshot === "string") {
              try {
                const parsed = JSON.parse(snapshot);
                itemName = parsed.title || parsed.name || itemName;
                sellerName = parsed.sellerName;
              } catch {
                itemName = snapshot;
              }
            } else if (snapshot && typeof snapshot === "object") {
              const snap = snapshot as { title?: string; name?: string; sellerName?: string };
              itemName = snap.title || snap.name || itemName;
              sellerName = snap.sellerName;
            }
            return {
              id: it.id,
              name: itemName,
              sellerName,
              price: formatMoney((it.priceAtPurchase || 0) / 100, currency),
              quantity: it.quantity || 1
            };
          }),
          history:
            rawHistory.length > 0
              ? rawHistory.map((item, idx) => ({
                  id: item.id || `history-${idx}`,
                  status: String(item.status || remoteOrder.status || "pending").toLowerCase(),
                  date: formatDateTime(
                    item.createdAt || item.created_at || remoteOrder.createdAt,
                    locale
                  ),
                  note: item.notes || null,
                }))
              : [
                  {
                    id: "init-order-placed",
                    status: String(remoteOrder.status || "pending").toLowerCase(),
                    date: formatDateTime(remoteOrder.createdAt, locale),
                    note: t("orderDetails.orderPlaced"),
                  },
                ]
        };
        setOrder(formatted);
        setTargetStatus(formatted.status);
        setTrackingNumber(formatted.trackingNumber);
        setCarrier(formatted.carrier);
        setTrackingUrl(formatted.trackingUrl);
      }
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOrderFromApi();
  }, [id]);

  if (loading) {
    return <div className="p-8 text-center text-muted-foreground">{t("orderDetails.loading")}</div>;
  }

  if (loadError) {
    return <ErrorState onRetry={fetchOrderFromApi} />;
  }

  if (!order) {
    return <div className="p-8 text-center text-muted-foreground">{t("orderDetails.notFound")}</div>;
  }

  const handleUpdateFulfillment = async (newStatus: string) => {
    try {
      await adminService.updateOrderStatus(order.id, newStatus.toLowerCase());
      notify.success(t("orderDetails.toast.statusUpdated", { status: newStatus }));
      window.dispatchEvent(new CustomEvent("refresh-admin-stats"));
      fetchOrderFromApi();
    } catch (err) {
      notify.error(apiErrorMessage(err, t("orderDetails.toast.statusFailed")));
      fetchOrderFromApi();
    }
  };

  const handleSaveTracking = async () => {
    if (!order) return;
    if (!trackingNumber.trim()) {
      notify.error(t("orderDetails.toast.enterTracking"));
      return;
    }
    try {
      await adminService.updateOrderTracking(order.id, {
        trackingNumber: trackingNumber.trim(),
        carrier: carrier.trim() || undefined,
        trackingUrl: trackingUrl.trim() || undefined,
      });
      notify.success(t("orderDetails.toast.trackingSaved"));
      fetchOrderFromApi();
    } catch (err) {
      notify.error(apiErrorMessage(err, t("orderDetails.toast.trackingFailed")));
    }
  };

  const statusSelectOptions = useMemo(() => {
    if (!order) return [];
    const cur = order.status.toLowerCase();
    const transitions: Record<string, string[]> = {
      pending: ["processing", "cancelled"],
      processing: ["shipped", "cancelled"],
      shipped: ["delivered", "cancelled"],
      delivered: [],
      cancelled: [],
    };
    const allowed = [cur, ...(transitions[cur] ?? [])];
    return allowed.map((st) => ({
      value: st.charAt(0).toUpperCase() + st.slice(1),
      label: t(`status.${st}`, { defaultValue: st.charAt(0).toUpperCase() + st.slice(1) }),
    }));
  }, [order?.status, t]);

  const handleUpdateProgress = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!order) return;
    const nextSt = (targetStatus || order.status).toLowerCase();
    const note = noteInput.trim();

    if (nextSt === order.status && !note) {
      return;
    }

    setUpdatingStatus(true);
    try {
      await adminService.updateOrderStatus(order.id, nextSt, note || undefined);
      notify.success(
        nextSt !== order.status
          ? t("orderDetails.toast.statusUpdated", { status: t(`status.${nextSt}`, { defaultValue: nextSt }) })
          : t("orderDetails.toast.noteAdded")
      );
      setNoteInput("");
      window.dispatchEvent(new CustomEvent("refresh-admin-stats"));
      await fetchOrderFromApi();
    } catch (err) {
      notify.error(apiErrorMessage(err, t("orderDetails.toast.statusFailed")));
    } finally {
      setUpdatingStatus(false);
    }
  };

  const handleAddTimelineNote = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!noteInput.trim()) return;

    try {
      await adminService.updateOrderStatus(order.id, order.status.toLowerCase(), noteInput.trim());
      notify.success(t("orderDetails.toast.noteAdded"));
      setNoteInput("");
      fetchOrderFromApi();
    } catch (err) {
      notify.error(apiErrorMessage(err, t("orderDetails.toast.noteFailed")));
    }
  };

  return (
    <div className="w-full space-y-5">
      <div>
        <button
          onClick={() => navigate.push("/admin/orders")}
          className="mb-2 inline-flex items-center gap-1.5 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="size-3.5 rtl:rotate-180" />
          {t("orderDetails.back")}
        </button>
        <PageHeader
          title={`${t("orderDetails.title")}: ${order.id}`}
          actions={
            <StatusPill tone={statusTone(order.status)}>{t(`status.${order.status}`, { defaultValue: order.status })}</StatusPill>
          }
        />
        <p className="mt-1 text-xs text-muted-foreground">{t("orderDetails.placedOn", { date: order.date })}</p>
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <CustomSelect
        	value={order.status.charAt(0).toUpperCase() + order.status.slice(1)}
        	onChange={(value) => handleUpdateFulfillment(value.toLowerCase())}
        	disabled={order.status === "cancelled" || order.status === "delivered"}
        	options={
        		order.status === "pending"
        		? [
        		{ value: "Pending", label: t("status.pending"), disabled: true },
        		{ value: "Processing", label: t("status.processing") },
        		{ value: "Cancelled", label: t("status.cancelled") },
        		]
        		: order.status === "processing"
        		? [
        		{ value: "Processing", label: t("status.processing"), disabled: true },
        		{ value: "Shipped", label: t("status.shipped") },
        		{ value: "Cancelled", label: t("status.cancelled") },
        		]
        		: order.status === "shipped"
        		? [
        		{ value: "Shipped", label: t("status.shipped"), disabled: true },
        		{ value: "Delivered", label: t("status.delivered") },
        		{ value: "Cancelled", label: t("status.cancelled") },
        		]
        		: order.status === "delivered"
        		? [{ value: "Delivered", label: t("status.deliveredFinal") }]
        		: [{ value: "Cancelled", label: t("status.cancelledFinal") }]
        	}
        	ariaLabel={t("orderDetails.statusLabel")}
/>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        {/* Left Column: Line Items & History */}
        <div className="space-y-4 lg:col-span-2">
          {/* Purchased Items */}
          <section className="space-y-4 rounded-lg border border-border bg-background p-4">
            <h2 className="border-b border-border pb-3 text-sm font-semibold text-foreground">{t("orderDetails.lineItems", { count: order.items.length })}</h2>
            <div className="divide-y divide-border">
              {order.items.map((item: OrderViewItem) => (
                <div key={item.id} className="py-3 flex items-center justify-between text-xs">
                  <div>
                    <div className="font-semibold text-foreground flex items-center gap-2">
                      <span>{item.name}</span>
                      {item.sellerName && (
                        <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-secondary text-muted-foreground border border-border">
                          {item.sellerName}
                        </span>
                      )}
                    </div>
                    <div className="text-muted-foreground">{t("orderDetails.qty", { count: item.quantity })}</div>
                  </div>
                  <div className="font-bold text-foreground">{item.price}</div>
                </div>
              ))}
            </div>
            {order.couponCode && (
              <div className="pt-2 border-t border-border flex justify-between items-center text-xs">
                <span className="text-emerald-600 dark:text-emerald-400 font-semibold flex items-center gap-1.5">
                  <span>{t("orderDetails.couponApplied")}</span>
                  <span className="font-mono bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/30 uppercase tracking-wide">
                    {order.couponCode}
                  </span>
                </span>
                {order.discount && (
                  <span className="text-emerald-600 dark:text-emerald-400 font-bold">
                    -{order.discount}
                  </span>
                )}
              </div>
            )}
            <div className="pt-3 border-t border-border flex justify-between items-center text-sm font-bold">
              <span>{t("orderDetails.totalPaid")}</span>
              <span className="text-base font-semibold text-foreground tabular-nums">{order.total}</span>
            </div>
          </section>

          {/* Tracking Number Editor */}
          <section className="space-y-4 rounded-lg border border-border bg-background p-4">
            <h2 className="text-sm font-semibold text-foreground">{t("orderDetails.shipmentTracking")}</h2>
            <div className="space-y-3">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                <input
                  type="text"
                  placeholder={t("orderDetails.carrierPlaceholder")}
                  value={carrier}
                  onChange={(e) => setCarrier(e.target.value)}
                  className="rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
                />
                <input
                  type="text"
                  placeholder={t("orderDetails.trackingPlaceholder")}
                  value={trackingNumber}
                  onChange={(e) => setTrackingNumber(e.target.value)}
                  className="px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
                />
              </div>
              <input
                type="url"
                placeholder={t("orderDetails.trackingUrlPlaceholder")}
                value={trackingUrl}
                onChange={(e) => setTrackingUrl(e.target.value)}
                className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
              />
              <div className="flex justify-end">
                <button
                  onClick={handleSaveTracking}
                  className="rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
                >
                  {t("orderDetails.saveTracking")}
                </button>
              </div>
            </div>
          </section>

          {/* History Timeline & Progress Manager */}
          <section className="space-y-4 rounded-lg border border-border bg-background p-4">
            <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border pb-3">
              <div>
                <h2 className="text-sm font-semibold text-foreground">{t("orderDetails.timeline")}</h2>
                <p className="text-xs text-muted-foreground">{t("orderDetails.progressControlSubtitle")}</p>
              </div>
              <StatusPill tone={statusTone(order.status)}>
                {t(`status.${order.status}`, { defaultValue: order.status })}
              </StatusPill>
            </div>

            {/* Vertical timeline of all history events */}
            <div className="relative ps-6 space-y-4 before:absolute before:start-2.5 before:top-2 before:bottom-2 before:w-0.5 before:bg-border/70">
              {order.history.map((log: OrderViewHistory, idx: number) => {
                const isLatest = idx === order.history.length - 1;
                const statusKey = log.status || order.status;
                const statusLabel = t(`status.${statusKey}`, { defaultValue: statusKey });

                return (
                  <div key={log.id || idx} className="relative flex flex-col gap-1.5">
                    <div
                      className={`absolute -start-6 top-1.5 size-2.5 rounded-full border-2 transition-transform ${
                        isLatest
                          ? "border-accent-brand bg-accent-brand ring-4 ring-accent-brand/15 scale-110"
                          : "border-border bg-background"
                      }`}
                    />
                    <div className="flex flex-wrap items-center gap-2 text-xs">
                      <StatusPill tone={statusTone(statusKey)}>{statusLabel}</StatusPill>
                      <span className="text-[11px] text-muted-foreground">{log.date}</span>
                    </div>
                    {log.note && (
                      <div className="rounded-md bg-secondary/50 border border-border/60 p-2.5 text-xs text-foreground leading-relaxed whitespace-pre-wrap">
                        {log.note}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>

            {/* Progress Update & Timeline Message Action Card */}
            <form onSubmit={handleUpdateProgress} className="pt-4 border-t border-border space-y-3">
              <div className="flex flex-col sm:flex-row gap-3">
                <div className="w-full sm:w-52">
                  <label className="block text-xs font-medium text-foreground mb-1">
                    {t("orderDetails.targetStatus")}
                  </label>
                  <CustomSelect
                    value={targetStatus ? targetStatus.charAt(0).toUpperCase() + targetStatus.slice(1) : order.status.charAt(0).toUpperCase() + order.status.slice(1)}
                    onChange={(val) => setTargetStatus(val.toLowerCase())}
                    disabled={updatingStatus || order.status === "cancelled" || order.status === "delivered"}
                    options={statusSelectOptions}
                    ariaLabel={t("orderDetails.targetStatus")}
                  />
                </div>
                <div className="flex-1">
                  <label className="block text-xs font-medium text-foreground mb-1">
                    {t("orderDetails.statusMessage")}
                  </label>
                  <input
                    type="text"
                    placeholder={t("orderDetails.statusMessagePlaceholder")}
                    value={noteInput}
                    onChange={(e) => setNoteInput(e.target.value)}
                    disabled={updatingStatus}
                    className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:border-accent-brand focus:outline-none"
                  />
                </div>
              </div>

              <div className="flex justify-end gap-2">
                <button
                  type="submit"
                  disabled={updatingStatus || (!noteInput.trim() && targetStatus === order.status)}
                  className="rounded-md bg-accent-brand text-white px-4 py-2 text-sm font-medium hover:bg-accent-brand/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-xs"
                >
                  {updatingStatus ? t("orderDetails.updating") : t("orderDetails.updateProgressBtn")}
                </button>
              </div>
            </form>
          </section>
        </div>

        {/* Right Column: Customer & Payment Details */}
        <div className="space-y-4">
          <section className="space-y-3 rounded-lg border border-border bg-background p-4">
            <h2 className="text-sm font-semibold text-foreground">{t("orderDetails.customer")}</h2>
            <div className="text-xs space-y-1">
              <div className="font-semibold text-foreground">{order.customer}</div>
              <div className="text-muted-foreground">{order.email}</div>
            </div>
          </section>

          <section className="space-y-3 rounded-lg border border-border bg-background p-4">
            <h2 className="text-sm font-semibold text-foreground">{t("orderDetails.shippingAddress")}</h2>
            <div className="text-xs text-muted-foreground leading-relaxed">
              {order.shippingAddress}
            </div>
          </section>

          <section className="space-y-3 rounded-lg border border-border bg-background p-4">
            <h2 className="text-sm font-semibold text-foreground">{t("orderDetails.paymentDetails")}</h2>
            <div className="text-xs space-y-2">
              <div className="flex justify-between">
                <span className="text-muted-foreground">{t("orderDetails.method")}</span>
                <span className="font-semibold text-foreground">{order.paymentMethod}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">{t("orderDetails.statusLabel")}</span>
                <StatusPill tone={statusTone(order.paymentStatus)}>{t(`status.${order.paymentStatus}`, { defaultValue: order.paymentStatus })}</StatusPill>
              </div>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
}
