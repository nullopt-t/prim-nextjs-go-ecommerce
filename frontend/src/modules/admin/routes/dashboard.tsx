import { useState, useEffect, useCallback } from "react";
import { useRouter } from "@/shared/i18n/navigation";
import { useLocale, useTranslations } from "next-intl";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
} from "recharts";
import { adminService } from "@/modules/admin/api/adminService";
import { PageHeader } from "@/shared/ui/pageHeader";
import { ResponsiveTable } from "@/shared/ui/responsiveTable";
import { StatusPill, statusTone } from "@/shared/ui/statusPill";

interface LowStockItem {
  variantId: string;
  productId: string;
  name: string;
  sku: string;
  stock: number;
  available: number;
}

interface RecentOrder {
  id: string;
  customer: string;
  status: string;
  total: string;
  createdAt: string;
}

export function AdminDashboard() {
  const navigate = useRouter();
  const locale = useLocale();
  const t = useTranslations("admin");
  const [timeframe, setTimeframe] = useState<"revenue" | "orders">("revenue");
  const [stats, setStats] = useState<any>(null);
  const [recentOrders, setRecentOrders] = useState<RecentOrder[]>([]);
  const [lowStockItems, setLowStockItems] = useState<LowStockItem[]>([]);
  const [loading, setLoading] = useState(true);

  const loadLowStockData = useCallback(async () => {
    try {
      const invRes = await adminService.getInventory({ pageSize: 100 });
      const items: any[] = invRes?.data ?? [];
      const lowItems: LowStockItem[] = items
        .filter((it) => (it.onHandQuantity !== undefined ? it.onHandQuantity : 0) <= 10)
        .map((it) => ({
          variantId: it.variantId,
          productId: it.productId,
          name: `${it.productName}${it.variantTitle && it.variantTitle !== "Default" && it.variantTitle !== "Standard" ? ` (${it.variantTitle})` : ""}`,
          sku: it.sku,
          stock: it.onHandQuantity ?? 0,
          available: it.availableQuantity ?? 0,
        }))
        .slice(0, 6);
      setLowStockItems(lowItems);
    } catch {
      // The stock panel is supplementary; leave it empty when inventory fails.
    }
  }, []);

  const loadDashboardData = useCallback(async () => {
    try {
      const [dashRes, ordersRes] = await Promise.all([
        adminService.getDashboard(),
        adminService.getOrders({ pageSize: 5 }),
        loadLowStockData(),
      ]);

      if (dashRes?.data) {
        setStats(dashRes.data);
      }
      const ords = Array.isArray(ordersRes?.data)
        ? ordersRes.data
        : Array.isArray(ordersRes)
          ? ordersRes
          : [];
      setRecentOrders(
        ords.map((o: any) => ({
          id: o.id,
          customer: o.customerEmail ? o.customerEmail.split("@")[0] : "—",
          status: o.status ? String(o.status).toLowerCase() : "pending",
          total: `EGP ${((o.totalAmount || 0) / 100).toFixed(2)}`,
          createdAt: o.createdAt,
        })),
      );
    } catch {
      // Handled by the empty state below.
    } finally {
      setLoading(false);
    }
  }, [loadLowStockData]);

  useEffect(() => {
    loadDashboardData();
  }, [loadDashboardData]);

  useEffect(() => {
    const handleStockUpdate = (e: Event) => {
      const customEvent = e as CustomEvent;
      const detail = customEvent.detail;
      if (detail && detail.variantId) {
        setLowStockItems((prev) => {
          const currentStock = detail.currentStock;
          const exists = prev.some((item) => item.variantId === detail.variantId);
          if (currentStock <= 10) {
            const displayName = `${detail.productName}${detail.variantTitle && detail.variantTitle !== "Default" && detail.variantTitle !== "Standard" ? ` (${detail.variantTitle})` : ""}`;
            if (exists) {
              return prev.map((item) =>
                item.variantId === detail.variantId
                  ? { ...item, stock: currentStock, available: Math.max(0, currentStock) }
                  : item,
              );
            }
            return [
              {
                variantId: detail.variantId,
                productId: detail.productId,
                name: displayName,
                sku: detail.sku,
                stock: currentStock,
                available: Math.max(0, currentStock),
              },
              ...prev,
            ].slice(0, 6);
          }
          if (exists) {
            return prev.filter((item) => item.variantId !== detail.variantId);
          }
          return prev;
        });
      }
      loadLowStockData();
    };

    window.addEventListener("admin-stock-updated", handleStockUpdate);
    return () => {
      window.removeEventListener("admin-stock-updated", handleStockUpdate);
    };
  }, [loadLowStockData]);

  const totalRevenue = stats ? stats.totalRevenue / 100 : 0;
  const totalOrders = stats?.totalOrders || 0;
  const totalCustomers = stats?.totalCustomers || 0;
  const totalProducts = stats?.totalProducts || 0;

  const chartData =
    stats?.recentSales && stats.recentSales.length > 0
      ? stats.recentSales.map((s: any) => ({
          month: s.date.slice(5),
          sales: s.sales / 100,
          orders: s.orders,
        }))
      : [];

  const metrics = [
    {
      label: t("dashboard.totalRevenue"),
      value: `EGP ${totalRevenue.toLocaleString(locale, {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      })}`,
      detail: t("dashboard.totalRevenueSub"),
    },
    {
      label: t("dashboard.totalOrders"),
      value: totalOrders.toLocaleString(locale),
      detail: t("dashboard.pendingProcessing", { count: stats?.pendingOrders || 0 }),
    },
    {
      label: t("dashboard.catalogProducts"),
      value: totalProducts.toLocaleString(locale),
      detail: t("dashboard.activeInCatalog"),
    },
    {
      label: t("dashboard.activeCustomers"),
      value: totalCustomers.toLocaleString(locale),
      detail: t("dashboard.registeredAccounts"),
    },
  ];

  return (
    <div className="space-y-5">
      <PageHeader title={t("dashboard.title")} />

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {metrics.map((metric) => (
          <div
            key={metric.label}
            className="rounded-lg border border-border bg-background p-4"
          >
            <div className="text-xs font-medium text-muted-foreground">
              {metric.label}
            </div>
            <div className="mt-2 text-2xl font-semibold tracking-tight text-foreground tabular-nums">
              {metric.value}
            </div>
            <div className="mt-1 text-xs text-muted-foreground">{metric.detail}</div>
          </div>
        ))}
      </div>

      <section className="rounded-lg border border-border bg-background">
        <div className="flex flex-col gap-3 border-b border-border p-4 sm:flex-row sm:items-center sm:justify-between">
          <h2 className="text-sm font-semibold text-foreground">
            {t("dashboard.revenueTrends")}
          </h2>
          <div className="inline-flex rounded-md border border-border p-0.5">
            <button
              onClick={() => setTimeframe("revenue")}
              className={`rounded px-2.5 py-1 text-xs font-medium transition-colors ${
                timeframe === "revenue"
                  ? "bg-secondary text-foreground"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              {t("dashboard.revenueEgp")}
            </button>
            <button
              onClick={() => setTimeframe("orders")}
              className={`rounded px-2.5 py-1 text-xs font-medium transition-colors ${
                timeframe === "orders"
                  ? "bg-secondary text-foreground"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              {t("dashboard.orderCount")}
            </button>
          </div>
        </div>

        <div className="h-72 w-full p-4">
          {chartData.length === 0 ? (
            <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
              {t("dashboard.inventoryHealthy")}
            </div>
          ) : (
            <ResponsiveContainer width="100%" height="100%">
              {timeframe === "revenue" ? (
                <AreaChart data={chartData} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
                  <defs>
                    <linearGradient id="colorSales" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="var(--accent-brand)" stopOpacity={0.18} />
                      <stop offset="95%" stopColor="var(--accent-brand)" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--color-border)" />
                  <XAxis dataKey="month" tick={{ fontSize: 11, fill: "var(--muted-text)" }} axisLine={false} tickLine={false} />
                  <YAxis tick={{ fontSize: 11, fill: "var(--muted-text)" }} axisLine={false} tickLine={false} tickFormatter={(val) => `${val / 1000}k`} />
                  <Tooltip
                    formatter={(val: any) => [`EGP ${Number(val).toLocaleString(locale)}`, t("dashboard.grossRevenue")]}
                    contentStyle={{
                      backgroundColor: "var(--card-bg)",
                      borderRadius: "6px",
                      border: "1px solid var(--color-border)",
                      fontSize: "12px",
                    }}
                  />
                  <Area type="monotone" dataKey="sales" stroke="var(--accent-brand)" strokeWidth={2} fillOpacity={1} fill="url(#colorSales)" />
                </AreaChart>
              ) : (
                <BarChart data={chartData} margin={{ top: 8, right: 8, left: 0, bottom: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--color-border)" />
                  <XAxis dataKey="month" tick={{ fontSize: 11, fill: "var(--muted-text)" }} axisLine={false} tickLine={false} />
                  <YAxis tick={{ fontSize: 11, fill: "var(--muted-text)" }} axisLine={false} tickLine={false} />
                  <Tooltip
                    formatter={(val: any) => [val, t("dashboard.orders")]}
                    contentStyle={{
                      backgroundColor: "var(--card-bg)",
                      borderRadius: "6px",
                      border: "1px solid var(--color-border)",
                      fontSize: "12px",
                    }}
                  />
                  <Bar dataKey="orders" fill="var(--accent-brand)" radius={[3, 3, 0, 0]} />
                </BarChart>
              )}
            </ResponsiveContainer>
          )}
        </div>
      </section>

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-3">
        <section className="rounded-lg border border-border bg-background xl:col-span-2">
          <div className="flex items-center justify-between border-b border-border p-4">
            <h2 className="text-sm font-semibold text-foreground">
              {t("dashboard.recentOrders")}
            </h2>
            <button
              onClick={() => navigate.push("/admin/orders")}
              className="text-xs font-medium text-muted-foreground hover:text-foreground"
            >
              {t("dashboard.viewAllOrders")}
            </button>
          </div>

          <div className="overflow-x-auto">
            <ResponsiveTable className="w-full text-start text-sm md:min-w-max">
              <thead className="text-xs text-muted-foreground">
                <tr className="border-b border-border">
                  <th className="px-4 py-2.5 text-start font-medium">{t("dashboard.orderId")}</th>
                  <th className="py-2.5 text-start font-medium">{t("dashboard.customer")}</th>
                  <th className="py-2.5 text-start font-medium">{t("dashboard.status")}</th>
                  <th className="py-2.5 text-start font-medium">{t("dashboard.total")}</th>
                  <th className="px-4 py-2.5 text-end font-medium">{t("dashboard.action")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {loading ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-10 text-center text-sm text-muted-foreground">
                      {t("dashboard.recentOrdersSub")}
                    </td>
                  </tr>
                ) : recentOrders.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-10 text-center text-sm text-muted-foreground">
                      {t("dashboard.inventoryHealthy")}
                    </td>
                  </tr>
                ) : (
                  recentOrders.map((order) => (
                    <tr key={order.id} className="transition-colors hover:bg-secondary/30">
                      <td className="max-w-40 truncate px-4 py-3 font-mono text-xs text-foreground">
                        {order.id}
                      </td>
                      <td className="py-3 text-foreground">{order.customer}</td>
                      <td className="py-3">
                        <StatusPill tone={statusTone(order.status)}>
                          {t(`status.${order.status}`)}
                        </StatusPill>
                      </td>
                      <td className="py-3 font-medium text-foreground tabular-nums">{order.total}</td>
                      <td className="px-4 py-3 text-end">
                        <button
                          onClick={() => navigate.push(`/admin/orders/${order.id}`)}
                          className="text-xs font-medium text-muted-foreground hover:text-foreground hover:underline"
                        >
                          {t("dashboard.viewOrders")}
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </ResponsiveTable>
          </div>
        </section>

        <section className="rounded-lg border border-border bg-background">
          <div className="flex items-center justify-between border-b border-border p-4">
            <h2 className="text-sm font-semibold text-foreground">
              {t("dashboard.stockAlerts")}
            </h2>
            <button
              onClick={() => navigate.push("/admin/inventory")}
              className="text-xs font-medium text-muted-foreground hover:text-foreground"
            >
              {t("dashboard.inventory")}
            </button>
          </div>

          {lowStockItems.length === 0 ? (
            <p className="px-4 py-10 text-center text-sm text-muted-foreground">
              {t("dashboard.inventoryHealthy")}
            </p>
          ) : (
            <ul className="divide-y divide-border">
              {lowStockItems.map((item) => (
                <li key={item.variantId} className="flex items-center justify-between gap-3 px-4 py-3">
                  <div className="min-w-0">
                    <p className="truncate text-sm text-foreground">{item.name}</p>
                    <p className="truncate font-mono text-xs text-muted-foreground">{item.sku}</p>
                  </div>
                  <StatusPill tone={item.stock <= 0 ? "danger" : "warning"}>
                    {item.stock <= 0
                      ? t("dashboard.outOfStock")
                      : t("dashboard.unitsLeft", { count: item.stock })}
                  </StatusPill>
                </li>
              ))}
            </ul>
          )}

          <div className="border-t border-border p-3">
            <button
              onClick={() => navigate.push("/admin/inventory")}
              className="w-full rounded-md border border-border px-3 py-2 text-xs font-medium text-foreground transition-colors hover:bg-secondary"
            >
              {t("dashboard.manageStockAdjustments")}
            </button>
          </div>
        </section>
      </div>
    </div>
  );
}
