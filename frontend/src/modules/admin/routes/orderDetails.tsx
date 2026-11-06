import { useState, useEffect } from "react";
import { useRouter } from "@/shared/i18n/navigation";
import { useParams } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { ArrowLeft } from "lucide-react";
import { PageHeader } from "@/shared/ui/pageHeader";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";
import { usePublicSettings } from "@/modules/settings/model/publicSettingsContext";
import { formatMoney } from "@/shared/lib/money";
import { adminService } from "@/modules/admin/api/adminService";
import type { Order, OrderItem } from "@/modules/orders/types";
import { notify, apiErrorMessage } from "@/shared/lib/notify";
import { formatDateTime } from "@/shared/lib/datetime";
import { ErrorState } from "@/shared/ui/errorState";

interface OrderViewItem {
  id: string;
  name: string;
  price: string;
  quantity: number;
  vendorName?: string;
}

interface OrderViewHistory {
  date: string;
  note: string;
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
  const [loadError, setLoadError] = useState(false);

  const fetchOrderFromApi = async () => {
    if (!id) return;
    try {
      setLoading(true);
      setLoadError(false);
      const res = await adminService.getOrderById(String(id));
      const remoteOrder: Order = res?.data || res;
      if (remoteOrder) {
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
            let vendorName: string | undefined;
            const snapshot: unknown = it.productSnapshot;
            if (typeof snapshot === "string") {
              try {
                const parsed = JSON.parse(snapshot);
                itemName = parsed.title || parsed.name || itemName;
                vendorName = parsed.vendorName;
              } catch {
                itemName = snapshot;
              }
            } else if (snapshot && typeof snapshot === "object") {
              const snap = snapshot as { title?: string; name?: string; vendorName?: string };
              itemName = snap.title || snap.name || itemName;
              vendorName = snap.vendorName;
            }
            return {
              id: it.id,
              name: itemName,
              vendorName,
              price: formatMoney((it.priceAtPurchase || 0) / 100, currency),
              quantity: it.quantity || 1
            };
          }),
          history: [
            { date: formatDateTime(remoteOrder.createdAt, locale), note: t("orderDetails.orderPlaced") }
          ]
        };
        setOrder(formatted);
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
        <select
          value={order.status.charAt(0).toUpperCase() + order.status.slice(1)}
          onChange={(e) => handleUpdateFulfillment(e.target.value.toLowerCase())}
          disabled={order.status === "cancelled" || order.status === "delivered"}
          aria-label={t("orderDetails.statusLabel")}
          className="rounded-md border border-border bg-background px-3 py-2 text-sm font-medium text-foreground disabled:cursor-not-allowed disabled:opacity-50"
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
                      {item.vendorName && (
                        <span className="px-1.5 py-0.5 rounded text-[10px] font-medium bg-secondary text-muted-foreground border border-border">
                          {item.vendorName}
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

          {/* History Timeline */}
          <section className="space-y-4 rounded-lg border border-border bg-background p-4">
            <h2 className="text-sm font-semibold text-foreground">{t("orderDetails.timeline")}</h2>
            <div className="space-y-3">
              {order.history.map((log: OrderViewHistory, idx: number) => (
                <div key={idx} className="flex items-start gap-3 text-xs">
                  <div className="size-2 rounded-full bg-accent-brand mt-1.5 shrink-0"></div>
                  <div>
                    <div className="font-medium text-foreground">{log.note}</div>
                    <div className="text-[10px] text-muted-foreground">{log.date}</div>
                  </div>
                </div>
              ))}
            </div>

            <form onSubmit={handleAddTimelineNote} className="pt-3 border-t border-border flex gap-2">
              <input
                type="text"
                placeholder={t("orderDetails.notePlaceholder")}
                value={noteInput}
                onChange={(e) => setNoteInput(e.target.value)}
                className="flex-1 rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground"
              />
              <button type="submit" className="rounded-md border border-border px-3 py-2 text-sm font-medium text-foreground hover:bg-secondary">
                {t("orderDetails.addNote")}
              </button>
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
