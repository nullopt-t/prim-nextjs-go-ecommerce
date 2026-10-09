"use client";

import { useEffect, useState, useMemo } from "react";
import { Link } from "@/i18n/navigation";
import { userService } from "@/services/user";
import { LazyImage } from "@/components/ui/lazyImage";
import {
  Package,
  Store,
  Search,
  Truck,
  RotateCcw,
  Star,
  CheckCircle2,
  Clock,
  ChevronRight,
  ChevronDown,
  ChevronUp,
  Receipt,
  Eye,
  SlidersHorizontal,
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

interface Address {
  street?: string;
  city?: string;
  state?: string;
  postalCode?: string;
  country?: string;
}

interface Order {
  id: string;
  customerEmail: string;
  shippingAddress?: Address;
  billingAddress?: Address;
  status: "pending" | "processing" | "shipped" | "delivered" | "cancelled";
  totalAmount: number;
  discountAmount: number;
  currency: string;
  items: OrderItem[];
  createdAt: string;
  updatedAt: string;
}

interface ParsedSnapshot {
  title?: string;
  name?: string;
  sku?: string;
  productId?: string;
  slug?: string;
  img?: string;
  image?: string;
  thumbnail?: string;
  thumbnailUrl?: string;
  price?: number;
  attributes?: Record<string, any>;
}

// ── Helpers ────────────────────────────────────────────────────────────────────
function parseSnapshot(snapshot?: string): ParsedSnapshot {
  if (!snapshot) return {};
  try {
    return JSON.parse(snapshot) ?? {};
  } catch {
    return {};
  }
}

function formatAddress(addr?: Address): string {
  if (!addr) return "";
  const parts = [addr.street, addr.city, addr.state, addr.country].filter(Boolean);
  return parts.join(", ");
}

function resolveItemImage(snap: ParsedSnapshot): string {
  if (snap.img) return snap.img;
  if (snap.image) return snap.image;
  if (snap.thumbnail) return snap.thumbnail;
  if (snap.thumbnailUrl) return snap.thumbnailUrl;
  return "";
}

const STATUS_CONFIG: Record<
  Order["status"],
  { label: string; badgeClass: string; icon: typeof CheckCircle2 }
> = {
  pending: {
    label: "Order Placed",
    badgeClass: "bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20",
    icon: Clock,
  },
  processing: {
    label: "Preparing for Dispatch",
    badgeClass: "bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20",
    icon: Package,
  },
  shipped: {
    label: "Shipped / On the way",
    badgeClass: "bg-violet-500/10 text-violet-600 dark:text-violet-400 border border-violet-500/20",
    icon: Truck,
  },
  delivered: {
    label: "Delivered",
    badgeClass: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20",
    icon: CheckCircle2,
  },
  cancelled: {
    label: "Cancelled",
    badgeClass: "bg-red-500/10 text-red-600 dark:text-red-400 border border-red-500/20",
    icon: RotateCcw,
  },
};

// ── Skeletons ──────────────────────────────────────────────────────────────────
function OrderSkeleton() {
  return (
    <div className="bg-card border border-border rounded-2xl overflow-hidden shadow-xs">
      <div className="bg-secondary/40 border-b border-border p-4 sm:p-5 flex justify-between animate-pulse">
        <div className="h-6 bg-secondary/80 rounded w-1/3" />
        <div className="h-6 bg-secondary/80 rounded w-1/4" />
      </div>
      <div className="p-6 space-y-4">
        <div className="h-20 bg-secondary/50 rounded-xl animate-pulse" />
        <div className="h-20 bg-secondary/50 rounded-xl animate-pulse" />
      </div>
    </div>
  );
}

// ── Amazon-style Order Box ─────────────────────────────────────────────────────
function AmazonOrderBox({ order, defaultExpanded = false }: { order: Order; defaultExpanded?: boolean }) {
  const [expanded, setExpanded] = useState(defaultExpanded);
  const statusCfg = STATUS_CONFIG[order.status] ?? STATUS_CONFIG.pending;
  const StatusIcon = statusCfg.icon;

  const formattedDate = new Date(order.createdAt).toLocaleDateString(undefined, {
    month: "long",
    day: "numeric",
    year: "numeric",
  });

  const totalFormatted = `${(order.totalAmount / 100).toFixed(2)} ${order.currency.toUpperCase()}`;
  const shortOrderId = order.id.slice(0, 8).toUpperCase();

  const items = order.items ?? [];
  const destinationAddress = formatAddress(order.shippingAddress);

  // Clean, non-repetitive status text
  let statusMessage = "";
  if (order.status === "delivered") {
    statusMessage = "Package was handed directly to resident or placed in secure location";
  } else if (order.status === "shipped") {
    statusMessage = "Package is in transit with the logistics carrier";
  } else if (order.status === "processing") {
    statusMessage = "Package is being prepared and packed for shipment";
  } else if (order.status === "cancelled") {
    statusMessage = "Order was cancelled and payment charge was refunded";
  } else {
    statusMessage = "Order received and queued for fulfillment";
  }

  return (
    <div className="bg-card border border-border rounded-xl sm:rounded-2xl overflow-hidden shadow-xs transition-all duration-200 hover:shadow-md hover:border-border/80">
      {/* ── Responsive Order Header Card ── */}
      <div className="bg-secondary/40 border-b border-border p-3.5 sm:px-5 sm:py-4">
        {/* Mobile top summary line */}
        <div className="flex sm:hidden items-center justify-between gap-2 pb-2.5 mb-2.5 border-b border-border/60 text-xs">
          <div className="flex items-center gap-2">
            <span className="font-semibold text-foreground">#{shortOrderId}</span>
            <span className="text-muted-foreground">•</span>
            <span className="text-muted-foreground">{formattedDate}</span>
          </div>
          <div className="flex items-center gap-2">
            <span className="font-bold text-foreground text-sm">{totalFormatted}</span>
            <button
              type="button"
              onClick={() => setExpanded((prev) => !prev)}
              aria-expanded={expanded}
              aria-label={expanded ? "Collapse order details" : "Expand order details"}
              className="p-1 rounded-md bg-secondary hover:bg-secondary/80 text-foreground cursor-pointer"
            >
              {expanded ? <ChevronUp className="size-4" /> : <ChevronDown className="size-4" />}
            </button>
          </div>
        </div>

        {/* Desktop & tablet header grid */}
        <div className="hidden sm:flex flex-wrap items-center justify-between gap-4 text-xs text-muted-foreground">
          <div className="flex flex-wrap items-center gap-6 sm:gap-10">
            <div>
              <span className="block uppercase text-[10px] font-semibold tracking-wider text-muted-foreground/80">
                Order Placed
              </span>
              <span className="font-medium text-foreground text-xs mt-0.5 block">
                {formattedDate}
              </span>
            </div>

            <div>
              <span className="block uppercase text-[10px] font-semibold tracking-wider text-muted-foreground/80">
                Total
              </span>
              <span className="font-semibold text-foreground text-xs mt-0.5 block">
                {totalFormatted}
              </span>
            </div>

            <div>
              <span className="block uppercase text-[10px] font-semibold tracking-wider text-muted-foreground/80">
                Shipped To
              </span>
              <span
                className="font-medium text-foreground text-xs mt-0.5 block truncate max-w-[240px]"
                title={destinationAddress || order.customerEmail}
              >
                {destinationAddress || order.customerEmail}
              </span>
            </div>
          </div>

          <div className="flex items-center gap-4 ml-auto">
            <div className="text-right">
              <span className="block uppercase text-[10px] font-semibold tracking-wider text-muted-foreground/80">
                Order Number
              </span>
              <span className="text-xs font-semibold text-foreground block mt-0.5">
                #{shortOrderId}
              </span>
            </div>

            <button
              type="button"
              onClick={() => setExpanded((prev) => !prev)}
              aria-expanded={expanded}
              aria-label={expanded ? "Collapse order details" : "Expand order details"}
              className="p-1.5 rounded-lg bg-secondary/80 hover:bg-secondary text-foreground hover:text-accent-brand transition-colors cursor-pointer border border-border/60 ml-1 shrink-0"
            >
              {expanded ? <ChevronUp className="size-4" /> : <ChevronDown className="size-4" />}
            </button>
          </div>
        </div>

        {/* Mobile destination address note */}
        {destinationAddress && (
          <div className="sm:hidden text-[11px] text-muted-foreground flex items-center gap-1.5 truncate">
            <span className="font-medium text-foreground/80 shrink-0">Ship to:</span>
            <span className="truncate">{destinationAddress}</span>
          </div>
        )}
      </div>

      {/* ── Status Banner (Touch-friendly accordion bar) ── */}
      <button
        type="button"
        onClick={() => setExpanded((prev) => !prev)}
        className="w-full text-left p-3 sm:px-5 sm:py-3.5 bg-background/50 border-b border-border/60 flex items-center justify-between gap-2.5 hover:bg-secondary/20 transition-colors cursor-pointer"
      >
        <div className="flex items-center gap-2 min-w-0">
          <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 sm:px-3 sm:py-1 rounded-full text-[11px] sm:text-xs font-semibold shrink-0 ${statusCfg.badgeClass}`}>
            <StatusIcon className="size-3.5 shrink-0" />
            <span>{statusCfg.label}</span>
          </span>
          <span className="text-[11px] sm:text-xs text-muted-foreground font-medium truncate hidden xs:inline">
            • {statusMessage}
          </span>
        </div>

        <div className="flex items-center gap-1.5 text-xs shrink-0 text-muted-foreground">
          <span className="hidden sm:inline">{items.length} item{items.length !== 1 ? "s" : ""} •</span>
          <span className="text-accent-brand font-semibold text-xs">
            {expanded ? "Hide" : "Details"}
          </span>
          <div className="sm:hidden text-accent-brand">
            {expanded ? <ChevronUp className="size-3.5" /> : <ChevronDown className="size-3.5" />}
          </div>
        </div>
      </button>

      {/* ── Order Line Items List ── */}
      {expanded && (
        <div className="divide-y divide-border/60 p-3.5 sm:p-5 space-y-4 sm:space-y-5 animate-in fade-in-50 duration-200">
          {items.map((item) => {
            const snap = parseSnapshot(item.productSnapshot);
            const title = snap.title || snap.name || "Purchased Item";
            const imgSrc = resolveItemImage(snap);
            const itemPrice = `${(item.priceAtPurchase / 100).toFixed(2)} ${order.currency.toUpperCase()}`;

            return (
              <div
                key={item.id}
                className="pt-4 first:pt-0 flex flex-col md:flex-row md:items-start justify-between gap-3.5 sm:gap-5"
              >
                {/* Product Thumbnail & Details */}
                <div className="flex items-start gap-3 sm:gap-4 flex-1 min-w-0">
                  <div className="size-16 sm:size-24 rounded-lg sm:rounded-xl border border-border/70 bg-secondary/30 overflow-hidden shrink-0 relative aspect-square flex items-center justify-center">
                    {imgSrc ? (
                      <LazyImage
                        src={imgSrc}
                        alt={title}
                        containerClassName="w-full h-full"
                        imageClassName="w-full h-full object-cover object-center"
                      />
                    ) : (
                      <div className="w-full h-full flex flex-col items-center justify-center p-1 sm:p-2 bg-secondary/50 text-muted-foreground/60 text-center">
                        <Package className="size-6 sm:size-8 text-muted-foreground/50 mb-0.5" />
                        <span className="text-[9px] sm:text-[10px] font-medium leading-tight text-muted-foreground/70">
                          Record
                        </span>
                      </div>
                    )}
                  </div>

                  <div className="space-y-1 flex-1 min-w-0">
                    <h4 className="font-semibold text-xs sm:text-sm text-foreground line-clamp-2 leading-snug">
                      {title}
                    </h4>

                    <div className="flex flex-wrap items-center gap-2 sm:gap-3 text-xs text-muted-foreground">
                      <span className="font-medium text-foreground">
                        Quantity: {item.quantity}
                      </span>
                      <span>•</span>
                      <span className="font-semibold text-accent-brand">
                        {itemPrice}
                      </span>
                    </div>

                    {order.status === "delivered" && (
                      <div className="inline-flex items-center gap-1 text-[11px] sm:text-xs text-emerald-600 dark:text-emerald-400 font-medium pt-0.5">
                        <CheckCircle2 className="size-3.5" />
                        <span>Verified Delivery</span>
                      </div>
                    )}
                  </div>
                </div>

                {/* Action Buttons for this item (Grid on mobile, stacked on desktop) */}
                <div className="grid grid-cols-2 md:flex md:flex-col items-stretch gap-2 shrink-0 md:w-44 pt-1 md:pt-0">
                  <Link
                    href="/products"
                    className="inline-flex items-center justify-center gap-1.5 px-3 py-2 rounded-lg sm:rounded-xl text-xs font-semibold bg-accent-brand text-accent-brand-foreground hover:opacity-90 transition-all shadow-xs"
                  >
                    <RotateCcw className="size-3.5" />
                    <span>Buy Again</span>
                  </Link>

                  <Link
                    href="/user/reviews"
                    className="inline-flex items-center justify-center gap-1.5 px-3 py-2 rounded-lg sm:rounded-xl text-xs font-medium bg-secondary hover:bg-secondary/80 text-foreground border border-border transition-all shadow-xs"
                  >
                    <Star className="size-3.5 text-amber-500" />
                    <span>Review</span>
                  </Link>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

// ── Main Component ─────────────────────────────────────────────────────────────
export default function OrdersContent() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Search & Filter state
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedStatus, setSelectedStatus] = useState<string>("all");

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

  // Filtered orders
  const filteredOrders = useMemo(() => {
    return orders
      .filter((order) => {
        if (selectedStatus !== "all" && order.status !== selectedStatus) {
          return false;
        }

        if (searchQuery.trim()) {
          const q = searchQuery.toLowerCase();
          const matchId = order.id.toLowerCase().includes(q);
          const matchItems = (order.items ?? []).some((item) => {
            const snap = parseSnapshot(item.productSnapshot);
            const title = (snap.title || snap.name || "").toLowerCase();
            const sku = (snap.sku || "").toLowerCase();
            return title.includes(q) || sku.includes(q);
          });
          return matchId || matchItems;
        }

        return true;
      })
      .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
  }, [orders, selectedStatus, searchQuery]);

  return (
    <div className="flex flex-col gap-6 pb-10">
      {/* ── Page Header ── */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-foreground tracking-tight">Your Orders</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Track packages, check receipt details, or reorder previous items.
          </p>
        </div>
        <Link
          href="/products"
          className="inline-flex items-center gap-2 px-4 py-2.5 rounded-xl text-xs font-semibold bg-secondary hover:bg-secondary/80 text-foreground transition-all border border-border shrink-0 self-start sm:self-auto shadow-xs"
        >
          <Store className="size-3.5 text-accent-brand" />
          <span>Continue Shopping</span>
        </Link>
      </div>

      {/* ── Filter & Search Toolbar (Amazon style) ── */}
      <div className="bg-card border border-border rounded-2xl p-4 shadow-xs flex flex-col md:flex-row items-stretch md:items-center justify-between gap-3">
        {/* Search by order number or product name */}
        <div className="relative flex-1">
          <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search your orders by product name or order number..."
            className="w-full pl-10 pr-4 py-2 rounded-xl bg-secondary/40 border border-border text-xs sm:text-sm text-foreground placeholder:text-muted-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30 transition-all"
          />
        </div>

        {/* Status filter buttons */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1 md:pb-0 scrollbar-none">
          {[
            { id: "all", label: "All Orders" },
            { id: "delivered", label: "Delivered" },
            { id: "shipped", label: "Shipped" },
            { id: "processing", label: "Processing" },
            { id: "cancelled", label: "Cancelled" },
          ].map((tab) => {
            const active = selectedStatus === tab.id;
            return (
              <button
                key={tab.id}
                type="button"
                onClick={() => setSelectedStatus(tab.id)}
                className={`px-3 py-1.5 rounded-xl text-xs font-medium transition-all shrink-0 cursor-pointer ${
                  active
                    ? "bg-accent-brand text-accent-brand-foreground shadow-xs font-semibold"
                    : "text-muted-foreground hover:bg-secondary hover:text-foreground"
                }`}
              >
                {tab.label}
              </button>
            );
          })}
        </div>
      </div>

      {/* ── Orders Content ── */}
      {loading ? (
        <div className="flex flex-col gap-4">
          <OrderSkeleton />
          <OrderSkeleton />
        </div>
      ) : error ? (
        <div className="p-4 rounded-xl border border-destructive/20 bg-destructive/10 text-destructive text-sm">
          {error}
        </div>
      ) : filteredOrders.length === 0 ? (
        <div className="bg-card border border-border rounded-2xl p-12 flex flex-col items-center justify-center text-center gap-3 shadow-xs">
          <Package className="size-12 text-muted-foreground/40" />
          <h3 className="font-semibold text-foreground text-base">
            {searchQuery || selectedStatus !== "all"
              ? "No orders match your search or filter"
              : "No orders placed yet"}
          </h3>
          <p className="text-xs text-muted-foreground max-w-sm">
            {searchQuery || selectedStatus !== "all"
              ? "Try clearing your search query or selecting 'All Orders'."
              : "Explore our collection and find what you need today."}
          </p>
          {searchQuery || selectedStatus !== "all" ? (
            <button
              type="button"
              onClick={() => {
                setSearchQuery("");
                setSelectedStatus("all");
              }}
              className="mt-2 text-xs font-semibold px-4 py-2 rounded-xl bg-secondary hover:bg-secondary/80 text-foreground border border-border cursor-pointer transition-colors"
            >
              Reset Filters
            </button>
          ) : (
            <Link
              href="/products"
              className="mt-2 text-xs font-semibold px-4 py-2.5 rounded-xl bg-accent-brand text-accent-brand-foreground hover:opacity-90 transition-all shadow-xs"
            >
              Start Shopping
            </Link>
          )}
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          {filteredOrders.map((order, idx) => (
            <AmazonOrderBox
              key={order.id}
              order={order}
              defaultExpanded={idx === 0}
            />
          ))}
        </div>
      )}
    </div>
  );
}
