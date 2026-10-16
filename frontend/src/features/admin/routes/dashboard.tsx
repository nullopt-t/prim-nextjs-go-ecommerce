import { useState, useEffect } from "react";
import { useRouter } from "@/i18n/navigation";
import { useTranslations } from "next-intl";
import {
  ArrowUpRight,
  Eye
} from "lucide-react";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  BarChart,
  Bar
} from "recharts";
import { adminService } from "@/services/admin";

export function AdminDashboard() {
  const navigate = useRouter();
  const t = useTranslations("admin");
  const [timeframe, setTimeframe] = useState("monthly");
  const [stats, setStats] = useState<any>(null);
  const [recentOrders, setRecentOrders] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let isMounted = true;
    const loadDashboardData = async () => {
      try {
        const [dashRes, ordersRes] = await Promise.all([
          adminService.getDashboard(),
          adminService.getOrders({ pageSize: 5 })
        ]);

        if (!isMounted) return;

        if (dashRes?.data) {
          setStats(dashRes.data);
        }
        const ords = Array.isArray(ordersRes?.data) ? ordersRes.data : (Array.isArray(ordersRes) ? ordersRes : []);
        setRecentOrders(ords.map((o: any) => ({
          id: o.id,
          customer: o.customerEmail ? o.customerEmail.split("@")[0] : "Customer",
          status: o.status ? o.status.charAt(0).toUpperCase() + o.status.slice(1) : "Pending",
          total: `EGP ${((o.totalAmount || 0) / 100).toFixed(2)}`
        })));
      } catch {
        // Handled silently
      } finally {
        if (isMounted) setLoading(false);
      }
    };

    loadDashboardData();
    return () => { isMounted = false; };
  }, []);

  const totalRev = stats ? stats.totalRevenue / 100 : 0;
  const totalOrds = stats?.totalOrders || 0;
  const totalCusts = stats?.totalCustomers || 0;
  const totalProds = stats?.totalProducts || 0;

  const chartData = (stats?.recentSales && stats.recentSales.length > 0)
    ? stats.recentSales.map((s: any) => ({
        month: s.date.slice(5),
        sales: s.sales / 100,
        orders: s.orders
      }))
    : [];

  return (
    <div className="space-y-6">
      {/* Metrics Row */}

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">{t("dashboard.totalRevenue")}</span>
          </div>
          <div className="mt-4">
            <div className="text-2xl font-bold text-foreground">EGP {totalRev.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</div>
            <div className="mt-1 text-xs text-muted-foreground font-medium">
              {t("dashboard.totalRevenueSub")}
            </div>
          </div>
        </div>

        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">{t("dashboard.totalOrders")}</span>
          </div>
          <div className="mt-4">
            <div className="text-2xl font-bold text-foreground">{totalOrds.toLocaleString()}</div>
            <div className="mt-1 text-xs text-muted-foreground font-medium">
              {t("dashboard.pendingProcessing", { count: stats?.pendingOrders || 0 })}
            </div>
          </div>
        </div>

        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">{t("dashboard.catalogProducts")}</span>
          </div>
          <div className="mt-4">
            <div className="text-2xl font-bold text-foreground">{totalProds.toLocaleString()}</div>
            <div className="mt-1 text-xs text-muted-foreground font-medium">
              {t("dashboard.activeInCatalog")}
            </div>
          </div>
        </div>

        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">{t("dashboard.activeCustomers")}</span>
          </div>
          <div className="mt-4">
            <div className="text-2xl font-bold text-foreground">{totalCusts.toLocaleString()}</div>
            <div className="mt-1 text-xs text-muted-foreground font-medium">
              {t("dashboard.registeredAccounts")}
            </div>
          </div>
        </div>
      </div>

      {/* Analytics Chart */}
      <div className="bg-background border border-border rounded-2xl p-6 shadow-xs">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
          <div>
            <h2 className="text-lg font-bold text-foreground">{t("dashboard.revenueTrends")}</h2>
            <p className="text-xs text-muted-foreground">{t("dashboard.revenueTrendsSub")}</p>
          </div>
          <div className="flex items-center gap-2 bg-secondary/50 p-1 rounded-xl border border-border">
            <button
              onClick={() => setTimeframe("monthly")}
              className={`px-3 py-1 text-xs font-semibold rounded-lg transition-all ${
                timeframe === "monthly" ? "bg-background text-foreground shadow-xs" : "text-muted-foreground hover:text-foreground"
              }`}
            >
              {t("dashboard.revenueEgp")}
            </button>
            <button
              onClick={() => setTimeframe("orders")}
              className={`px-3 py-1 text-xs font-semibold rounded-lg transition-all ${
                timeframe === "orders" ? "bg-background text-foreground shadow-xs" : "text-muted-foreground hover:text-foreground"
              }`}
            >
              {t("dashboard.orderCount")}
            </button>
          </div>
        </div>

        <div className="h-72 w-full">
          <ResponsiveContainer width="100%" height="100%">
            {timeframe === "monthly" ? (
              <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorSales" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#6366f1" stopOpacity={0.4} />
                    <stop offset="95%" stopColor="#6366f1" stopOpacity={0.0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--color-border, #e5e7eb)" />
                <XAxis dataKey="month" tick={{ fontSize: 12, fill: "gray" }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fontSize: 12, fill: "gray" }} axisLine={false} tickLine={false} tickFormatter={(val) => `${val / 1000}k EGP`} />
                <Tooltip
                  formatter={(val: any) => [`EGP ${val.toLocaleString()}`, t("dashboard.grossRevenue")]}
                  contentStyle={{ backgroundColor: "var(--color-bg, #ffffff)", borderRadius: "12px", border: "1px solid #e5e7eb" }}
                />
                <Area type="monotone" dataKey="sales" stroke="#6366f1" strokeWidth={3} fillOpacity={1} fill="url(#colorSales)" />
              </AreaChart>
            ) : (
              <BarChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--color-border, #e5e7eb)" />
                <XAxis dataKey="month" tick={{ fontSize: 12, fill: "gray" }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fontSize: 12, fill: "gray" }} axisLine={false} tickLine={false} />
                <Tooltip
                  formatter={(val: any) => [val, t("dashboard.orders")]}
                  contentStyle={{ backgroundColor: "var(--color-bg, #ffffff)", borderRadius: "12px", border: "1px solid #e5e7eb" }}
                />
                <Bar dataKey="orders" fill="#3b82f6" radius={[6, 6, 0, 0]} />
              </BarChart>
            )}
          </ResponsiveContainer>
        </div>
      </div>

      {/* Grid Section: Recent Orders & Inventory Warnings */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Recent Orders */}
        <div className="lg:col-span-2 bg-background border border-border rounded-2xl p-6 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="font-bold text-base text-foreground">{t("dashboard.recentOrders")}</h3>
                <p className="text-xs text-muted-foreground">{t("dashboard.recentOrdersSub")}</p>
              </div>
              <button
                onClick={() => navigate.push("/admin/orders")}
                className="text-xs font-semibold text-accent-brand hover:underline flex items-center gap-1"
              >
                <span>{t("dashboard.viewAllOrders")}</span>
                <ArrowUpRight className="size-3.5" />
              </button>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
                    <th className="pb-3 pl-2">{t("dashboard.orderId")}</th>
                    <th className="pb-3">{t("dashboard.customer")}</th>
                    <th className="pb-3">{t("dashboard.status")}</th>
                    <th className="pb-3">{t("dashboard.total")}</th>
                    <th className="pb-3 pr-2 text-right">{t("dashboard.action")}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/60">
                  {recentOrders.map((order) => (
                    <tr key={order.id} className="hover:bg-secondary/20 transition-colors">
                      <td className="py-3.5 pl-2 font-mono font-bold text-foreground text-xs line-clamp-1">{order.id}</td>
                      <td className="py-3.5">
                        <div className="font-medium text-foreground">{order.customer}</div>
                      </td>
                      <td className="py-3.5">
                        <span
                          className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${
                            order.status === "Delivered"
                              ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-400"
                              : order.status === "Processing"
                              ? "bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-400"
                              : order.status === "Shipped"
                              ? "bg-purple-100 text-purple-800 dark:bg-purple-900/40 dark:text-purple-400"
                              : "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-400"
                          }`}
                        >
                          {order.status}
                        </span>
                      </td>
                      <td className="py-3.5 font-bold text-foreground">{order.total}</td>
                      <td className="py-3.5 pr-2 text-right">
                        <button
                          onClick={() => navigate.push(`/admin/orders/${order.id}`)}
                          className="p-1.5 text-muted-foreground hover:text-accent-brand hover:bg-secondary rounded-lg transition-colors"
                          title={t("dashboard.viewOrderDetails")}
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

        {/* Low Stock Alerts */}
        <div className="bg-background border border-border rounded-2xl p-6 shadow-xs flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="font-bold text-base text-foreground">{t("dashboard.stockAlerts")}</h3>
                <p className="text-xs text-muted-foreground">{t("dashboard.stockAlertsSub")}</p>
              </div>
              <button
                onClick={() => navigate.push("/admin/inventory")}
                className="text-xs font-semibold text-accent-brand hover:underline"
              >
                {t("dashboard.inventory")}
              </button>
            </div>

            <div className="space-y-3">
              <div className="text-center text-xs text-muted-foreground py-6">
                {t("dashboard.inventoryHealthy")}
              </div>
            </div>
          </div>

          <button
            onClick={() => navigate.push("/admin/inventory")}
            className="w-full mt-4 py-2.5 bg-secondary text-secondary-foreground text-xs font-semibold rounded-xl border border-border hover:bg-secondary/80 transition-colors"
          >
            {t("dashboard.manageStockAdjustments")}
          </button>
        </div>
      </div>
    </div>
  );
}
