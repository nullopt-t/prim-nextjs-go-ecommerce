import { useState } from "react";
import { useRouter } from "@/i18n/navigation";
import {
  DollarSign,
  ShoppingBag,
  Package,
  Users,
  TrendingUp,
  AlertTriangle,
  ArrowUpRight,
  Plus,
  Tag,
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
import {
  INITIAL_ORDERS,
  INITIAL_PRODUCTS,
  getLocalAdminData
} from "../data/mockAdminData";

const CHART_DATA_MONTHLY = [
  { month: "Jan", sales: 28400, orders: 240 },
  { month: "Feb", sales: 32100, orders: 280 },
  { month: "Mar", sales: 29800, orders: 260 },
  { month: "Apr", sales: 38900, orders: 320 },
  { month: "May", sales: 42500, orders: 380 },
  { month: "Jun", sales: 48920, orders: 410 },
  { month: "Jul", sales: 52100, orders: 450 },
  { month: "Aug", sales: 61400, orders: 520 }
];

export function AdminDashboard() {
  const navigate = useRouter();
  const [timeframe, setTimeframe] = useState("monthly");
  
  const products = getLocalAdminData("products", INITIAL_PRODUCTS);
  const orders = getLocalAdminData("orders", INITIAL_ORDERS);

  // Compute live low stock products
  const lowStockItems = products.filter(
    (p) => p.stock <= (p.lowStockThreshold || 10)
  );

  return (
    <div className="space-y-6">
      {/* Header Banner */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-background p-6 rounded-2xl border border-border shadow-xs">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Dashboard Overview</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Real-time business performance, sales metrics, and store activity.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => navigate.push("/admin/products")}
            className="flex items-center gap-2 px-4 py-2 bg-accent-brand text-white font-semibold text-xs rounded-xl shadow-xs hover:bg-accent-brand/90 transition-all"
          >
            <Plus className="size-4" />
            <span>Add Product</span>
          </button>
          <button
            onClick={() => navigate.push("/admin/promotions")}
            className="flex items-center gap-2 px-4 py-2 bg-secondary text-secondary-foreground font-semibold text-xs rounded-xl border border-border hover:bg-secondary/80 transition-all"
          >
            <Tag className="size-4" />
            <span>New Promotion</span>
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Total Revenue</span>
            <div className="size-9 rounded-xl bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-400 flex items-center justify-center">
              <DollarSign className="size-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-2xl font-bold text-foreground">$61,400.00</div>
            <div className="flex items-center gap-1 mt-1 text-xs text-emerald-600 dark:text-emerald-400 font-semibold">
              <TrendingUp className="size-3.5" />
              <span>+18.4% from last month</span>
            </div>
          </div>
        </div>

        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Total Orders</span>
            <div className="size-9 rounded-xl bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400 flex items-center justify-center">
              <ShoppingBag className="size-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-2xl font-bold text-foreground">520</div>
            <div className="flex items-center gap-1 mt-1 text-xs text-emerald-600 dark:text-emerald-400 font-semibold">
              <TrendingUp className="size-3.5" />
              <span>+12.1% from last month</span>
            </div>
          </div>
        </div>

        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Active Inventory</span>
            <div className="size-9 rounded-xl bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-400 flex items-center justify-center">
              <Package className="size-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-2xl font-bold text-foreground">{products.length} Products</div>
            <div className="flex items-center gap-1 mt-1 text-xs text-amber-600 dark:text-amber-400 font-semibold">
              <AlertTriangle className="size-3.5" />
              <span>{lowStockItems.length} items low in stock</span>
            </div>
          </div>
        </div>

        <div className="bg-background border border-border rounded-2xl p-5 shadow-xs flex flex-col justify-between">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Active Customers</span>
            <div className="size-9 rounded-xl bg-purple-100 text-purple-700 dark:bg-purple-900/40 dark:text-purple-400 flex items-center justify-center">
              <Users className="size-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-2xl font-bold text-foreground">1,248</div>
            <div className="flex items-center gap-1 mt-1 text-xs text-emerald-600 dark:text-emerald-400 font-semibold">
              <TrendingUp className="size-3.5" />
              <span>+9.5% new signups</span>
            </div>
          </div>
        </div>
      </div>

      {/* Analytics Chart */}
      <div className="bg-background border border-border rounded-2xl p-6 shadow-xs">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
          <div>
            <h2 className="text-lg font-bold text-foreground">Revenue & Sales Trends</h2>
            <p className="text-xs text-muted-foreground">Monthly gross revenue and total completed orders.</p>
          </div>
          <div className="flex items-center gap-2 bg-secondary/50 p-1 rounded-xl border border-border">
            <button
              onClick={() => setTimeframe("monthly")}
              className={`px-3 py-1 text-xs font-semibold rounded-lg transition-all ${
                timeframe === "monthly" ? "bg-background text-foreground shadow-xs" : "text-muted-foreground hover:text-foreground"
              }`}
            >
              Revenue ($)
            </button>
            <button
              onClick={() => setTimeframe("orders")}
              className={`px-3 py-1 text-xs font-semibold rounded-lg transition-all ${
                timeframe === "orders" ? "bg-background text-foreground shadow-xs" : "text-muted-foreground hover:text-foreground"
              }`}
            >
              Order Count
            </button>
          </div>
        </div>

        <div className="h-72 w-full">
          <ResponsiveContainer width="100%" height="100%">
            {timeframe === "monthly" ? (
              <AreaChart data={CHART_DATA_MONTHLY} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorSales" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#6366f1" stopOpacity={0.4} />
                    <stop offset="95%" stopColor="#6366f1" stopOpacity={0.0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--color-border, #e5e7eb)" />
                <XAxis dataKey="month" tick={{ fontSize: 12, fill: "gray" }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fontSize: 12, fill: "gray" }} axisLine={false} tickLine={false} tickFormatter={(val) => `$${val / 1000}k`} />
                <Tooltip
                  formatter={(val) => [`$${val.toLocaleString()}`, "Gross Revenue"]}
                  contentStyle={{ backgroundColor: "var(--color-bg, #ffffff)", borderRadius: "12px", border: "1px solid #e5e7eb" }}
                />
                <Area type="monotone" dataKey="sales" stroke="#6366f1" strokeWidth={3} fillOpacity={1} fill="url(#colorSales)" />
              </AreaChart>
            ) : (
              <BarChart data={CHART_DATA_MONTHLY} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--color-border, #e5e7eb)" />
                <XAxis dataKey="month" tick={{ fontSize: 12, fill: "gray" }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fontSize: 12, fill: "gray" }} axisLine={false} tickLine={false} />
                <Tooltip
                  formatter={(val) => [val, "Orders"]}
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
                <h3 className="font-bold text-base text-foreground">Recent Customer Orders</h3>
                <p className="text-xs text-muted-foreground">Latest purchases requiring processing and fulfillment.</p>
              </div>
              <button
                onClick={() => navigate.push("/admin/orders")}
                className="text-xs font-semibold text-accent-brand hover:underline flex items-center gap-1"
              >
                <span>View All Orders</span>
                <ArrowUpRight className="size-3.5" />
              </button>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-border text-muted-foreground font-semibold uppercase tracking-wider">
                    <th className="pb-3 pl-2">Order ID</th>
                    <th className="pb-3">Customer</th>
                    <th className="pb-3">Status</th>
                    <th className="pb-3">Total</th>
                    <th className="pb-3 pr-2 text-right">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/60">
                  {orders.slice(0, 5).map((order) => (
                    <tr key={order.id} className="hover:bg-secondary/20 transition-colors">
                      <td className="py-3.5 pl-2 font-mono font-bold text-foreground">{order.id}</td>
                      <td className="py-3.5">
                        <div className="font-medium text-foreground">{order.customer}</div>
                        <div className="text-[10px] text-muted-foreground">{order.email}</div>
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
                          title="View Order Details"
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
                <h3 className="font-bold text-base text-foreground">Stock Alerts</h3>
                <p className="text-xs text-muted-foreground">Products reaching critical inventory threshold.</p>
              </div>
              <button
                onClick={() => navigate.push("/admin/inventory")}
                className="text-xs font-semibold text-accent-brand hover:underline"
              >
                Inventory
              </button>
            </div>

            <div className="space-y-3">
              {lowStockItems.length > 0 ? (
                lowStockItems.map((prod) => (
                  <div
                    key={prod.id}
                    className="p-3 border border-border rounded-xl bg-secondary/10 flex items-center justify-between"
                  >
                    <div className="flex items-center gap-3">
                      <img src={prod.image} alt={prod.name} className="size-10 rounded-lg object-cover border border-border" />
                      <div>
                        <div className="font-semibold text-xs text-foreground line-clamp-1">{prod.name}</div>
                        <div className="text-[10px] text-muted-foreground font-mono">{prod.sku}</div>
                      </div>
                    </div>
                    <div className="text-right shrink-0">
                      <span className="text-xs font-bold text-amber-600 dark:text-amber-400">{prod.stock} left</span>
                    </div>
                  </div>
                ))
              ) : (
                <div className="text-center text-xs text-muted-foreground py-6">All products have sufficient stock levels.</div>
              )}
            </div>
          </div>

          <button
            onClick={() => navigate.push("/admin/inventory")}
            className="w-full mt-4 py-2.5 bg-secondary text-secondary-foreground text-xs font-semibold rounded-xl border border-border hover:bg-secondary/80 transition-colors"
          >
            Manage Stock Adjustments
          </button>
        </div>
      </div>
    </div>
  );
}
