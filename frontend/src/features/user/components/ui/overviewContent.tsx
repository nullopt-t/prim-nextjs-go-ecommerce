"use client";

import { useAuthContext } from "@/context/AuthContext";
import { Link } from "@/i18n/navigation";
import {
  Package,
  Clock,
  Heart,
  Star,
  MapPin,
  CreditCard,
  Settings,
  ShoppingBag,
} from "lucide-react";

function getGreeting() {
  const hour = new Date().getHours();
  if (hour < 12) return "Good morning";
  if (hour < 17) return "Good afternoon";
  return "Good evening";
}

const stats = [
  { id: "total-orders", icon: Package, label: "Total Orders", value: 12, color: "text-blue-500" },
  { id: "pending", icon: Clock, label: "Pending", value: 2, color: "text-amber-500" },
  { id: "wishlist", icon: Heart, label: "Wishlist", value: 6, color: "text-rose-500" },
  { id: "reviews", icon: Star, label: "Reviews", value: 5, color: "text-yellow-500" },
];

const quickLinks = [
  { id: "ql-orders", icon: Package, label: "Orders", path: "/user/orders" },
  { id: "ql-wishlist", icon: Heart, label: "Wishlist", path: "/user/wishlist" },
  { id: "ql-addresses", icon: MapPin, label: "Addresses", path: "/user/address" },
  { id: "ql-settings", icon: Settings, label: "Settings", path: "/user/settings" },
];

const recentOrders = [
  {
    id: "PR-00482",
    date: "Jun 10, 2026",
    items: 3,
    total: "$124.48",
    status: "Delivered",
  },
];

const tbHeaders = ["# Order", "Date", "Items", "Total", "Status"];

export default function OverviewContent() {
  const { user } = useAuthContext();
  const firstName = user?.name?.split(" ")[0] || "there";

  return (
    <div className="w-full flex flex-col gap-6">
      {/* Welcome header */}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold text-foreground">
            {getGreeting()}, {firstName}! 👋
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            Here&apos;s what&apos;s happening with your account.
          </p>
        </div>
        <span className="shrink-0 text-xs font-medium px-2.5 py-1 rounded-full bg-accent-brand/10 text-accent-brand">
          Customer
        </span>
      </div>

      {/* Stats grid */}
      <div className="grid gap-4 grid-cols-2 md:grid-cols-4">
        {stats.map((stat) => (
          <div
            key={stat.id}
            className="bg-card border border-border rounded-lg p-4 shadow-sm flex flex-col gap-2"
          >
            <stat.icon className={`size-5 ${stat.color}`} />
            <p className="text-xs text-muted-foreground">{stat.label}</p>
            <p className="text-2xl font-bold text-foreground leading-none">{stat.value}</p>
          </div>
        ))}
      </div>

      {/* Quick links */}
      <div>
        <h2 className="text-sm font-semibold text-foreground mb-3">Quick Access</h2>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          {quickLinks.map((ql) => (
            <Link
              key={ql.id}
              href={ql.path}
              className="bg-card border border-border rounded-lg p-4 flex flex-col items-center gap-2 text-sm font-medium text-foreground hover:bg-secondary/50 hover:border-accent-brand/30 transition-colors shadow-sm group"
            >
              <ql.icon className="size-5 text-muted-foreground group-hover:text-accent-brand transition-colors" />
              <span>{ql.label}</span>
            </Link>
          ))}
        </div>
      </div>

      {/* Recent Orders */}
      <div>
        <h2 className="text-sm font-semibold text-foreground mb-3">Recent Orders</h2>
        <div className="bg-card border border-border rounded-lg overflow-hidden shadow-sm">
          <div className="overflow-x-auto">
            <table className="w-full text-left">
              <thead className="bg-secondary/50 border-b border-border">
                <tr>
                  {tbHeaders.map((th) => (
                    <th
                      key={th}
                      className="px-4 py-3 text-xs font-medium text-muted-foreground uppercase tracking-wide"
                    >
                      {th}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {recentOrders.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-12 text-center">
                      <div className="flex flex-col items-center gap-3">
                        <ShoppingBag className="size-10 text-muted-foreground/40" />
                        <p className="font-medium text-foreground text-sm">No recent orders</p>
                        <p className="text-xs text-muted-foreground">
                          When you place an order, it will appear here.
                        </p>
                        <Link
                          href="/products"
                          className="mt-1 text-xs font-medium text-accent-brand hover:underline"
                        >
                          Start shopping →
                        </Link>
                      </div>
                    </td>
                  </tr>
                ) : (
                  recentOrders.map((order) => (
                    <tr
                      key={order.id}
                      className="text-sm hover:bg-secondary/30 transition-colors"
                    >
                      <td className="px-4 py-3 text-foreground font-medium">{order.id}</td>
                      <td className="px-4 py-3 text-muted-foreground">{order.date}</td>
                      <td className="px-4 py-3 text-foreground">{order.items}</td>
                      <td className="px-4 py-3 text-foreground font-medium">{order.total}</td>
                      <td className="px-4 py-3">
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium text-emerald-600 bg-emerald-50 dark:bg-emerald-950/30 dark:text-emerald-400">
                          {order.status}
                        </span>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}
