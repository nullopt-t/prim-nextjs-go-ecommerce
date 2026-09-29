"use client";

import { useEffect, useState } from "react";
import { Link } from "@/i18n/navigation";
import { userService } from "@/services/user";
import { Package, ChevronDown, ChevronUp } from "lucide-react";

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

// ── Status badge map ───────────────────────────────────────────────────────────
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

// ── Helpers ────────────────────────────────────────────────────────────────────
function parseSnapshot(snapshot: string): { name?: string } {
  try {
    return JSON.parse(snapshot) ?? {};
  } catch {
    return {};
  }
}

// ── Skeleton card ──────────────────────────────────────────────────────────────
function OrderSkeleton() {
  return (
    <div className="animate-pulse bg-secondary rounded-lg h-20 w-full" />
  );
}

// ── Single order card ──────────────────────────────────────────────────────────
function OrderCard({ order }: { order: Order }) {
  const [expanded, setExpanded] = useState(false);
  const badge = STATUS_BADGE[order.status] ?? STATUS_BADGE.pending;
  const total = `${(order.totalAmount / 100).toFixed(2)} ${order.currency.toUpperCase()}`;
  const orderId = `#${order.id.slice(0, 8).toUpperCase()}`;
  const date = new Date(order.createdAt).toLocaleDateString();

  return (
    <div className="bg-card border border-border rounded-lg shadow-sm overflow-hidden">
      {/* Header row */}
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="w-full text-left px-4 py-3 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 hover:bg-secondary/30 transition-colors"
        aria-expanded={expanded}
      >
        <div className="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-3">
          <span className="text-sm font-semibold text-foreground">{orderId}</span>
          <span className="text-xs text-muted-foreground">{date}</span>
          <span className="text-xs text-muted-foreground">
            {order.items.length} item{order.items.length !== 1 ? "s" : ""}
          </span>
        </div>
        <div className="flex items-center gap-3">
          <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${badge.className}`}>
            {badge.label}
          </span>
          <span className="text-sm font-semibold text-foreground">{total}</span>
          {expanded ? (
            <ChevronUp className="size-4 text-muted-foreground shrink-0" />
          ) : (
            <ChevronDown className="size-4 text-muted-foreground shrink-0" />
          )}
        </div>
      </button>

      {/* Expanded items */}
      {expanded && (
        <div className="border-t border-border divide-y divide-border">
          {order.items.map((item) => {
            const snap = parseSnapshot(item.productSnapshot);
            const name = snap.name || item.variantId;
            return (
              <div
                key={item.id}
                className="px-4 py-2.5 flex items-center justify-between gap-3 text-sm"
              >
                <div className="flex flex-col">
                  <span className="text-foreground font-medium">{name}</span>
                  <span className="text-xs text-muted-foreground">
                    Qty: {item.quantity}
                  </span>
                </div>
                <span className="text-foreground font-medium shrink-0">
                  {(item.priceAtPurchase / 100).toFixed(2)}{" "}
                  {order.currency.toUpperCase()}
                </span>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

// ── Main component ─────────────────────────────────────────────────────────────
export default function OrdersContent() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const res = await userService.getOrders();
        if (cancelled) return;
        const raw: Order[] = Array.isArray(res?.data)
          ? res.data
          : Array.isArray(res)
          ? res
          : [];
        setOrders(raw);
      } catch (err: unknown) {
        if (!cancelled) setError(err instanceof Error ? err.message : "Failed to load orders.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="flex flex-col gap-6">
      {/* Page title */}
      <div>
        <h1 className="text-xl font-semibold text-foreground">My Orders</h1>
        <p className="text-sm text-muted-foreground mt-1">
          View and track all your recent orders.
        </p>
      </div>

      {/* Content */}
      {loading ? (
        <div className="flex flex-col gap-3">
          <OrderSkeleton />
          <OrderSkeleton />
          <OrderSkeleton />
        </div>
      ) : error ? (
        <p className="text-destructive text-sm">Error: {error}</p>
      ) : orders.length === 0 ? (
        <div className="bg-card border border-border rounded-lg p-12 flex flex-col items-center justify-center text-center gap-3 shadow-sm">
          <Package className="size-10 text-muted-foreground/50" />
          <p className="font-medium text-foreground">No orders yet</p>
          <p className="text-sm text-muted-foreground">
            Your orders will appear here once you&apos;ve made a purchase.
          </p>
          <Link
            href="/products"
            className="mt-2 h-9 px-4 inline-flex items-center rounded-md bg-accent-brand text-white text-sm font-medium hover:bg-accent-brand/90 transition shadow-sm"
          >
            Start Shopping
          </Link>
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          {orders
            .slice()
            .sort(
              (a, b) =>
                new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
            )
            .map((order) => (
              <OrderCard key={order.id} order={order} />
            ))}
        </div>
      )}
    </div>
  );
}
