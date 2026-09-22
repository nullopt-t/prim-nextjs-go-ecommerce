import { useState, useEffect } from "react";
import { useRouter } from "@/i18n/navigation";
import { useParams } from "next/navigation";
import { ArrowLeft, Truck, Clock, MapPin, CreditCard, User } from "lucide-react";
import { INITIAL_ORDERS, getLocalAdminData, setLocalAdminData } from "../data/mockAdminData";
import { orderService } from "@/services/orders";
import { toast } from "sonner";

export function AdminOrderDetails() {
  const { id } = useParams();
  const navigate = useRouter();

  const [orders, setOrders] = useState(() => getLocalAdminData("orders", INITIAL_ORDERS));
  const order = orders.find((o) => o.id === id) || orders[0];

  const [trackingNumber, setTrackingNumber] = useState(() => order?.trackingNumber || "");
  const [noteInput, setNoteInput] = useState("");

  useEffect(() => {
    let isMounted = true;
    const fetchOrderFromApi = async () => {
      if (!id) return;
      try {
        const remoteOrder = await orderService.getOrderById(id);
        if (isMounted && remoteOrder) {
          setOrders((current) => {
            const existingIdx = current.findIndex((o) => o.id === remoteOrder.id);
            const formatted = {
              id: remoteOrder.id,
              customer: remoteOrder.customer || "Online Customer",
              email: remoteOrder.email || "customer@example.com",
              date: remoteOrder.date || new Date().toISOString().slice(0, 10),
              total: typeof remoteOrder.total === "string" ? remoteOrder.total : `$${(remoteOrder.total || 0).toFixed(2)}`,
              status: remoteOrder.status || "Processing",
              fulfillmentStatus: remoteOrder.fulfillmentStatus || (remoteOrder.status === "Shipped" || remoteOrder.status === "Delivered" ? "Fulfilled" : "Unfulfilled"),
              paymentStatus: remoteOrder.paymentStatus || "Paid",
              items: remoteOrder.items || remoteOrder.orderItems?.length || 1,
              shippingAddress: remoteOrder.shippingAddress || "123 Main St, New York, NY",
              billingAddress: remoteOrder.billingAddress || remoteOrder.shippingAddress || "123 Main St, New York, NY",
              paymentMethod: remoteOrder.paymentMethod || "Credit Card",
              trackingNumber: remoteOrder.trackingNumber || "TRK-PRM-9921",
              itemsList: remoteOrder.orderItems || remoteOrder.itemsList || [],
              history: remoteOrder.timeline?.map((t) => ({ date: t.date, note: t.title })) || [
                { date: remoteOrder.date || "Today", note: "Order placed successfully." }
              ]
            };

            const updated = [...current];
            if (existingIdx >= 0) {
              updated[existingIdx] = { ...updated[existingIdx], ...formatted };
            } else {
              updated.unshift(formatted);
            }
            setLocalAdminData("orders", updated);
            return updated;
          });
        }
      } catch {
        // Fallback to local admin orders
      }
    };

    fetchOrderFromApi();
    return () => { isMounted = false; };
  }, [id]);

  if (!order) {
    return <div className="p-8 text-center text-muted-foreground">Order not found.</div>;
  }

  const saveOrders = (updated) => {
    setOrders(updated);
    setLocalAdminData("orders", updated);
  };

  const handleUpdateFulfillment = async (newStatus) => {
    const updated = orders.map((o) =>
      o.id === order.id
        ? {
            ...o,
            status: newStatus,
            fulfillmentStatus: newStatus === "Shipped" || newStatus === "Delivered" ? "Fulfilled" : "Unfulfilled",
            history: [
              ...(o.history || []),
              {
                date: new Date().toISOString().replace("T", " ").substring(0, 16),
                note: `Order status changed to ${newStatus}.`
              }
            ]
          }
        : o
    );
    saveOrders(updated);

    try {
      await orderService.updateOrder(order.id, { status: newStatus });
    } catch {
      // Ignored in mock
    }

    toast.success(`Fulfillment status updated to ${newStatus}`);
  };

  const handleSaveTracking = async () => {
    const updated = orders.map((o) =>
      o.id === order.id
        ? {
            ...o,
            trackingNumber,
            history: [
              ...(o.history || []),
              {
                date: new Date().toISOString().replace("T", " ").substring(0, 16),
                note: `Tracking number set to ${trackingNumber}.`
              }
            ]
          }
        : o
    );
    saveOrders(updated);

    try {
      await orderService.updateOrder(order.id, { trackingNumber });
    } catch {
      // Ignored in mock
    }

    toast.success("Tracking information saved!");
  };

  const handleAddTimelineNote = async (e) => {
    e.preventDefault();
    if (!noteInput.trim()) return;

    const newNote = noteInput.trim();
    const updated = orders.map((o) =>
      o.id === order.id
        ? {
            ...o,
            history: [
              ...(o.history || []),
              {
                date: new Date().toISOString().replace("T", " ").substring(0, 16),
                note: newNote
              }
            ]
          }
        : o
    );
    saveOrders(updated);
    setNoteInput("");

    try {
      await orderService.updateOrder(order.id, { note: newNote });
    } catch {
      // Ignored in mock
    }

    toast.success("Note added to order history timeline");
  };

  return (
    <div className="space-y-6 max-w-5xl">
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
            value={order.status}
            onChange={(e) => handleUpdateFulfillment(e.target.value)}
            className="px-3 py-2 bg-secondary border border-border rounded-xl text-xs font-semibold text-foreground"
          >
            <option value="Processing">Processing</option>
            <option value="Shipped">Shipped</option>
            <option value="Delivered">Delivered</option>
            <option value="Cancelled">Cancelled</option>
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
