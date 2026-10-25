import { useState, useEffect } from "react";
import { Link } from "@/shared/i18n/navigation";
import {
  DollarSign,
  ShoppingBag,
  TrendingUp,
  Package,
  ArrowUpRight,
  Clock
} from "lucide-react";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer
} from "recharts";
import {
  INITIAL_VENDOR_PRODUCTS,
  INITIAL_VENDOR_ORDERS,
  getLocalVendorData
} from "../data/mockVendorData";
import { vendorService } from "@/modules/vendor/api/vendorService";
import { orderService } from "@/modules/orders/api/orderService";
import { catalogService } from "@/modules/catalog/api/catalogService";
import { useTranslations } from "next-intl";

const CHART_DATA = [
  { month: "Mar", sales: 1200, earnings: 1080 },
  { month: "Apr", sales: 1900, earnings: 1710 },
  { month: "May", sales: 2400, earnings: 2160 },
  { month: "Jun", sales: 3100, earnings: 2790 },
  { month: "Jul", sales: 4200, earnings: 3780 },
  { month: "Aug", sales: 5400, earnings: 4860 }
];

import { useVendorContext } from "./vendorContext";

export function VendorDashboard() {
  const t = useTranslations("vendor");
  const { vendor } = useVendorContext();

  const [products, setProducts] = useState(() => getLocalVendorData("products", INITIAL_VENDOR_PRODUCTS));
  const [orders, setOrders] = useState(() => getLocalVendorData("orders", INITIAL_VENDOR_ORDERS));
  const [statsData, setStatsData] = useState<Record<string, number> | null>(null);

  useEffect(() => {
    async function loadDashboardData() {
      try {
        const [statsRes, productsRes, ordersRes] = await Promise.allSettled([
          vendorService.getVendorStats(vendor?.id),
          catalogService.getProducts(),
          orderService.getOrders()
        ]);

        if (statsRes.status === "fulfilled" && statsRes.value) {
          const stats = statsRes.value?.data || statsRes.value;
          if (stats) setStatsData(stats);
        }

        if (productsRes.status === "fulfilled" && productsRes.value) {
          const rawProds = Array.isArray(productsRes.value) ? productsRes.value : Array.isArray(productsRes.value?.data) ? productsRes.value.data : [];
          const matched = rawProds
            .filter((p: any) => !p.vendorId || p.vendorId === (vendor?.id || "v-1"))
            .map((p: any) => ({
              id: p.id,
              name: p.product?.en || p.name || p.title || "Product",
              sku: p.sku || `SKU-${p.id}`,
              price: typeof p.price === "number" ? p.price : parseFloat(String(p.price).replace(/[^0-9.]/g, "")) || 0,
              stock: p.stock ?? p.stockCount ?? 20,
              salesCount: p.salesCount || Math.floor(Math.random() * 40 + 10),
              image: p.img || p.image || "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400&q=80"
            }));
          if (matched.length > 0) setProducts(matched as any);
        }

        if (ordersRes.status === "fulfilled" && ordersRes.value) {
          const rawOrders: any[] = Array.isArray(ordersRes.value) ? ordersRes.value : Array.isArray((ordersRes.value as any)?.data) ? (ordersRes.value as any).data : [];
          const mappedOrders = rawOrders.map((o: any) => {
            const first = o.orderItems?.[0];
            const price = first?.price ? parseFloat(String(first.price).replace(/[^0-9.]/g, "")) : 89.99;
            return {
              id: o.id,
              orderId: o.id,
              item: first?.name || "Audio Item",
              unitPrice: price,
              quantity: first?.quantity || 1,
              vendorEarnings: Math.round(price * 0.9 * 100) / 100,
              fulfillmentStatus: o.status || "Processing"
            };
          });
          if (mappedOrders.length > 0) setOrders(mappedOrders as any);
        }
      } catch {
        // Retain local state
      }
    }
    loadDashboardData();
  }, [vendor?.id]);

  const totalGrossSales =
    statsData?.grossSales ?? orders.reduce((acc: number, o: any) => acc + (o.unitPrice || 0) * (o.quantity || 1), 0);
  const totalNetEarnings =
    statsData?.netEarnings ?? orders.reduce((acc: number, o: any) => acc + (o.vendorEarnings || 0), 0);
  const totalUnitsSold =
    statsData?.totalUnitsSold ?? products.reduce((acc: number, p: any) => acc + (p.salesCount || 0), 0);
  const pendingOrdersCount =
    statsData?.pendingOrdersCount ?? orders.filter((o: any) => o.fulfillmentStatus === "Processing").length;

  const statusLabel = (status: string) => {
    const map: Record<string, string> = {
      Processing: "status.processing",
      Shipped: "status.shipped",
      Delivered: "status.delivered",
      Cancelled: "status.cancelled",
      Pending: "status.pending",
      Paid: "status.paid",
      Completed: "status.completed"
    };
    return map[status] ? t(map[status]) : status;
  };

  return (
    <div className="space-y-6 w-full">
      {/* Header Banner */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold text-foreground">{t("dashboard.title")}</h1>
            <span className="px-2.5 py-0.5 rounded-full text-xs font-bold bg-accent-brand/10 text-accent-brand border border-accent-brand/20">
              {vendor?.name || "SoundMaster Studio"}
            </span>
          </div>
          <p className="text-sm text-muted-foreground mt-1">
            {t("dashboard.subtitle")}
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Link
            href="/vendor/products"
            className="px-4 py-2 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all flex items-center gap-1.5"
          >
            <Package className="size-4" />
            <span>{t("actions.manageProducts")}</span>
          </Link>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex items-center justify-between">
          <div>
            <div className="text-xs font-semibold text-muted-foreground">{t("dashboard.grossSales")}</div>
            <div className="text-2xl font-bold text-foreground mt-1">EGP {totalGrossSales.toFixed(2)}</div>
            <div className="flex items-center gap-1 text-[11px] font-semibold text-emerald-600 dark:text-emerald-400 mt-1">
              <TrendingUp className="size-3" />
              <span>{t("dashboard.growth")}</span>
            </div>
          </div>
          <div className="size-11 rounded-2xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center">
            <DollarSign className="size-5" />
          </div>
        </div>

        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex items-center justify-between">
          <div>
            <div className="text-xs font-semibold text-muted-foreground">{t("dashboard.netEarnings")}</div>
            <div className="text-2xl font-bold text-foreground mt-1">EGP {totalNetEarnings.toFixed(2)}</div>
            <div className="text-[11px] text-muted-foreground mt-1">
              {t("dashboard.commissionNote", { rate: vendor?.commissionRate || 10 })}
            </div>
          </div>
          <div className="size-11 rounded-2xl bg-blue-500/10 text-blue-600 dark:text-blue-400 flex items-center justify-center">
            <TrendingUp className="size-5" />
          </div>
        </div>

        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex items-center justify-between">
          <div>
            <div className="text-xs font-semibold text-muted-foreground">{t("dashboard.totalUnitsSold")}</div>
            <div className="text-2xl font-bold text-foreground mt-1">{totalUnitsSold}</div>
            <div className="text-[11px] text-muted-foreground mt-1">{t("dashboard.acrossCatalog")}</div>
          </div>
          <div className="size-11 rounded-2xl bg-purple-500/10 text-purple-600 dark:text-purple-400 flex items-center justify-center">
            <ShoppingBag className="size-5" />
          </div>
        </div>

        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex items-center justify-between">
          <div>
            <div className="text-xs font-semibold text-muted-foreground">{t("dashboard.pendingFulfillment")}</div>
            <div className="text-2xl font-bold text-foreground mt-1">{t("dashboard.ordersCount", { count: pendingOrdersCount })}</div>
            <div className="text-[11px] text-amber-600 dark:text-amber-400 font-semibold mt-1">
              {t("dashboard.requiresShipping")}
            </div>
          </div>
          <div className="size-11 rounded-2xl bg-amber-500/10 text-amber-600 dark:text-amber-400 flex items-center justify-center">
            <Clock className="size-5" />
          </div>
        </div>
      </div>

      {/* Chart Section */}
      <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="font-bold text-base text-foreground">{t("dashboard.chartTitle")}</h3>
            <p className="text-xs text-muted-foreground mt-0.5">{t("dashboard.chartSubtitle")}</p>
          </div>
          <span className="text-xs font-semibold px-3 py-1 bg-secondary rounded-lg border border-border">
            {t("dashboard.lastSixMonths")}
          </span>
        </div>

        <div className="h-64 w-full pt-4">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={CHART_DATA}>
              <defs>
                <linearGradient id="colorSales" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="var(--color-accent-brand, #3b82f6)" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="var(--color-accent-brand, #3b82f6)" stopOpacity={0} />
                </linearGradient>
                <linearGradient id="colorEarnings" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#10b981" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="rgba(128,128,128,0.15)" />
              <XAxis dataKey="month" stroke="#888888" fontSize={11} tickLine={false} />
              <YAxis stroke="#888888" fontSize={11} tickLine={false} tickFormatter={(v) => `${v} EGP`} />
              <Tooltip
                contentStyle={{ backgroundColor: "rgba(15, 23, 42, 0.9)", borderRadius: "12px", border: "none", color: "#fff", fontSize: "12px" }}
              />
              <Area type="monotone" dataKey="sales" name={t("dashboard.chartGross")} stroke="var(--color-accent-brand, #3b82f6)" strokeWidth={2} fillOpacity={1} fill="url(#colorSales)" />
              <Area type="monotone" dataKey="earnings" name={t("dashboard.chartNet")} stroke="#10b981" strokeWidth={2} fillOpacity={1} fill="url(#colorEarnings)" />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Grid: Top Products & Recent Orders */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top Performing Vendor Products */}
        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
          <div className="flex justify-between items-center border-b border-border pb-3">
            <h3 className="font-bold text-base text-foreground">{t("dashboard.topProducts")}</h3>
            <Link href="/vendor/products" className="text-xs font-semibold text-accent-brand hover:underline flex items-center gap-1">
              <span>{t("actions.viewAll")}</span>
              <ArrowUpRight className="size-3.5" />
            </Link>
          </div>

          <div className="divide-y divide-border/60">
            {products.slice(0, 3).map((prod: any) => (
              <div key={prod.id} className="py-3 flex items-center gap-3">
                <img src={prod.image} alt={prod.name} className="size-12 rounded-xl object-cover border border-border" />
                <div className="flex-1 min-w-0">
                  <div className="font-semibold text-xs text-foreground truncate">{prod.name}</div>
                  <div className="text-[11px] text-muted-foreground mt-0.5">{t("dashboard.skuStock", { sku: prod.sku, stock: prod.stock })}</div>
                </div>
                <div className="text-right">
                  <div className="font-bold text-xs text-foreground">EGP {prod.price.toFixed(2)}</div>
                  <div className="text-[10px] text-emerald-600 dark:text-emerald-400 font-semibold">{t("dashboard.soldCount", { count: prod.salesCount })}</div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Recent Orders containing Vendor Products */}
        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs space-y-4">
          <div className="flex justify-between items-center border-b border-border pb-3">
            <h3 className="font-bold text-base text-foreground">{t("dashboard.recentOrders")}</h3>
            <Link href="/vendor/orders" className="text-xs font-semibold text-accent-brand hover:underline flex items-center gap-1">
              <span>{t("actions.viewOrders")}</span>
              <ArrowUpRight className="size-3.5" />
            </Link>
          </div>

          <div className="divide-y divide-border/60">
            {orders.slice(0, 3).map((ord: any) => (
              <div key={ord.id} className="py-3 flex items-center justify-between text-xs">
                <div>
                  <div className="font-mono font-bold text-foreground">{ord.orderId}</div>
                  <div className="text-[11px] text-muted-foreground truncate max-w-[200px]">{ord.item}</div>
                </div>
                <div className="text-right">
                  <div className="font-bold text-foreground">EGP {ord.unitPrice.toFixed(2)}</div>
                  <span
                    className={`inline-block px-2 py-0.5 rounded-full text-[10px] font-bold ${
                      ord.fulfillmentStatus === "Delivered"
                        ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400"
                        : "bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-400"
                    }`}
                  >
                    {statusLabel(ord.fulfillmentStatus)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
