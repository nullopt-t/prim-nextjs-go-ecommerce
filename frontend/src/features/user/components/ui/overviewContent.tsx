"use client";

import { useEffect, useState } from "react";
import { useAuthContext } from "@/context/AuthContext";
import { useWishlist } from "@/hooks/useWishlist";
import { Link } from "@/i18n/navigation";
import { userService } from "@/services/user";
import {
  Package,
  Heart,
  Star,
  MapPin,
  CreditCard,
  Settings,
  ShoppingBag,
  Store,
  ArrowRight,
  ShieldCheck,
  Truck,
  RotateCcw,
  CheckCircle2,
  Clock,
  Sparkles,
  ExternalLink,
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

function parseSnapshot(snapshot?: string): { title?: string; name?: string; sku?: string } {
  if (!snapshot) return {};
  try {
    return JSON.parse(snapshot) ?? {};
  } catch {
    return {};
  }
}

const STATUS_BADGE: Record<
  Order["status"],
  { label: string; className: string }
> = {
  pending: {
    label: "Pending",
    className: "text-amber-600 bg-amber-50 dark:bg-amber-950/30 dark:text-amber-400 border border-amber-200/50",
  },
  processing: {
    label: "Processing",
    className: "text-blue-600 bg-blue-50 dark:bg-blue-950/30 dark:text-blue-400 border border-blue-200/50",
  },
  shipped: {
    label: "Shipped",
    className: "text-violet-600 bg-violet-50 dark:bg-violet-950/30 dark:text-violet-400 border border-violet-200/50",
  },
  delivered: {
    label: "Delivered",
    className: "text-emerald-600 bg-emerald-50 dark:bg-emerald-950/30 dark:text-emerald-400 border border-emerald-200/50",
  },
  cancelled: {
    label: "Cancelled",
    className: "text-red-600 bg-red-50 dark:bg-red-950/30 dark:text-red-400 border border-red-200/50",
  },
};

// Amazon-style main hub tiles
const amazonHubCards = [
  {
    id: "hub-orders",
    title: "Your Orders",
    description: "Track packages, view details, or buy again",
    icon: Package,
    iconBg: "bg-amber-500/10 text-amber-600 dark:text-amber-400",
    path: "/user/orders",
    badge: "Most Visited",
  },
  {
    id: "hub-wishlist",
    title: "Wishlist & Saved",
    description: "View items you saved for later and price updates",
    icon: Heart,
    iconBg: "bg-rose-500/10 text-rose-600 dark:text-rose-400",
    path: "/user/wishlist",
  },
  {
    id: "hub-addresses",
    title: "Your Addresses",
    description: "Edit delivery addresses and default shipping destinations",
    icon: MapPin,
    iconBg: "bg-blue-500/10 text-blue-600 dark:text-blue-400",
    path: "/user/address",
  },
  {
    id: "hub-payment",
    title: "Payment Methods",
    description: "Manage saved cards, billing addresses, and invoices",
    icon: CreditCard,
    iconBg: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400",
    path: "/user/payment",
  },
  {
    id: "hub-reviews",
    title: "Customer Reviews",
    description: "Share feedback on verified purchases and view badges",
    icon: Star,
    iconBg: "bg-yellow-500/10 text-yellow-600 dark:text-yellow-400",
    path: "/user/reviews",
  },
  {
    id: "hub-settings",
    title: "Login & Security",
    description: "Edit name, password, email, and preferences",
    icon: Settings,
    iconBg: "bg-purple-500/10 text-purple-600 dark:text-purple-400",
    path: "/user/settings",
  },
];

const tbHeaders = ["Order #", "Placed On", "Items", "Total", "Status", "Actions"];

// ── Skeletons ──────────────────────────────────────────────────────────────────
function HubSkeleton() {
  return (
    <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-6">
      {Array.from({ length: 6 }).map((_, i) => (
        <div key={i} className="animate-pulse bg-secondary/60 rounded-xl h-28" />
      ))}
    </div>
  );
}

function TableSkeleton() {
  return (
    <div className="bg-card border border-border rounded-xl overflow-hidden shadow-xs">
      <div className="animate-pulse bg-secondary/50 h-10 border-b border-border" />
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="animate-pulse bg-secondary/40 rounded-lg h-14 mx-4 my-2.5" />
      ))}
    </div>
  );
}

