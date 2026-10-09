"use client";

import { useEffect, useState, useMemo } from "react";
import { Link } from "@/i18n/navigation";
import { userService } from "@/services/user";
import { LazyImage } from "@/components/ui/lazyImage";
import { useTranslations, useLocale } from "next-intl";
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

// ── Order Card (Clickable Card with Product Thumbnails, No Extra Buttons) ─────
function OrderCard({ order }: { order: Order }) {
  const t = useTranslations("orders");
  const locale = useLocale();
  const statusCfg = STATUS_CONFIG[order.status] ?? STATUS_CONFIG.pending;

  const formattedDate = new Date(order.createdAt).toLocaleDateString(locale, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });

  const totalFormatted = `${(order.totalAmount / 100).toFixed(2)} ${order.currency.toUpperCase()}`;
  const shortOrderId = order.id.slice(0, 8).toUpperCase();
  const items = order.items ?? [];
  const destinationAddress = formatAddress(order.shippingAddress);

  // Extract snapshots, image icons, and product link targets
  const itemSnapshots = items.map((it) => {
    const snap = parseSnapshot(it.productSnapshot);
    const title = snap.title || snap.name || "Product";
    const rawImg = snap.image || snap.thumbnailUrl || snap.thumbnail || snap.img;
    let img = rawImg || "/placeholder-product.png";
    if (rawImg && !rawImg.startsWith("http") && !rawImg.startsWith("/")) {
      img = `http://localhost:9000/catalog/${rawImg}`;
    }
    const productHref = snap.slug ? `/products/${snap.slug}` : snap.productId ? `/products/${snap.productId}` : null;
    return { title, img, quantity: it.quantity, productHref };
  });

  return (
    <div className="group relative bg-card border border-border hover:border-accent-brand/50 rounded-2xl p-5 shadow-2xs hover:shadow-md transition-all duration-200">
      {/* Background card link to order details */}
      <Link
        href={`/user/orders/${order.id}`}
        className="absolute inset-0 z-0 rounded-2xl"
        aria-label={`View order ${shortOrderId}`}
      />

      <div className="relative z-1 flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-3.5 border-b border-border/50 pointer-events-none">
        <div className="flex flex-wrap items-center gap-2.5">
          <span className="font-mono font-bold text-sm text-foreground group-hover:text-accent-brand transition-colors">
            #{shortOrderId}
          </span>
          <span className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${statusCfg.badgeClass}`}>
            {t(`status.${order.status}`)}
          </span>
          <span className="text-xs text-muted-foreground">•</span>
          <span className="text-xs text-muted-foreground">
            {formattedDate}
          </span>
        </div>

        <div className="flex items-center gap-4 text-xs">
          <div>
            <span className="text-muted-foreground">{t("card.total")}: </span>
            <span className="font-bold text-foreground text-sm">{totalFormatted}</span>
          </div>
        </div>
      </div>

      <div className="relative z-1 pt-3.5 flex flex-col md:flex-row md:items-center justify-between gap-4">
        {/* Item image thumbnails / icons (clickable to product page) */}
        <div className="flex items-center gap-3.5 min-w-0">
          <div className="flex items-center -space-x-2.5 rtl:space-x-reverse shrink-0">
            {itemSnapshots.slice(0, 4).map((it, idx) => {
              const thumbnailElement = (
                <div
                  className="size-12 rounded-xl bg-secondary/80 border-2 border-card overflow-hidden shrink-0 shadow-2xs hover:scale-110 hover:z-10 transition-all duration-200 cursor-pointer"
                  title={it.title}
                >
                  <img
                    src={it.img}
                    alt={it.title}
                    className="w-full h-full object-cover"
                  />
                </div>
              );

              return it.productHref ? (
                <Link
                  key={idx}
                  href={it.productHref}
                  className="relative z-10 shrink-0"
                  onClick={(e) => e.stopPropagation()}
                >
                  {thumbnailElement}
                </Link>
              ) : (
                <div key={idx} className="relative shrink-0">
                  {thumbnailElement}
                </div>
              );
            })}
            {itemSnapshots.length > 4 && (
              <div className="size-12 rounded-xl bg-secondary border-2 border-card flex items-center justify-center text-xs font-semibold text-muted-foreground shrink-0 shadow-2xs">
                +{itemSnapshots.length - 4}
              </div>
            )}
          </div>

          <div className="flex flex-col min-w-0 text-xs">
            <span className="font-semibold text-foreground truncate pointer-events-none">
              {itemSnapshots.length} {itemSnapshots.length === 1 ? "item" : "items"}
            </span>
            <div className="text-muted-foreground truncate max-w-md flex flex-wrap items-center gap-1">
              {itemSnapshots.map((it, idx) => (
                <span key={idx}>
                  {it.productHref ? (
                    <Link
                      href={it.productHref}
                      className="relative z-10 hover:text-accent-brand hover:underline transition-colors"
                      onClick={(e) => e.stopPropagation()}
                    >
                      {it.title}
                    </Link>
                  ) : (
                    it.title
                  )}
                  {idx < itemSnapshots.length - 1 && ", "}
                </span>
              ))}
            </div>
          </div>
        </div>

        {destinationAddress && (
          <div className="text-xs text-muted-foreground truncate max-w-xs self-start md:self-center pointer-events-none">
            <span>{t("card.shippedTo")}: </span>
            <span className="font-medium text-foreground">{destinationAddress}</span>
          </div>
        )}
      </div>
    </div>
  );
}

export default function OrdersContent() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const t = useTranslations("orders");

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
          <h1 className="text-2xl font-bold text-foreground tracking-tight">{t("title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t("subtitle")}
          </p>
        </div>
        <Link
          href="/products"
          className="px-4 py-2 rounded-xl text-xs font-semibold bg-secondary hover:bg-secondary/80 text-foreground transition-all border border-border shrink-0 self-start sm:self-auto shadow-xs"
        >
          {t("continueShopping")}
        </Link>
      </div>

      {/* ── Filter & Search Toolbar (Amazon style) ── */}
      <div className="bg-card border border-border rounded-2xl p-4 shadow-xs flex flex-col md:flex-row items-stretch md:items-center justify-between gap-3">
        {/* Search by order number or product name */}
        <div className="relative flex-1">
          <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 size-4 text-muted-foreground rtl:left-auto rtl:right-3.5" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder={t("searchPlaceholder")}
            className="w-full pl-10 pr-4 rtl:pl-4 rtl:pr-10 py-2 rounded-xl bg-secondary/40 border border-border text-xs sm:text-sm text-foreground placeholder:text-muted-foreground focus:outline-hidden focus:ring-2 focus:ring-accent-brand/30 transition-all"
          />
        </div>

        {/* Status filter buttons */}
        <div className="flex items-center gap-1.5 overflow-x-auto pb-1 md:pb-0 scrollbar-none">
          {[
            { id: "all", label: t("tabs.all") },
            { id: "delivered", label: t("tabs.delivered") },
            { id: "shipped", label: t("tabs.shipped") },
            { id: "processing", label: t("tabs.processing") },
            { id: "cancelled", label: t("tabs.cancelled") },
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
            {t("empty.title")}
          </h3>
          <p className="text-xs text-muted-foreground max-w-sm">
            {t("empty.subtitle")}
          </p>
          <Link
            href="/products"
            className="mt-2 text-xs font-semibold px-4 py-2.5 rounded-xl bg-accent-brand text-accent-brand-foreground hover:opacity-90 transition-all shadow-xs"
          >
            {t("empty.action")}
          </Link>
        </div>
      ) : (
        <div className="flex flex-col gap-4">
          {filteredOrders.map((order) => (
            <OrderCard key={order.id} order={order} />
          ))}
        </div>
      )}
    </div>
  );
}
