import { useState, useEffect } from "react";
import { useRouter } from "@/i18n/navigation";
import { useParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { ArrowLeft, Truck, Clock, MapPin, CreditCard, User, Tag } from "lucide-react";
import { adminService } from "@/services/admin";
import { toast } from "sonner";

export function AdminOrderDetails() {
  const { id } = useParams();
  const navigate = useRouter();
  const t = useTranslations("admin");

  const [order, setOrder] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [trackingNumber, setTrackingNumber] = useState("");
  const [noteInput, setNoteInput] = useState("");

  const fetchOrderFromApi = async () => {
    if (!id) return;
    try {
      setLoading(true);
      const res: any = await adminService.getOrderById(String(id));
      const remoteOrder = res?.data || res;
      if (remoteOrder) {
        const formatted = {
          id: remoteOrder.id,
          customer: remoteOrder.customerEmail ? remoteOrder.customerEmail.split("@")[0] : "Customer",
          email: remoteOrder.customerEmail || "customer@example.com",
          date: remoteOrder.createdAt ? new Date(remoteOrder.createdAt).toISOString().slice(0, 10) : new Date().toISOString().slice(0, 10),
          total: `EGP ${((remoteOrder.totalAmount || 0) / 100).toFixed(2)}`,
          discount: remoteOrder.discountAmount ? `EGP ${((remoteOrder.discountAmount || 0) / 100).toFixed(2)}` : null,
          discountAmount: remoteOrder.discountAmount || 0,
          couponCode: remoteOrder.couponCode || null,
          couponId: remoteOrder.couponId || null,
          status: remoteOrder.status ? remoteOrder.status.toLowerCase() : "pending",
          fulfillmentStatus: remoteOrder.status === "shipped" || remoteOrder.status === "delivered" ? "Fulfilled" : "Unfulfilled",
          paymentStatus: remoteOrder.status === "cancelled" ? "Canceled" : "Paid",
          shippingAddress: remoteOrder.shippingAddress ? `${remoteOrder.shippingAddress.street || ""}, ${remoteOrder.shippingAddress.city || ""}, ${remoteOrder.shippingAddress.state || ""} ${remoteOrder.shippingAddress.postalCode || ""}` : "No shipping address provided",
          paymentMethod: remoteOrder.paymentMethod || "—",
          trackingNumber: remoteOrder.trackingNumber || "",
          items: (Array.isArray(remoteOrder.items) ? remoteOrder.items : []).map((it: any) => {
            let itemName = "Order Item";
            if (typeof it.productSnapshot === "string") {
              try {
                const parsed = JSON.parse(it.productSnapshot);
                itemName = parsed.title || parsed.name || itemName;
              } catch {
                itemName = it.productSnapshot;
              }
            } else if (it.productSnapshot && typeof it.productSnapshot === "object") {
              itemName = it.productSnapshot.title || it.productSnapshot.name || itemName;
            }
            return {
              id: it.id,
              name: itemName,
              price: `EGP ${((it.priceAtPurchase || 0) / 100).toFixed(2)}`,
              quantity: it.quantity || 1
            };
          }),
          history: [
            { date: remoteOrder.createdAt ? new Date(remoteOrder.createdAt).toISOString().slice(0, 10) : "Today", note: "Order placed successfully." }
          ]
        };
        setOrder(formatted);
        setTrackingNumber(formatted.trackingNumber);
      }
    } catch {
      toast.error(t("orderDetails.toast.loadFailed"));
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

  if (!order) {
    return <div className="p-8 text-center text-muted-foreground">{t("orderDetails.notFound")}</div>;
  }

  const handleUpdateFulfillment = async (newStatus: string) => {
    try {
      await adminService.updateOrderStatus(order.id, newStatus.toLowerCase());
      toast.success(t("orderDetails.toast.statusUpdated", { status: newStatus }));
      fetchOrderFromApi();
    } catch (err: any) {
      toast.error(err?.message || err?.response?.data?.message || t("orderDetails.toast.statusFailed"));
      fetchOrderFromApi();
    }
  };

  const handleSaveTracking = async () => {
    if (!trackingNumber.trim()) {
      toast.error(t("orderDetails.toast.enterTracking"));
      return;
    }
    try {
      await adminService.updateOrderTracking(order.id, trackingNumber);
      toast.success(t("orderDetails.toast.trackingSaved"));
    } catch (err: any) {
      toast.error(err?.response?.data?.message || t("orderDetails.toast.trackingFailed"));
    }
  };

  const handleAddTimelineNote = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!noteInput.trim()) return;

    try {
      await adminService.updateOrderStatus(order.id, order.status.toLowerCase(), noteInput.trim());
      toast.success(t("orderDetails.toast.noteAdded"));
      setNoteInput("");
      fetchOrderFromApi();
    } catch (err: any) {
      toast.error(err?.response?.data?.message || t("orderDetails.toast.noteFailed"));
    }
  };

  return (
    <div className="space-y-6 w-full">
      <div className="flex items-center justify-between">
        <button
          onClick={() => navigate.push("/admin/orders")}
          className="flex items-center gap-2 text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors"
        >
          <ArrowLeft className="size-4" />
          <span>{t("orderDetails.back")}</span>
        </button>
        <span className="text-xs font-mono font-bold px-3 py-1 bg-secondary rounded-lg border border-border">
          {order.id}
        </span>
      </div>

      <div className="bg-background border border-border rounded-2xl p-6 shadow-xs flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold text-foreground">{t("orderDetails.title")}</h1>
            <span
              className={`px-3 py-0.5 rounded-full text-xs font-bold ${
                order.status === "Delivered"
                  ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400"
                  : order.status === "Processing"
                  ? "bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-400"
                  : order.status === "Canceled" || order.status === "Cancelled"
                  ? "bg-rose-100 text-rose-800 dark:bg-rose-900/40 dark:text-rose-400"
                  : "bg-purple-100 text-purple-800 dark:bg-purple-900/40 dark:text-purple-400"
              }`}
            >
              {order.status}
            </span>
          </div>
          <p className="text-xs text-muted-foreground mt-1">{t("orderDetails.placedOn", { date: order.date })}</p>
        </div>

        <div className="flex items-center gap-2">
          <select
            value={order.status.charAt(0).toUpperCase() + order.status.slice(1)}
            onChange={(e) => handleUpdateFulfillment(e.target.value.toLowerCase())}
            disabled={order.status === "cancelled" || order.status === "delivered"}
            className="px-3 py-2 bg-secondary border border-border rounded-xl text-xs font-semibold text-foreground disabled:opacity-50 disabled:cursor-not-allowed"
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
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column: Line Items & History */}
        <div className="lg:col-span-2 space-y-6">
          {/* Purchased Items */}
          <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
            <h3 className="font-bold text-base text-foreground border-b border-border pb-3">{t("orderDetails.lineItems", { count: order.items.length })}</h3>
            <div className="divide-y divide-border">
              {order.items.map((item) => (
                <div key={item.id} className="py-3 flex items-center justify-between text-xs">
                  <div>
                    <div className="font-semibold text-foreground">{item.name}</div>
                    <div className="text-muted-foreground">{t("orderDetails.qty", { count: item.quantity })}</div>
                  </div>
                  <div className="font-bold text-foreground">{item.price}</div>
                </div>
              ))}
            </div>
            {order.couponCode && (
              <div className="pt-2 border-t border-border flex justify-between items-center text-xs">
                <span className="text-emerald-600 dark:text-emerald-400 font-semibold flex items-center gap-1.5">
                  <Tag className="size-3.5" />
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
              <span className="text-accent-brand text-base">{order.total}</span>
            </div>
          </div>

          {/* Tracking Number Editor */}
          <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
            <h3 className="font-bold text-base text-foreground flex items-center gap-2">
              <Truck className="size-5 text-accent-brand" />
              <span>{t("orderDetails.shipmentTracking")}</span>
            </h3>
            <div className="flex gap-3">
              <input
                type="text"
                placeholder={t("orderDetails.trackingPlaceholder")}
                value={trackingNumber}
                onChange={(e) => setTrackingNumber(e.target.value)}
                className="flex-1 px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
              <button
                onClick={handleSaveTracking}
                className="px-4 py-2 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90"
              >
                {t("orderDetails.saveTracking")}
              </button>
            </div>
          </div>

          {/* History Timeline */}
          <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
            <h3 className="font-bold text-base text-foreground flex items-center gap-2">
              <Clock className="size-5 text-accent-brand" />
              <span>{t("orderDetails.timeline")}</span>
            </h3>
            <div className="space-y-3">
              {order.history.map((log, idx) => (
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
                className="flex-1 px-3 py-1.5 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
              <button type="submit" className="px-3 py-1.5 bg-secondary text-foreground text-xs font-semibold rounded-xl border border-border">
                {t("orderDetails.addNote")}
              </button>
            </form>
          </div>
        </div>

        {/* Right Column: Customer & Payment Details */}
        <div className="space-y-6">
          <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
            <h3 className="font-bold text-base text-foreground flex items-center gap-2">
              <User className="size-4 text-accent-brand" />
              <span>{t("orderDetails.customer")}</span>
            </h3>
            <div className="text-xs space-y-1">
              <div className="font-semibold text-foreground">{order.customer}</div>
              <div className="text-muted-foreground">{order.email}</div>
            </div>
          </div>

          <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
            <h3 className="font-bold text-base text-foreground flex items-center gap-2">
              <MapPin className="size-4 text-accent-brand" />
              <span>{t("orderDetails.shippingAddress")}</span>
            </h3>
            <div className="text-xs text-muted-foreground leading-relaxed">
              {order.shippingAddress}
            </div>
          </div>

          <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
            <h3 className="font-bold text-base text-foreground flex items-center gap-2">
              <CreditCard className="size-4 text-accent-brand" />
              <span>{t("orderDetails.paymentDetails")}</span>
            </h3>
            <div className="text-xs space-y-2">
              <div className="flex justify-between">
                <span className="text-muted-foreground">{t("orderDetails.method")}</span>
                <span className="font-semibold text-foreground">{order.paymentMethod}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">{t("orderDetails.statusLabel")}</span>
                <span className="font-bold text-emerald-600">{order.paymentStatus}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