// ── Component ──────────────────────────────────────────────────────────────────
export default function OverviewContent() {
  const { user } = useAuthContext();
  const { wishlistCount } = useWishlist();
  const firstName = user?.name?.split(" ")[0] || "Customer";

  const [greeting, setGreeting] = useState("Welcome");
  const [orders, setOrders] = useState<Order[]>([]);
  const [reviewCount, setReviewCount] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setGreeting(getGreeting());
  }, []);

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

        const rawOrders: Order[] = Array.isArray(ordersRes?.data)
          ? ordersRes.data
          : Array.isArray(ordersRes)
          ? ordersRes
          : [];

        const rawReviews: unknown[] = Array.isArray(reviewsRes?.data)
          ? reviewsRes.data
          : Array.isArray(reviewsRes)
          ? reviewsRes
          : [];

        setOrders(rawOrders);
        setReviewCount(rawReviews.length);
      } catch (err: unknown) {
        if (!cancelled) setError(err instanceof Error ? err.message : "Failed to load data.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, []);

  // Sorting & highlights
  const sortedOrders = [...orders].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );
  const latestOrder = sortedOrders[0] ?? null;
  const recentOrders = sortedOrders.slice(0, 5);

  const deliveredCount = orders.filter((o) => o.status === "delivered").length;
  const inTransitCount = orders.filter(
    (o) => o.status === "processing" || o.status === "shipped" || o.status === "pending"
  ).length;

  return (
    <div className="w-full flex flex-col gap-8 pb-8">
      {/* ── Amazon-style Account Greeting & Quick Banner ── */}
      <div className="relative overflow-hidden bg-gradient-to-r from-card via-card to-secondary/30 border border-border rounded-2xl p-6 sm:p-7 shadow-xs">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
          <div className="space-y-1.5 max-w-xl">
            <div className="inline-flex items-center gap-2 px-2.5 py-1 rounded-full text-xs font-semibold bg-accent-brand/10 text-accent-brand border border-accent-brand/20">
              <Sparkles className="size-3" />
              <span>Your Account Hub</span>
            </div>
            <h1 className="text-2xl sm:text-3xl font-bold text-foreground tracking-tight">
              {greeting}, {firstName}
            </h1>
            <p className="text-sm text-muted-foreground leading-relaxed">
              Manage your orders, returns, account preferences, and shopping experience all in one place.
            </p>
          </div>

          <div className="flex flex-wrap items-center gap-3 shrink-0">
            <Link
              href="/"
              className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-semibold bg-accent-brand text-accent-brand-foreground hover:opacity-90 transition-all shadow-xs"
            >
              <Store className="size-4" />
              <span>Return to Store</span>
            </Link>
            <Link
              href="/products"
              className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-semibold bg-secondary hover:bg-secondary/80 text-foreground transition-all border border-border shadow-xs"
            >
              <ShoppingBag className="size-4 text-muted-foreground" />
              <span>Browse Catalog</span>
            </Link>
          </div>
        </div>

        {/* Quick summary strip */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mt-6 pt-5 border-t border-border/80">
          <div className="flex items-center gap-3">
            <div className="size-9 rounded-lg bg-blue-500/10 text-blue-600 flex items-center justify-center shrink-0">
              <Package className="size-4" />
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Total Orders</p>
              <p className="text-lg font-bold text-foreground">{orders.length}</p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <div className="size-9 rounded-lg bg-amber-500/10 text-amber-600 flex items-center justify-center shrink-0">
              <Clock className="size-4" />
            </div>
            <div>
              <p className="text-xs text-muted-foreground">In Progress</p>
              <p className="text-lg font-bold text-foreground">{inTransitCount}</p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <div className="size-9 rounded-lg bg-emerald-500/10 text-emerald-600 flex items-center justify-center shrink-0">
              <CheckCircle2 className="size-4" />
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Delivered</p>
              <p className="text-lg font-bold text-foreground">{deliveredCount}</p>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <div className="size-9 rounded-lg bg-rose-500/10 text-rose-600 flex items-center justify-center shrink-0">
              <Heart className="size-4" />
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Saved Items</p>
              <p className="text-lg font-bold text-foreground">{wishlistCount}</p>
            </div>
          </div>
        </div>
      </div>

      {/* ── Latest Order Spotlight (Amazon "Buy it again / Track package" bar) ── */}
      {latestOrder && (
        <div className="bg-card border border-border rounded-2xl p-5 shadow-xs">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-border">
            <div className="flex items-center gap-3">
              <div className="size-10 rounded-xl bg-accent-brand/10 text-accent-brand flex items-center justify-center shrink-0">
                <Truck className="size-5" />
              </div>
              <div>
                <span className="text-xs font-semibold text-accent-brand uppercase tracking-wider">
                  Latest Order Spotlight
                </span>
                <p className="text-sm font-semibold text-foreground">
                  Order #{latestOrder.id.slice(0, 8).toUpperCase()} • {new Date(latestOrder.createdAt).toLocaleDateString()}
                </p>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold ${STATUS_BADGE[latestOrder.status]?.className}`}>
                {STATUS_BADGE[latestOrder.status]?.label ?? latestOrder.status}
              </span>
              <Link
                href="/user/orders"
                className="text-xs font-medium text-accent-brand hover:underline inline-flex items-center gap-1 ml-2"
              >
                Track package <ArrowRight className="size-3" />
              </Link>
            </div>
          </div>

          {/* Quick list of items inside latest order */}
          <div className="pt-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div className="space-y-1.5 flex-1 min-w-0">
              {(latestOrder.items ?? []).slice(0, 2).map((item) => {
                const snap = parseSnapshot(item.productSnapshot);
                const title = snap.title || snap.name || "Purchased Product";
                return (
                  <div key={item.id} className="flex items-center gap-2 text-sm text-foreground">
                    <CheckCircle2 className="size-4 text-emerald-500 shrink-0" />
                    <span className="truncate font-medium">{title}</span>
                    <span className="text-xs text-muted-foreground shrink-0">(Quantity: {item.quantity})</span>
                  </div>
                );
              })}
              {(latestOrder.items?.length ?? 0) > 2 && (
                <p className="text-xs text-muted-foreground pl-6">
                  + {(latestOrder.items?.length ?? 0) - 2} more items in this shipment
                </p>
              )}
            </div>

            <div className="flex items-center gap-3 shrink-0">
              <Link
                href="/products"
                className="px-3.5 py-2 rounded-lg text-xs font-semibold bg-secondary hover:bg-secondary/80 text-foreground border border-border inline-flex items-center gap-1.5 transition-colors"
              >
                <RotateCcw className="size-3.5" />
                <span>Buy Again</span>
              </Link>
              <Link
                href="/user/orders"
                className="px-3.5 py-2 rounded-lg text-xs font-semibold bg-accent-brand text-accent-brand-foreground hover:opacity-90 inline-flex items-center gap-1.5 transition-colors"
              >
                <span>View Order</span>
              </Link>
            </div>
          </div>
        </div>
      )}

      {/* ── Amazon-style Account Hub Grid ── */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-lg font-bold text-foreground tracking-tight">Your Account Sections</h2>
            <p className="text-xs text-muted-foreground">Jump directly to your personal services & records</p>
          </div>
        </div>

        {loading ? (
          <HubSkeleton />
        ) : (
          <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-6">
            {amazonHubCards.map((card) => (
              <Link
                key={card.id}
                href={card.path}
                className="group relative bg-card hover:bg-card/90 border border-border hover:border-accent-brand/40 rounded-2xl p-5 flex items-start gap-4 transition-all duration-200 hover:-translate-y-0.5 hover:shadow-md shadow-xs"
              >
                <div className={`size-12 rounded-xl flex items-center justify-center shrink-0 ${card.iconBg} transition-transform group-hover:scale-105`}>
                  <card.icon className="size-6" />
                </div>
                <div className="flex-1 min-w-0 pr-2">
                  <div className="flex items-center gap-2">
                    <h3 className="font-semibold text-sm text-foreground group-hover:text-accent-brand transition-colors">
                      {card.title}
                    </h3>
                    {card.badge && (
                      <span className="text-[10px] font-semibold uppercase tracking-wider px-1.5 py-0.5 rounded-sm bg-accent-brand/10 text-accent-brand">
                        {card.badge}
                      </span>
                    )}
                  </div>
                  <p className="text-xs text-muted-foreground line-clamp-2 mt-1 leading-relaxed">
                    {card.description}
                  </p>
                </div>
                <ArrowRight className="size-4 text-muted-foreground/40 group-hover:text-accent-brand group-hover:translate-x-0.5 transition-all shrink-0 mt-1" />
              </Link>
            ))}
          </div>
        )}
      </div>

      {/* ── Recent Orders Table ── */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-lg font-bold text-foreground tracking-tight">Recent Orders</h2>
            <p className="text-xs text-muted-foreground">Review and track your recent purchase history</p>
          </div>
          <Link
            href="/user/orders"
            className="text-xs font-semibold text-accent-brand hover:underline inline-flex items-center gap-1"
          >
            View all {orders.length} orders <ArrowRight className="size-3" />
          </Link>
        </div>

        {loading ? (
          <TableSkeleton />
        ) : error ? (
          <div className="p-4 rounded-xl border border-destructive/20 bg-destructive/10 text-destructive text-sm">
            {error}
          </div>
        ) : (
          <div className="bg-card border border-border rounded-2xl overflow-hidden shadow-xs">
            <div className="overflow-x-auto">
              <table className="w-full text-left">
                <thead className="bg-secondary/40 border-b border-border">
                  <tr>
                    {tbHeaders.map((th) => (
                      <th
                        key={th}
                        className="px-4 py-3.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider"
                      >
                        {th}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {recentOrders.length === 0 ? (
                    <tr>
                      <td colSpan={6} className="px-4 py-14 text-center">
                        <div className="flex flex-col items-center gap-3">
                          <ShoppingBag className="size-10 text-muted-foreground/40" />
                          <p className="font-semibold text-foreground text-sm">No recent orders found</p>
                          <p className="text-xs text-muted-foreground max-w-sm">
                            When you place an order, package tracking and receipt details will show up right here.
                          </p>
                          <Link
                            href="/products"
                            className="mt-2 text-xs font-semibold px-4 py-2 rounded-xl bg-accent-brand text-accent-brand-foreground hover:opacity-90 transition-all shadow-xs"
                          >
                            Explore Store Catalog
                          </Link>
                        </div>
                      </td>
                    </tr>
                  ) : (
                    recentOrders.map((order) => {
                      const badge = STATUS_BADGE[order.status] ?? STATUS_BADGE.pending;
                      const dateStr = new Date(order.createdAt).toLocaleDateString(undefined, {
                        month: "short",
                        day: "numeric",
                        year: "numeric",
                      });
                      return (
                        <tr
                          key={order.id}
                          className="text-sm hover:bg-secondary/30 transition-colors"
                        >
                          <td className="px-4 py-4 text-foreground font-semibold font-mono text-xs">
                            #{order.id.slice(0, 8).toUpperCase()}
                          </td>
                          <td className="px-4 py-4 text-muted-foreground text-xs">
                            {dateStr}
                          </td>
                          <td className="px-4 py-4 text-foreground text-xs">
                            <span className="font-medium">{order.items?.length ?? 0}</span> item{(order.items?.length ?? 0) !== 1 ? "s" : ""}
                          </td>
                          <td className="px-4 py-4 text-foreground font-semibold text-xs">
                            {(order.totalAmount / 100).toFixed(2)} {order.currency.toUpperCase()}
                          </td>
                          <td className="px-4 py-4">
                            <span
                              className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold ${badge.className}`}
                            >
                              {badge.label}
                            </span>
                          </td>
                          <td className="px-4 py-4 text-xs">
                            <Link
                              href="/user/orders"
                              className="text-accent-brand hover:underline font-semibold inline-flex items-center gap-1"
                            >
                              Details <ArrowRight className="size-3" />
                            </Link>
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

      {/* ── Customer Guarantees (Amazon-style confidence footer banner) ── */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 pt-2">
        <div className="bg-card border border-border rounded-xl p-4 flex items-center gap-3.5">
          <div className="size-10 rounded-lg bg-emerald-500/10 text-emerald-600 flex items-center justify-center shrink-0">
            <ShieldCheck className="size-5" />
          </div>
          <div>
            <h4 className="text-xs font-semibold text-foreground">Buyer Protection</h4>
            <p className="text-[11px] text-muted-foreground mt-0.5">Secure payments & encrypted customer data</p>
          </div>
        </div>

        <div className="bg-card border border-border rounded-xl p-4 flex items-center gap-3.5">
          <div className="size-10 rounded-lg bg-blue-500/10 text-blue-600 flex items-center justify-center shrink-0">
            <Truck className="size-5" />
          </div>
          <div>
            <h4 className="text-xs font-semibold text-foreground">Trackable Shipping</h4>
            <p className="text-[11px] text-muted-foreground mt-0.5">Real-time status updates on all shipments</p>
          </div>
        </div>

        <div className="bg-card border border-border rounded-xl p-4 flex items-center gap-3.5">
          <div className="size-10 rounded-lg bg-purple-500/10 text-purple-600 flex items-center justify-center shrink-0">
            <RotateCcw className="size-5" />
          </div>
          <div>
            <h4 className="text-xs font-semibold text-foreground">Hassle-Free Returns</h4>
            <p className="text-[11px] text-muted-foreground mt-0.5">30-day return policy on eligible orders</p>
          </div>
        </div>
      </div>
    </div>
  );
}
