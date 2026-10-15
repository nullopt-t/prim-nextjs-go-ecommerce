import { useState, useEffect } from "react";
import { useRouter } from "@/i18n/navigation";
import { Search, Filter, Eye } from "lucide-react";
import { adminService } from "@/services/admin";
import { toast } from "sonner";

export function AdminOrders() {
  const navigate = useRouter();
  const [orders, setOrders] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("All");

  const loadOrders = async () => {
    try {
      setLoading(true);
      const res = await adminService.getOrders({
        pageSize: 50,
        status: statusFilter !== "All" ? statusFilter.toLowerCase() : undefined
      });
      const list = Array.isArray(res?.data) ? res.data : (Array.isArray(res) ? res : []);
      const formatted = list.map((so: any) => ({
        id: so.id,
        customer: so.customerEmail ? so.customerEmail.split("@")[0] : "Customer",
        email: so.customerEmail || "customer@example.com",
        date: so.createdAt ? new Date(so.createdAt).toISOString().slice(0, 10) : new Date().toISOString().slice(0, 10),
        total: `$${((so.totalAmount || 0) / 100).toFixed(2)}`,
        status: so.status ? so.status.charAt(0).toUpperCase() + so.status.slice(1) : "Processing",
        fulfillmentStatus: so.status === "shipped" || so.status === "delivered" ? "Fulfilled" : "Unfulfilled",
        paymentStatus: so.status === "canceled" ? "Canceled" : "Paid",
        items: so.items?.length || 1,
        shippingAddress: so.shippingAddress ? `${so.shippingAddress.street || ""}, ${so.shippingAddress.city || ""}` : "Online Order",
        paymentMethod: "Credit Card",
        trackingNumber: "TRK-PRM-9921",
        itemsList: so.items || [],
        history: [
          { date: so.createdAt ? new Date(so.createdAt).toISOString().slice(0, 10) : "Today", note: "Order placed successfully." }
        ]
      }));
      setOrders(formatted);
    } catch {
      toast.error("Failed to load orders from server");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadOrders();
  }, [statusFilter]);

  const handleUpdateStatus = async (id: string, newStatus: string) => {
    try {
      await adminService.updateOrderStatus(id, newStatus.toLowerCase());
      toast.success(`Order status updated to ${newStatus}`);
      loadOrders();
    } catch (err: any) {
      toast.error(err?.response?.data?.message || "Failed to update order status");
    }
  };

  const filtered = orders.filter((o) => {
    const matchesSearch =
      o.id.toLowerCase().includes(search.toLowerCase()) ||
      o.customer.toLowerCase().includes(search.toLowerCase()) ||
      o.email.toLowerCase().includes(search.toLowerCase());
    const matchesStatus = statusFilter === "All" || o.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Customer Orders</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Track customer purchases, manage fulfillment statuses, and process order updates.
          </p>
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs flex flex-col md:flex-row gap-4 justify-between items-center">
        <div className="relative w-full md:w-80">
          <Search className="absolute left-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search by order ID or customer name..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
          />
        </div>

        <div className="flex items-center gap-3 w-full md:w-auto">
          <Filter className="size-4 text-muted-foreground" />
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
          >
            <option value="All">All Statuses</option>
            <option value="Processing">Processing</option>
            <option value="Shipped">Shipped</option>
            <option value="Delivered">Delivered</option>
            <option value="Cancelled">Cancelled</option>
          </select>
        </div>
      </div>

      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 pl-4">Order ID</th>
                <th className="py-3.5">Date & Time</th>
                <th className="py-3.5">Customer</th>
                <th className="py-3.5">Total</th>
                <th className="py-3.5">Payment</th>
                <th className="py-3.5">Fulfillment Status</th>
                <th className="py-3.5 pr-4 text-right">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {filtered.map((order) => (
                <tr key={order.id} className="hover:bg-secondary/20 transition-colors">
                  <td className="py-3.5 pl-4 font-mono font-bold text-foreground">{order.id}</td>
                  <td className="py-3.5 text-muted-foreground">{order.date}</td>
                  <td className="py-3.5">
                    <div className="font-semibold text-foreground">{order.customer}</div>
                    <div className="text-[10px] text-muted-foreground">{order.email}</div>
                  </td>
                  <td className="py-3.5 font-bold text-foreground">{order.total}</td>
                  <td className="py-3.5">
                    <span className="px-2 py-0.5 bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400 font-bold rounded-full text-[10px]">
                      {order.paymentStatus}
                    </span>
                  </td>
                  <td className="py-3.5">
                    <select
                      value={order.status}
                      onChange={(e) => handleUpdateStatus(order.id, e.target.value)}
                      className="px-2.5 py-1 bg-secondary/50 border border-border rounded-lg text-xs font-semibold text-foreground focus:outline-hidden"
                    >
                      <option value="Processing">Processing</option>
                      <option value="Shipped">Shipped</option>
                      <option value="Delivered">Delivered</option>
                      <option value="Cancelled">Cancelled</option>
                    </select>
                  </td>
                  <td className="py-3.5 pr-4 text-right">
                    <button
                      onClick={() => navigate.push(`/admin/orders/${order.id}`)}
                      className="p-1.5 text-muted-foreground hover:text-accent-brand hover:bg-secondary rounded-lg transition-colors"
                      title="Manage Order"
                    >
                      <Eye className="size-4" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
