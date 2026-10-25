import { useState, useEffect } from "react";

import { Search } from "lucide-react";
import {
  INITIAL_VENDOR_ORDERS,
  getLocalVendorData,
  setLocalVendorData,
  type VendorOrder
} from "../data/mockVendorData";
import { orderService } from "@/modules/orders/api/orderService";
import { notify } from "@/shared/lib/notify";
import { useTranslations } from "next-intl";

import { useVendorContext } from "./vendorContext";

interface MarketplaceOrder {
	id?: string;
	status?: string;
	date?: string;
	shippingAddress?: string | { street?: string };
	orderItems?: Array<{
		name?: string;
		price?: string | number;
		quantity?: number;
	}>;
}

export function VendorOrders() {
  const t = useTranslations("vendor");
  const { vendor } = useVendorContext();

  const [orders, setOrders] = useState<VendorOrder[]>(() =>
    getLocalVendorData("orders", INITIAL_VENDOR_ORDERS)
  );
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("All");

  useEffect(() => {
    async function loadOrders() {
      try {
        const res = await orderService.getOrders();
        const list: MarketplaceOrder[] = (res?.data ?? []) as unknown as MarketplaceOrder[];
        if (list.length > 0) {
          const commissionRate = (vendor?.commissionRate || 10) / 100;

          // Convert marketplace orders to vendor orders view
          const serverMappedOrders = list.map((o: MarketplaceOrder, idx: number) => {
            const firstItem = o.orderItems?.[0];
            const rawPrice = firstItem?.price ? parseFloat(String(firstItem.price).replace(/[^0-9.]/g, "")) : 99.99;
            const quantity = firstItem?.quantity || 1;
            const gross = rawPrice * quantity;
            const comm = Math.round(gross * commissionRate * 100) / 100;
            const net = Math.round((gross - comm) * 100) / 100;

            return {
              id: `VORD-${o.id || idx}`,
              vendorId: vendor?.id || "v-1",
              orderId: String(o.id || `ORD-${idx}`),
              customer: typeof o.shippingAddress === "string" ? o.shippingAddress.split(",")[0] : (o.shippingAddress?.street || "Customer"),
              email: "customer@market.com",
              date: o.date || "Recent",
              item: firstItem?.name || "Marketplace Product",
              quantity,
              unitPrice: rawPrice,
              vendorEarnings: net,
              commissionFee: comm,
              fulfillmentStatus: o.status || "Processing",
              payoutStatus: o.status === "Delivered" ? "Paid" : "Pending"
            };
          });

          // Merge server mapped orders with initial/local vendor orders, keeping unique orderIds
          setOrders((currentOrders) => {
            const existingIds = new Set(serverMappedOrders.map((m) => m.orderId));
            const localOnly = currentOrders.filter((o) => !existingIds.has(o.orderId));
            const merged = [...serverMappedOrders, ...localOnly];
            setLocalVendorData("orders", merged);
            return merged;
          });
        }
      } catch {
        // Fallback to local
      }
    }
    loadOrders();
  }, [vendor?.id, vendor?.commissionRate]);

  const saveOrders = (updated: VendorOrder[]) => {
    setOrders(updated);
    setLocalVendorData("orders", updated);
  };

  const statusLabel = (status: string) => {
    const map: Record<string, string> = {
      Processing: "status.processing",
      Shipped: "status.shipped",
      Delivered: "status.delivered",
      Cancelled: "status.cancelled",
      Pending: "status.pending"
    };
    return map[status] ? t(map[status]) : status;
  };

  const handleUpdateFulfillment = async (orderId: string, newFulfillment: string) => {
    const targetOrder = orders.find((o: VendorOrder) => o.id === orderId || o.orderId === orderId);
    const updated = orders.map((o: VendorOrder) =>
      o.id === orderId ? { ...o, fulfillmentStatus: newFulfillment } : o
    );
    saveOrders(updated);

    if (targetOrder?.orderId) {
      try {
        await orderService.updateOrder(targetOrder.orderId, { status: newFulfillment });
      } catch {
        // Ignored in mock
      }
    }

    notify.success(t("toasts.orderStatusUpdated", { status: statusLabel(newFulfillment) }));
  };

  const filtered = orders.filter((o: VendorOrder) => {
    const matchesSearch =
      o.orderId.toLowerCase().includes(search.toLowerCase()) ||
      o.customer.toLowerCase().includes(search.toLowerCase()) ||
      o.item.toLowerCase().includes(search.toLowerCase());
    const matchesStatus = statusFilter === "All" ? true : o.fulfillmentStatus === statusFilter;
    return matchesSearch && matchesStatus;
  });

  return (
    <div className="space-y-6 w-full">
      {/* Top Banner */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">{t("orders.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t("orders.subtitle", { store: vendor?.name || t("orders.storeFallback") })}
          </p>
        </div>
      </div>

      {/* Filter and Search */}
      <div className="bg-background border border-border rounded-2xl p-4 shadow-xs flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div className="relative max-w-sm flex-1">
          <Search className="absolute left-3.5 top-2.5 size-4 text-muted-foreground" />
          <input
            type="text"
            placeholder={t("orders.searchPlaceholder")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-secondary/30 border border-border rounded-xl text-xs text-foreground"
          />
        </div>

        <div className="flex items-center gap-2">
          <span className="text-xs text-muted-foreground font-semibold">{t("orders.statusLabel")}</span>
          {["All", "Processing", "Shipped", "Delivered"].map((st) => (
            <button
              key={st}
              onClick={() => setStatusFilter(st)}
              className={`px-3 py-1.5 text-xs font-semibold rounded-xl border transition-colors ${
                statusFilter === st
                  ? "bg-accent-brand text-white border-accent-brand"
                  : "bg-secondary text-foreground border-border hover:bg-secondary/80"
              }`}
            >
              {st === "All" ? t("status.all") : statusLabel(st)}
            </button>
          ))}
        </div>
      </div>

      {/* Orders Table */}
      <div className="bg-background border border-border rounded-2xl overflow-hidden shadow-xs">
        <div className="overflow-x-auto">
          <table className="w-full min-w-max text-left text-xs">
            <thead className="bg-secondary/40 border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
              <tr>
                <th className="py-3.5 pl-4">{t("orders.table.orderRef")}</th>
                <th className="py-3.5">{t("orders.table.date")}</th>
                <th className="py-3.5">{t("orders.table.customer")}</th>
                <th className="py-3.5">{t("orders.table.item")}</th>
                <th className="py-3.5">{t("orders.table.grossRevenue")}</th>
                <th className="py-3.5">{t("orders.table.netEarnings")}</th>
                <th className="py-3.5">{t("orders.table.fulfillmentStatus")}</th>
                <th className="py-3.5 pr-4 text-right">{t("orders.table.actions")}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {filtered.map((ord: VendorOrder) => (
                <tr key={ord.id} className="hover:bg-secondary/20 transition-colors">
                  <td className="py-3.5 pl-4 font-mono font-bold text-foreground">{ord.orderId}</td>
                  <td className="py-3.5 text-muted-foreground">{ord.date}</td>
                  <td className="py-3.5">
                    <div className="font-semibold text-foreground">{ord.customer}</div>
                    <div className="text-[10px] text-muted-foreground">{ord.email}</div>
                  </td>
                  <td className="py-3.5">
                    <div className="font-semibold text-foreground">{ord.item}</div>
                    <div className="text-[10px] text-muted-foreground">{t("orders.qtyPrice", { qty: ord.quantity, price: ord.unitPrice.toFixed(2) })}</div>
                  </td>
                  <td className="py-3.5 font-bold text-foreground">
                    EGP {(ord.unitPrice * ord.quantity).toFixed(2)}
                  </td>
                  <td className="py-3.5">
                    <div className="font-bold text-emerald-600 dark:text-emerald-400">
                      EGP {ord.vendorEarnings.toFixed(2)}
                    </div>
                    <div className="text-[10px] text-muted-foreground">{t("orders.commission", { amount: ord.commissionFee.toFixed(2) })}</div>
                  </td>
                  <td className="py-3.5">
                    <span
                      className={`px-2.5 py-0.5 rounded-full text-[10px] font-bold ${
                        ord.fulfillmentStatus === "Delivered"
                          ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400"
                          : ord.fulfillmentStatus === "Shipped"
                          ? "bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-400"
                          : "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-400"
                      }`}
                    >
                      {statusLabel(ord.fulfillmentStatus)}
                    </span>
                  </td>
                  <td className="py-3.5 pr-4 text-right">
                    <select
                      value={ord.fulfillmentStatus}
                      onChange={(e) => handleUpdateFulfillment(ord.id, e.target.value)}
                      className="px-2.5 py-1 bg-secondary border border-border rounded-lg text-xs font-semibold text-foreground focus:outline-none"
                    >
                      <option value="Processing">{t("status.processing")}</option>
                      <option value="Shipped">{t("status.shipped")}</option>
                      <option value="Delivered">{t("status.delivered")}</option>
                    </select>
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
