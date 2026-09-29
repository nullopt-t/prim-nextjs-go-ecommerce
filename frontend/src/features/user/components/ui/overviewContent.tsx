"use client";

import { useEffect, useState } from "react";
import { useAuthContext } from "@/context/AuthContext";
import { Link } from "@/i18n/navigation";
import { userService } from "@/services/user";
import {
  Package,
  Clock,
  Heart,
  Star,
  MapPin,
  Settings,
  ShoppingBag,
} from "lucide-react";

// ── Types ──────────────────────────────────────────────────────────────────────
interface OrderItem {
  id: string;
  variantId: string;
  quantity: number;
  priceAtPurchase: number;
  productSnapshot: string;
}

interface Order {
  id: string;
  customerEmail: string;
  status: "pending" | "processing" | "shipped" | "delivered" | "cancelled";
  totalAmount: number;
  discountAmount: number;
  currency: string;
  items: OrderItem[];
  createdAt: string;
  updatedAt: string;
}

// ── Helpers ────────────────────────────────────────────────────────────────────
function getGreeting() {
  const hour = new Date().getHours();
  if (hour < 12) return "Good morning";
  if (hour < 17) return "Good afternoon";
  return "Good evening";
}

const STATUS_BADGE: Record<
  Order["status"],
  { label: string; className: string }
> = {
  pending: {
    label: "Pending",
    className:
      "text-amber-600 bg-amber-50 dark:bg-amber-950/30 dark:text-amber-400",
  },
  processing: {
    label: "Processing",
    className:
      "text-blue-600 bg-blue-50 dark:bg-blue-950/30 dark:text-blue-400",
  },
  shipped: {
    label: "Shipped",
    className:
      "text-violet-600 bg-violet-50 dark:bg-violet-950/30 dark:text-violet-400",
  },
  delivered: {
    label: "Delivered",
    className:
      "text-emerald-600 bg-emerald-50 dark:bg-emerald-950/30 dark:text-emerald-400",
  },
  cancelled: {
    label: "Cancelled",
    className: "text-red-600 bg-red-50 dark:bg-red-950/30 dark:text-red-400",
  },
};

const quickLinks = [
  { id: "ql-orders", icon: Package, label: "Orders", path: "/user/orders" },
  { id: "ql-wishlist", icon: Heart, label: "Wishlist", path: "/user/wishlist" },
  { id: "ql-addresses", icon: MapPin, label: "Addresses", path: "/user/address" },
  { id: "ql-settings", icon: Settings, label: "Settings", path: "/user/settings" },
];

const tbHeaders = ["# Order", "Date", "Items", "Total", "Status"];

// ── Skeleton ───────────────────────────────────────────────────────────────────
function StatsSkeleton() {
  return (
    <div className="grid gap-4 grid-cols-2 md:grid-cols-4">
      {Array.from({ length: 4 }).map((_, i) => (
        <div
          key={i}
          className="animate-pulse bg-secondary rounded-lg h-24"
        />
      ))}
    </div>
  );
}

function TableSkeleton() {
  return (
    <div className="bg-card border border-border rounded-lg overflow-hidden shadow-sm">
      <div className="animate-pulse bg-secondary/50 h-10 border-b border-border" />
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="animate-pulse bg-secondary rounded h-12 mx-4 my-2" />
      ))}
    </div>
  );
}

// ── Component ──────────────────────────────────────────────────────────────────
export default function OverviewContent() {
  const { user } = useAuthContext();
  const firstName = user?.name?.split(" ")[0] || "there";

  const [orders, setOrders] = useState<Order[]>([]);
  const [reviewCount, setReviewCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const [ordersRes, reviewsRes] = await Promise.all([
          userService.getOrders(),
          userService.getMyReviews(),
        ]);

        if (cancelled) return;

        // Orders endpoint: { data: { orders: [] }, meta: {...} }
        const rawOrders: Order[] =
          ordersRes?.data?.orders ?? ordersRes?.orders ?? ordersRes ?? [];

        // Reviews endpoint returns array directly
        const rawReviews: unknown[] = Array.isArray(reviewsRes)
          ? reviewsRes
          : reviewsRes?.data ?? [];

        setOrders(rawOrders);
        setReviewCount(rawReviews.length);
      } catch (err: any) {
        if (!cancelled) setError(err?.message || "Failed to load data.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, []);

  // Derived stats
  const totalOrders = orders.length;
  const pendingOrders = orders.filter(
    (o) => o.status === "pending" || o.status === "processing"
  ).length;
  const recentOrders = [...orders]
    .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
    .slice(0, 5);

  const stats = [
    { id: "total-orders", icon: Package, label: "Total Orders", value: totalOrders, color: "text-blue-500" },
    { id: "pending", icon: Clock, label: "Pending", value: pendingOrders, color: "text-amber-500" },
    { id: "wishlist", icon: Heart, label: "Wishlist", value: 0, color: "text-rose-500" },
    { id: "reviews", icon: Star, label: "Reviews", value: reviewCount, color: "text-yellow-500" },
  ];

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
      {loading ? (
        <StatsSkeleton />
      ) : (
        <div className="grid gap-4 grid-cols-2 md:grid-cols-4">
          {stats.map((stat) => (
            <div
              key={stat.id}
              className="bg-card border border-border rounded-lg p-4 shadow-sm flex flex-col gap-2"
            >
              <stat.icon className={`size-5 ${stat.color}`} />
              <p className="text-xs text-muted-foreground">{stat.label}</p>
              <p className="text-2xl font-bold text-foreground leading-none">
                {stat.value}
              </p>
            </div>
          ))}
        </div>
      )}

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

        {loading ? (
          <TableSkeleton />
        ) : error ? (
          <p className="text-destructive text-sm">Error: {error}</p>
        ) : (
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
                          <p className="font-medium text-foreground text-sm">No orders yet</p>
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
                    recentOrders.map((order) => {
                      const badge = STATUS_BADGE[order.status] ?? STATUS_BADGE.pending;
                      return (
                        <tr
                          key={order.id}
                          className="text-sm hover:bg-secondary/30 transition-colors"
                        >
                          <td className="px-4 py-3 text-foreground font-medium">
                            #{order.id.slice(0, 8).toUpperCase()}
                          </td>
                          <td className="px-4 py-3 text-muted-foreground">
                            {new Date(order.createdAt).toLocaleDateString()}
                          </td>
                          <td className="px-4 py-3 text-foreground">
                            {order.items.length}
                          </td>
                          <td className="px-4 py-3 text-foreground font-medium">
                            {(order.totalAmount / 100).toFixed(2)}{" "}
                            {order.currency.toUpperCase()}
                          </td>
                          <td className="px-4 py-3">
                            <span
                              className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${badge.className}`}
                            >
                              {badge.label}
                            </span>
                          </td>
                        </tr>
                      );
                    })
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
