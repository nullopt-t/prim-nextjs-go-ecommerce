"use client";

import { useEffect, useState } from "react";
import { useAuthContext } from "@/context/AuthContext";
import { Link } from "@/i18n/navigation";
import { userService } from "@/services/user";
import { wishlistService, WishlistItem } from "@/services/wishlist";
import { useTranslations, useLocale } from "next-intl";
import {
  Package,
  Heart,
  MapPin,
  CreditCard,
  Star,
  Settings,
  ChevronRight,
  ChevronLeft,
  ArrowRight,
  ArrowLeft,
  ShoppingBag,
  ExternalLink,
  RotateCcw,
  CheckCircle2,
  Clock,
  Truck,
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

import { normalizeMediaUrl } from "@/utils/media";

function getGreeting() {
  const hour = new Date().getHours();
  if (hour < 12) return "Good morning";
  if (hour < 17) return "Good afternoon";
  return "Good evening";
}

function parseSnapshot(snapshot?: string): { title: string; img: string } {
  if (!snapshot) return { title: "Product", img: "/placeholder-product.png" };
  try {
    const snap = JSON.parse(snapshot) ?? {};
    const title = snap.title || snap.name || "Product";
    const rawImg = snap.image || snap.thumbnailUrl || snap.thumbnail || snap.img;
    const img = normalizeMediaUrl(rawImg);
    return { title, img };
  } catch {
    return { title: "Product", img: "/placeholder-product.png" };
  }
}

const STATUS_BADGE: Record<
  Order["status"],
  { label: string; icon: typeof Clock; className: string }
> = {
  pending: {
    label: "Pending",
    icon: Clock,
    className: "text-amber-600 bg-amber-500/10 border-amber-500/20",
  },
  processing: {
    label: "Processing",
    icon: Clock,
    className: "text-blue-600 bg-blue-500/10 border-blue-500/20",
  },
  shipped: {
    label: "Shipped",
    icon: Truck,
    className: "text-violet-600 bg-violet-500/10 border-violet-500/20",
  },
  delivered: {
    label: "Delivered",
    icon: CheckCircle2,
    className: "text-emerald-600 bg-emerald-500/10 border-emerald-500/20",
  },
  cancelled: {
    label: "Cancelled",
    icon: Clock,
    className: "text-destructive bg-destructive/10 border-destructive/20",
  },
};

export default function OverviewContent() {
  const { user } = useAuthContext();
  const t = useTranslations("overview");
  const locale = useLocale();
  const isRtl = locale === "ar";
  const Chevron = isRtl ? ChevronLeft : ChevronRight;
  const Arrow = isRtl ? ArrowLeft : ArrowRight;

  const [orders, setOrders] = useState<Order[]>([]);
  const [wishlistPreview, setWishlistPreview] = useState<WishlistItem[]>([]);
  const [wishlistCount, setWishlistCount] = useState(0);
  const [reviewCount, setReviewCount] = useState(0);
  const [addressCount, setAddressCount] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      try {
        const [ordersRes, reviewsRes, addressesRes, wishlistRes] = await Promise.all([
          userService.getOrders().catch(() => ({ data: [] })),
          userService.getMyReviews().catch(() => ({ data: [] })),
          userService.getAddresses().catch(() => ({ data: [] })),
          wishlistService.getItems(1, 4).catch(() => ({ data: [], meta: { totalItems: 0 } })),
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

        const rawAddresses: unknown[] = Array.isArray(addressesRes?.data)
          ? addressesRes.data
          : Array.isArray(addressesRes)
          ? addressesRes
          : [];

        const rawWishlist: WishlistItem[] = Array.isArray(wishlistRes?.data)
          ? wishlistRes.data
          : [];

        setOrders(rawOrders);
        setReviewCount(rawReviews.length);
        setAddressCount(rawAddresses.length);
        setWishlistPreview(rawWishlist);
        setWishlistCount(wishlistRes?.meta?.totalItems ?? rawWishlist.length);
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, []);

  const sortedOrders = [...orders].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );
  // Current active orders waiting for fulfillment or delivery
  const activeOrders = sortedOrders.filter(
    (o) => o.status === "pending" || o.status === "processing" || o.status === "shipped"
  );
  const activeOrdersPreview = activeOrders.slice(0, 4);

  // Secondary account shortcuts: if there are no active in-progress orders,
  // include an Orders tile here so the user can easily access their past/all orders.
  const secondaryAccountTiles = [
    ...(activeOrders.length === 0
      ? [
          {
            id: "orders",
            title: t("accountShortcuts.orders"),
            desc: t("accountShortcuts.ordersDesc"),
            href: "/user/orders",
            icon: Package,
          },
        ]
      : []),
    {
      id: "addresses",
      title: t("accountShortcuts.addresses"),
      desc: t("accountShortcuts.addressesDesc"),
      badge: addressCount > 0 ? `${addressCount}` : undefined,
      href: "/user/address",
      icon: MapPin,
    },
    {
      id: "payment",
      title: t("accountShortcuts.payment"),
      desc: t("accountShortcuts.paymentDesc"),
      href: "/user/payment",
      icon: CreditCard,
    },
    {
      id: "reviews",
      title: t("accountShortcuts.reviews"),
      desc: t("accountShortcuts.reviewsDesc"),
      badge: reviewCount > 0 ? `${reviewCount}` : undefined,
      href: "/user/reviews",
      icon: Star,
    },
    {
      id: "settings",
      title: t("accountShortcuts.settings"),
      desc: t("accountShortcuts.settingsDesc"),
      href: "/user/settings",
      icon: Settings,
    },
  ];

  return (
    <div className="w-full flex flex-col gap-6 pb-16">
      {/* ── 1. Greeting & Identity Card ── */}
      <div className="bg-card border border-border/80 rounded-2xl p-4.5 sm:p-5 shadow-2xs flex items-center justify-between gap-4">
        <div className="flex items-center gap-3.5 min-w-0">
          <div className="size-11 sm:size-12 rounded-full bg-accent-brand/10 text-accent-brand flex items-center justify-center font-bold text-base sm:text-lg border border-accent-brand/20 shrink-0">
            {(user?.name || "U")[0]?.toUpperCase()}
          </div>
          <div className="min-w-0">
            <h1 className="text-lg sm:text-xl font-bold text-foreground tracking-tight truncate">
              {t("title", { greeting: getGreeting(), name: user?.name || t("defaultUser") })}
            </h1>
            <p className="text-xs text-muted-foreground mt-0.5 truncate font-mono">
              {user?.email || "—"}
            </p>
          </div>
        </div>
      </div>

      {/* ── 2. Open Orders Sub-Section with Live Preview (Hidden when no active orders) ── */}
      {(loading || activeOrdersPreview.length > 0) && (
        <div className="flex flex-col gap-3">
          <div className="flex items-center justify-between">
            <div className="flex items-baseline gap-2">
              <h2 className="text-sm sm:text-base font-bold text-foreground tracking-tight">
                {t("yourOrders.title")}
              </h2>
              {activeOrders.length > 0 && (
                <span className="text-xs text-muted-foreground">({activeOrders.length})</span>
              )}
            </div>
            <Link
              href="/user/orders"
              className="inline-flex items-center gap-1 text-xs font-semibold text-accent-brand hover:underline"
            >
              <span>{t("yourOrders.seeAll")}</span>
              <Arrow className="size-3.5" />
            </Link>
          </div>

          {loading ? (
            <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
              {Array.from({ length: 4 }).map((_, i) => (
                <div key={i} className="animate-pulse bg-secondary/50 rounded-2xl aspect-square border border-border/60" />
              ))}
            </div>
          ) : (
            <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
              {activeOrdersPreview.map((order) => {
                const items = order.items ?? [];
                const snaps = items.map((it) => parseSnapshot(it.productSnapshot));
                const firstSnap = snaps[0] ?? { title: "Product", img: "/placeholder-product.png" };

                return (
                  <Link
                    key={order.id}
                    href={`/user/orders/${order.id}`}
                    className="group relative block aspect-square w-full rounded-2xl bg-secondary/35 border border-border/80 hover:border-accent-brand/50 p-2.5 overflow-hidden shadow-2xs hover:shadow-xs transition-all"
                  >
                    {items.length <= 1 ? (
                      // Single item: fills the card cleanly
                      <div className="w-full h-full rounded-xl overflow-hidden bg-background border border-border/40">
                        <img
                          src={firstSnap.img}
                          alt={firstSnap.title}
                          className="w-full h-full object-cover object-center group-hover:scale-105 transition-transform duration-300"
                        />
                      </div>
                    ) : items.length === 2 ? (
                      // 2 items: large angled overlapping pile filling card
                      <div className="relative w-full h-full flex items-center justify-center">
                        {/* Back / 2nd item */}
                        <div className="absolute w-[76%] h-[76%] rounded-2xl overflow-hidden bg-background border border-border/80 shadow-xs rotate-6 translate-x-3 -translate-y-2 group-hover:rotate-10 group-hover:translate-x-4 transition-transform duration-300">
                          <img
                            src={snaps[1]?.img}
                            alt={snaps[1]?.title}
                            className="w-full h-full object-cover object-center"
                          />
                        </div>
                        {/* Front / 1st item */}
                        <div className="relative w-[76%] h-[76%] rounded-2xl overflow-hidden bg-background border border-border shadow-md -rotate-4 -translate-x-3 translate-y-2 group-hover:-rotate-6 group-hover:-translate-x-4 transition-transform duration-300 z-10">
                          <img
                            src={snaps[0]?.img}
                            alt={snaps[0]?.title}
                            className="w-full h-full object-cover object-center"
                          />
                        </div>
                      </div>
                    ) : (
                      // 3+ items: large fanned 3-card pile filling card generously
                      <div className="relative w-full h-full flex items-center justify-center">
                        {/* Back 3rd card */}
                        <div className="absolute w-[72%] h-[72%] rounded-2xl overflow-hidden bg-background border border-border/70 shadow-xs rotate-12 translate-x-4 -translate-y-2.5 opacity-90 group-hover:rotate-16 transition-transform duration-300">
                          <img
                            src={snaps[2]?.img}
                            alt={snaps[2]?.title}
                            className="w-full h-full object-cover object-center"
                          />
                        </div>
                        {/* Middle 2nd card */}
                        <div className="absolute w-[72%] h-[72%] rounded-2xl overflow-hidden bg-background border border-border/90 shadow-sm -rotate-8 -translate-x-3 translate-y-0.5 group-hover:-rotate-12 transition-transform duration-300 z-5">
                          <img
                            src={snaps[1]?.img}
                            alt={snaps[1]?.title}
                            className="w-full h-full object-cover object-center"
                          />
                        </div>
                        {/* Front 1st card */}
                        <div className="relative w-[72%] h-[72%] rounded-2xl overflow-hidden bg-background border border-border shadow-md rotate-2 translate-y-2 group-hover:rotate-0 transition-transform duration-300 z-10">
                          <img
                            src={snaps[0]?.img}
                            alt={snaps[0]?.title}
                            className="w-full h-full object-cover object-center"
                          />
                        </div>
                        {/* Count badge for overflow */}
                        {items.length > 3 && (
                          <div className="absolute bottom-2 end-2 z-20">
                            <span className="px-2 py-0.5 rounded-md text-[10px] font-bold bg-background/95 backdrop-blur-md border border-border text-foreground shadow-sm">
                              +{items.length - 3}
                            </span>
                          </div>
                        )}
                      </div>
                    )}
                  </Link>
                );
              })}
            </div>
          )}
        </div>
      )}

      {/* ── 3. Your Wishlist Sub-Section with Live Product Preview ── */}
      <div className="flex flex-col gap-3">
        <div className="flex items-center justify-between">
          <div className="flex items-baseline gap-2">
            <h2 className="text-sm sm:text-base font-bold text-foreground tracking-tight">
              {t("yourLists.title")}
            </h2>
            {wishlistCount > 0 && (
              <span className="text-xs text-muted-foreground">({wishlistCount})</span>
            )}
          </div>
          <Link
            href="/user/wishlist"
            className="inline-flex items-center gap-1 text-xs font-semibold text-accent-brand hover:underline"
          >
            <span>{t("yourLists.seeAll")}</span>
            <Arrow className="size-3.5" />
          </Link>
        </div>

        {wishlistPreview.length === 0 ? (
          <div className="bg-card border border-dashed border-border/80 rounded-2xl p-8 text-center flex flex-col items-center justify-center gap-2.5">
            <div className="size-10 rounded-full bg-secondary flex items-center justify-center text-muted-foreground">
              <Heart className="size-5" />
            </div>
            <p className="text-xs text-muted-foreground">{t("yourLists.empty")}</p>
            <Link
              href="/products"
              className="mt-1 inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg bg-secondary hover:bg-secondary/80 text-foreground border border-border text-xs font-semibold transition-colors"
            >
              <ExternalLink className="size-3.5" />
              <span>{t("yourLists.exploreProducts")}</span>
            </Link>
          </div>
        ) : (
          <div className="grid grid-cols-4 gap-2 sm:gap-3 max-w-xl">
            {wishlistPreview.map((item) => {
              const product = item.product;
              const imgSrc = normalizeMediaUrl(product?.thumbnailUrl);

              return (
                <Link
                  key={item.id}
                  href={`/products/${product?.slug || product?.id}`}
                  title={product?.title}
                  className="group block aspect-square w-full rounded-2xl bg-secondary/50 border border-border/80 hover:border-accent-brand/50 overflow-hidden shadow-2xs hover:shadow-xs transition-all"
                >
                  <img
                    src={imgSrc}
                    alt={product?.title || "Product"}
                    className="w-full h-full object-cover object-center group-hover:scale-110 transition-transform duration-300"
                  />
                </Link>
              );
            })}
          </div>
        )}
      </div>

      {/* ── 4. Account Management & Preferences Tiles ── */}
      <div className="flex flex-col gap-3">
        <h2 className="text-sm sm:text-base font-bold text-foreground tracking-tight">
          {t("accountShortcuts.title")}
        </h2>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
          {secondaryAccountTiles.map((card) => {
            const Icon = card.icon;

            return (
              <Link
                key={card.id}
                href={card.href}
                className="group bg-card border border-border/80 hover:border-accent-brand/50 rounded-xl p-3.5 sm:p-4 shadow-2xs hover:shadow-xs transition-all flex items-center justify-between gap-3 cursor-pointer"
              >
                <div className="flex items-center gap-3 min-w-0">
                  <div className="size-9 rounded-xl bg-secondary group-hover:bg-accent-brand/10 text-foreground group-hover:text-accent-brand flex items-center justify-center shrink-0 border border-border/60 transition-colors">
                    <Icon className="size-4.5" />
                  </div>
                  <div className="flex flex-col min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="text-xs sm:text-sm font-semibold text-foreground group-hover:text-accent-brand transition-colors truncate">
                        {card.title}
                      </span>
                      {card.badge && (
                        <span className="px-1.5 py-0.2 rounded-full text-[10px] font-semibold bg-secondary text-muted-foreground border border-border">
                          {card.badge}
                        </span>
                      )}
                    </div>
                    <span className="text-[11px] text-muted-foreground line-clamp-1 mt-0.5">
                      {card.desc}
                    </span>
                  </div>
                </div>

                <Chevron className="size-4 text-muted-foreground group-hover:text-accent-brand group-hover:translate-x-0.5 rtl:group-hover:-translate-x-0.5 transition-all shrink-0" />
              </Link>
            );
          })}
        </div>
      </div>
    </div>
  );
}
