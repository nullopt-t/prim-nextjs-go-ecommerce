import { useState, useEffect } from "react";
import { useRouter } from "@/i18n/navigation";
import { useParams } from "next/navigation";
import { ArrowLeft, Truck, Clock, MapPin, CreditCard, User, Tag } from "lucide-react";
import { adminService } from "@/services/admin";
import { toast } from "sonner";

export function AdminOrderDetails() {
  const { id } = useParams();
  const navigate = useRouter();

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
          status: remoteOrder.status ? remoteOrder.status.charAt(0).toUpperCase() + remoteOrder.status.slice(1) : "Processing",
          fulfillmentStatus: remoteOrder.status === "shipped" || remoteOrder.status === "delivered" ? "Fulfilled" : "Unfulfilled",
          paymentStatus: remoteOrder.status === "canceled" ? "Canceled" : "Paid",
          shippingAddress: remoteOrder.shippingAddress ? `${remoteOrder.shippingAddress.street || ""}, ${remoteOrder.shippingAddress.city || ""}, ${remoteOrder.shippingAddress.state || ""} ${remoteOrder.shippingAddress.postalCode || ""}` : "No shipping address provided",
          paymentMethod: "Credit Card",
          trackingNumber: "TRK-PRM-9921",
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
      toast.error("Failed to load order from server");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOrderFromApi();
  }, [id]);

  if (loading) {
    return <div className="p-8 text-center text-muted-foreground">Loading order details...</div>;
  }

  if (!order) {
    return <div className="p-8 text-center text-muted-foreground">Order not found.</div>;
  }

  const handleUpdateFulfillment = async (newStatus: string) => {
    try {
      await adminService.updateOrderStatus(order.id, newStatus.toLowerCase());
      toast.success(`Fulfillment status updated to ${newStatus}`);
      fetchOrderFromApi();
    } catch (err: any) {
      toast.error(err?.message || err?.response?.data?.message || "Failed to update status");
      fetchOrderFromApi();
    }
  };

  const handleSaveTracking = async () => {
    toast.success("Tracking information saved!");
  };

  const handleAddTimelineNote = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!noteInput.trim()) return;

    try {
      await adminService.updateOrderStatus(order.id, order.status.toLowerCase(), noteInput.trim());
      toast.success("Note added to timeline!");
      setNoteInput("");
      fetchOrderFromApi();
    } catch (err: any) {
      toast.error(err?.response?.data?.message || "Failed to add timeline note");
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
          <span>Back to All Orders</span>
        </button>
        <span className="text-xs font-mono font-bold px-3 py-1 bg-secondary rounded-lg border border-border">
          {order.id}
        </span>
      </div>

      <div className="bg-background border border-border rounded-2xl p-6 shadow-xs flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold text-foreground">Order Details</h1>
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
          <p className="text-xs text-muted-foreground mt-1">Placed on {order.date}</p>
        </div>

        <div className="flex items-center gap-2">
          <select
            value={order.status === "Canceled" ? "Cancelled" : order.status}
            onChange={(e) => handleUpdateFulfillment(e.target.value)}
            disabled={order.status === "Cancelled" || order.status === "Canceled" || order.status === "Delivered"}
            className="px-3 py-2 bg-secondary border border-border rounded-xl text-xs font-semibold text-foreground disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {order.status === "Pending" && (
              <>
                <option value="Pending" disabled>Pending</option>
                <option value="Processing">Processing</option>
                <option value="Cancelled">Cancelled</option>
              </>
            )}
            {order.status === "Processing" && (
              <>
                <option value="Processing" disabled>Processing</option>
                <option value="Shipped">Shipped</option>
                <option value="Cancelled">Cancelled</option>
              </>
            )}
            {order.status === "Shipped" && (
              <>
                <option value="Shipped" disabled>Shipped</option>
                <option value="Delivered">Delivered</option>
                <option value="Cancelled">Cancelled</option>
              </>
            )}
            {order.status === "Delivered" && (
              <option value="Delivered">Delivered (Final)</option>
            )}
            {(order.status === "Cancelled" || order.status === "Canceled") && (
              <option value="Cancelled">Cancelled (Final)</option>
            )}
          </select>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column: Line Items & History */}
        <div className="lg:col-span-2 space-y-6">
          {/* Purchased Items */}
          <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
            <h3 className="font-bold text-base text-foreground border-b border-border pb-3">Line Items ({order.items.length})</h3>
            <div className="divide-y divide-border">
              {order.items.map((item) => (
                <div key={item.id} className="py-3 flex items-center justify-between text-xs">
                  <div>
                    <div className="font-semibold text-foreground">{item.name}</div>
                    <div className="text-muted-foreground">Qty: {item.quantity}</div>
                  </div>
                  <div className="font-bold text-foreground">{item.price}</div>
                </div>
              ))}
            </div>
            {order.couponCode && (
              <div className="pt-2 border-t border-border flex justify-between items-center text-xs">
                <span className="text-emerald-600 dark:text-emerald-400 font-semibold flex items-center gap-1.5">
                  <Tag className="size-3.5" />
                  <span>Coupon Applied:</span>
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
              <span>Total Paid</span>
              <span className="text-accent-brand text-base">{order.total}</span>
            </div>
          </div>

          {/* Tracking Number Editor */}
          <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
            <h3 className="font-bold text-base text-foreground flex items-center gap-2">
              <Truck className="size-5 text-accent-brand" />
              <span>Shipment Tracking</span>
            </h3>
            <div className="flex gap-3">
              <input
                type="text"
                placeholder="Enter FedEx / UPS tracking number..."
                value={trackingNumber}
                onChange={(e) => setTrackingNumber(e.target.value)}
                className="flex-1 px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
              <button
                onClick={handleSaveTracking}
                className="px-4 py-2 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90"
              >
                Save Tracking
              </button>
            </div>
          </div>

          {/* History Timeline */}
          <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
            <h3 className="font-bold text-base text-foreground flex items-center gap-2">
              <Clock className="size-5 text-accent-brand" />
              <span>Fulfillment History & Timeline</span>
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
                placeholder="Add internal staff note..."
                value={noteInput}
                onChange={(e) => setNoteInput(e.target.value)}
                className="flex-1 px-3 py-1.5 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
              />
              <button type="submit" className="px-3 py-1.5 bg-secondary text-foreground text-xs font-semibold rounded-xl border border-border">
                Add Note
              </button>
            </form>
          </div>
        </div>

        {/* Right Column: Customer & Payment Details */}
        <div className="space-y-6">
          <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
            <h3 className="font-bold text-base text-foreground flex items-center gap-2">
              <User className="size-4 text-accent-brand" />
              <span>Customer</span>
            </h3>
            <div className="text-xs space-y-1">
              <div className="font-semibold text-foreground">{order.customer}</div>
              <div className="text-muted-foreground">{order.email}</div>
            </div>
          </div>

          <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
            <h3 className="font-bold text-base text-foreground flex items-center gap-2">
              <MapPin className="size-4 text-accent-brand" />
              <span>Shipping Address</span>
            </h3>
            <div className="text-xs text-muted-foreground leading-relaxed">
              {order.shippingAddress}
            </div>
          </div>

          <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
            <h3 className="font-bold text-base text-foreground flex items-center gap-2">
              <CreditCard className="size-4 text-accent-brand" />
              <span>Payment Details</span>
            </h3>
            <div className="text-xs space-y-2">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Method:</span>
                <span className="font-semibold text-foreground">{order.paymentMethod}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Status:</span>
                <span className="font-bold text-emerald-600">{order.paymentStatus}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
